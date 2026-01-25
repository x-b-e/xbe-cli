package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type commitmentsListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Sort     string
	Status   string
	Broker   string
	BrokerID string
}

type commitmentRow struct {
	ID           string `json:"id"`
	Status       string `json:"status,omitempty"`
	Label        string `json:"label,omitempty"`
	BuyerType    string `json:"buyer_type,omitempty"`
	BuyerID      string `json:"buyer_id,omitempty"`
	SellerType   string `json:"seller_type,omitempty"`
	SellerID     string `json:"seller_id,omitempty"`
	TruckScopeID string `json:"truck_scope_id,omitempty"`
}

func newCommitmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commitments",
		Long: `List commitments.

Output Columns:
  ID           Commitment identifier
  STATUS       Commitment status
  LABEL        Commitment label
  BUYER        Buyer (type/id)
  SELLER       Seller (type/id)
  TRUCK_SCOPE  Truck scope ID

Filters:
  --status     Filter by status (editing, active, inactive)
  --broker     Filter by broker ID (matches buyer or seller broker)
  --broker-id  Alias for --broker

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List commitments
  xbe view commitments list

  # Filter by status
  xbe view commitments list --status active

  # Filter by broker
  xbe view commitments list --broker 123

  # Output as JSON
  xbe view commitments list --json`,
		Args: cobra.NoArgs,
		RunE: runCommitmentsList,
	}
	initCommitmentsListFlags(cmd)
	return cmd
}

func init() {
	commitmentsCmd.AddCommand(newCommitmentsListCmd())
}

func initCommitmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("broker-id", "", "Alias for --broker")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommitmentsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[commitments]", "status,label,notes,buyer,seller,truck-scope")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)

	brokerFilter := opts.Broker
	if brokerFilter == "" {
		brokerFilter = opts.BrokerID
	}
	setFilterIfPresent(query, "filter[broker]", brokerFilter)

	body, _, err := client.Get(cmd.Context(), "/v1/commitments", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildCommitmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommitmentsTable(cmd, rows)
}

func parseCommitmentsListOptions(cmd *cobra.Command) (commitmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	broker, _ := cmd.Flags().GetString("broker")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentsListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Sort:     sort,
		Status:   status,
		Broker:   broker,
		BrokerID: brokerID,
	}, nil
}

func buildCommitmentRows(resp jsonAPIResponse) []commitmentRow {
	rows := make([]commitmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCommitmentRow(resource))
	}
	return rows
}

func buildCommitmentRow(resource jsonAPIResource) commitmentRow {
	row := commitmentRow{
		ID:     resource.ID,
		Status: stringAttr(resource.Attributes, "status"),
		Label:  strings.TrimSpace(stringAttr(resource.Attributes, "label")),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		row.BuyerType = rel.Data.Type
		row.BuyerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		row.SellerType = rel.Data.Type
		row.SellerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["truck-scope"]; ok && rel.Data != nil {
		row.TruckScopeID = rel.Data.ID
	}

	return row
}

func buildCommitmentRowFromSingle(resp jsonAPISingleResponse) commitmentRow {
	return buildCommitmentRow(resp.Data)
}

func renderCommitmentsTable(cmd *cobra.Command, rows []commitmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No commitments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tLABEL\tBUYER\tSELLER\tTRUCK_SCOPE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Label, 26),
			truncateString(formatTypeID(row.BuyerType, row.BuyerID), 32),
			truncateString(formatTypeID(row.SellerType, row.SellerID), 32),
			row.TruckScopeID,
		)
	}
	return writer.Flush()
}

func formatTypeID(resourceType, resourceID string) string {
	if resourceType == "" {
		return resourceID
	}
	if resourceID == "" {
		return resourceType
	}
	return resourceType + "/" + resourceID
}
