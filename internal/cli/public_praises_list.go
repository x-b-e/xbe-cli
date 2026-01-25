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

type publicPraisesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	OrganizationType string
	OrganizationID   string
	Broker           string
	GivenBy          string
	Recipient        string
	ReceivedBy       string
	CultureValueIDs  string
}

type publicPraiseRow struct {
	ID               string `json:"id"`
	Description      string `json:"description,omitempty"`
	GivenByID        string `json:"given_by_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
}

func newPublicPraisesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List public praises",
		Long: `List public praises (employee recognition).

Output Columns:
  ID               Public praise identifier
  DESCRIPTION      Description of the praise
  GIVEN BY         User ID who gave the praise
  ORG TYPE         Organization type
  ORG ID           Organization ID
  BROKER ID        Associated broker ID

Filters:
  --organization-type    Filter by organization type
  --organization-id      Filter by organization ID
  --broker               Filter by broker ID
  --given-by             Filter by user ID who gave the praise
  --recipient            Filter by recipient user ID
  --received-by          Filter by received-by user ID
  --culture-value-ids    Filter by culture value IDs (comma-separated)`,
		Example: `  # List all public praises
  xbe view public-praises list

  # Filter by broker
  xbe view public-praises list --broker 123

  # Filter by who gave the praise
  xbe view public-praises list --given-by 456

  # Output as JSON
  xbe view public-praises list --json`,
		RunE: runPublicPraisesList,
	}
	initPublicPraisesListFlags(cmd)
	return cmd
}

func init() {
	publicPraisesCmd.AddCommand(newPublicPraisesListCmd())
}

func initPublicPraisesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("organization-type", "", "Filter by organization type")
	cmd.Flags().String("organization-id", "", "Filter by organization ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("given-by", "", "Filter by user ID who gave the praise")
	cmd.Flags().String("recipient", "", "Filter by recipient user ID")
	cmd.Flags().String("received-by", "", "Filter by received-by user ID")
	cmd.Flags().String("culture-value-ids", "", "Filter by culture value IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPublicPraisesList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePublicPraisesListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	if opts.OrganizationType != "" && opts.OrganizationID != "" {
		query.Set("filter[by_organization]", opts.OrganizationType+"|"+opts.OrganizationID)
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[given_by]", opts.GivenBy)
	setFilterIfPresent(query, "filter[recipient]", opts.Recipient)
	setFilterIfPresent(query, "filter[received_by]", opts.ReceivedBy)
	setFilterIfPresent(query, "filter[culture_value_ids]", opts.CultureValueIDs)

	body, _, err := client.Get(cmd.Context(), "/v1/public-praises", query)
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

	rows := buildPublicPraiseRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPublicPraisesTable(cmd, rows)
}

func parsePublicPraisesListOptions(cmd *cobra.Command) (publicPraisesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	broker, _ := cmd.Flags().GetString("broker")
	givenBy, _ := cmd.Flags().GetString("given-by")
	recipient, _ := cmd.Flags().GetString("recipient")
	receivedBy, _ := cmd.Flags().GetString("received-by")
	cultureValueIDs, _ := cmd.Flags().GetString("culture-value-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return publicPraisesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
		Broker:           broker,
		GivenBy:          givenBy,
		Recipient:        recipient,
		ReceivedBy:       receivedBy,
		CultureValueIDs:  cultureValueIDs,
	}, nil
}

func buildPublicPraiseRows(resp jsonAPIResponse) []publicPraiseRow {
	rows := make([]publicPraiseRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := publicPraiseRow{
			ID:          resource.ID,
			Description: stringAttr(resource.Attributes, "description"),
		}

		if rel, ok := resource.Relationships["given-by"]; ok && rel.Data != nil {
			row.GivenByID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderPublicPraisesTable(cmd *cobra.Command, rows []publicPraiseRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No public praises found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDESCRIPTION\tGIVEN BY\tORG TYPE\tORG ID\tBROKER ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Description, 40),
			row.GivenByID,
			row.OrganizationType,
			row.OrganizationID,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
