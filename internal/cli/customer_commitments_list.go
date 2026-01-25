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

type customerCommitmentsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	Status                      string
	Customer                    string
	CustomerID                  string
	Broker                      string
	BrokerID                    string
	ExternalIdentificationValue string
	ExternalJobNumber           string
	CreatedAtMin                string
	CreatedAtMax                string
	UpdatedAtMin                string
	UpdatedAtMax                string
	NotID                       string
}

type customerCommitmentRow struct {
	ID                    string `json:"id"`
	Status                string `json:"status,omitempty"`
	Label                 string `json:"label,omitempty"`
	CustomerID            string `json:"customer_id,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	Tons                  string `json:"tons,omitempty"`
	TonsPerShift          string `json:"tons_per_shift,omitempty"`
	ExternalJobNumber     string `json:"external_job_number,omitempty"`
	PrecedingCommitmentID string `json:"preceding_commitment_id,omitempty"`
}

func newCustomerCommitmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer commitments",
		Long: `List customer commitments with filtering and pagination.

Output Columns:
  ID               Commitment identifier
  STATUS           Commitment status
  CUSTOMER         Customer ID
  BROKER           Broker ID
  TONS             Committed tons
  TONS/SHIFT       Committed tons per shift
  JOB NUMBER       External job number
  LABEL            Commitment label

Filters:
  --status                        Filter by status (editing, active, inactive)
  --customer                      Filter by customer ID
  --customer-id                   Filter by customer ID (buyer_id)
  --broker                        Filter by broker ID
  --broker-id                     Filter by broker ID (seller_id)
  --external-job-number           Filter by external job number
  --external-identification-value Filter by external identification value
  --created-at-min                Filter by created-at on/after (ISO 8601)
  --created-at-max                Filter by created-at on/before (ISO 8601)
  --updated-at-min                Filter by updated-at on/after (ISO 8601)
  --updated-at-max                Filter by updated-at on/before (ISO 8601)
  --not-id                        Exclude by commitment ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List customer commitments
  xbe view customer-commitments list

  # Filter by customer and status
  xbe view customer-commitments list --customer 123 --status active

  # Filter by external job number
  xbe view customer-commitments list --external-job-number JOB-42

  # Output as JSON
  xbe view customer-commitments list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerCommitmentsList,
	}
	initCustomerCommitmentsListFlags(cmd)
	return cmd
}

func init() {
	customerCommitmentsCmd.AddCommand(newCustomerCommitmentsListCmd())
}

func initCustomerCommitmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (editing, active, inactive)")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("customer-id", "", "Filter by customer ID (buyer_id)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (seller_id)")
	cmd.Flags().String("external-job-number", "", "Filter by external job number")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("not-id", "", "Exclude by commitment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerCommitmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerCommitmentsListOptions(cmd)
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
	query.Set("fields[customer-commitments]", "status,label,tons,tons-per-shift,external-job-number,customer,broker,buyer,seller")

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
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[customer-id]", opts.CustomerID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[external-job-number]", opts.ExternalJobNumber)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[not-id]", opts.NotID)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-commitments", query)
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

	rows := buildCustomerCommitmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerCommitmentsTable(cmd, rows)
}

func parseCustomerCommitmentsListOptions(cmd *cobra.Command) (customerCommitmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	customer, _ := cmd.Flags().GetString("customer")
	customerID, _ := cmd.Flags().GetString("customer-id")
	broker, _ := cmd.Flags().GetString("broker")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	externalJobNumber, _ := cmd.Flags().GetString("external-job-number")
	externalIdentificationValue, _ := cmd.Flags().GetString("external-identification-value")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	notID, _ := cmd.Flags().GetString("not-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerCommitmentsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		Status:                      status,
		Customer:                    customer,
		CustomerID:                  customerID,
		Broker:                      broker,
		BrokerID:                    brokerID,
		ExternalJobNumber:           externalJobNumber,
		ExternalIdentificationValue: externalIdentificationValue,
		CreatedAtMin:                createdAtMin,
		CreatedAtMax:                createdAtMax,
		UpdatedAtMin:                updatedAtMin,
		UpdatedAtMax:                updatedAtMax,
		NotID:                       notID,
	}, nil
}

func buildCustomerCommitmentRows(resp jsonAPIResponse) []customerCommitmentRow {
	rows := make([]customerCommitmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCustomerCommitmentRow(resource))
	}
	return rows
}

func buildCustomerCommitmentRow(resource jsonAPIResource) customerCommitmentRow {
	attrs := resource.Attributes
	row := customerCommitmentRow{
		ID:                resource.ID,
		Status:            stringAttr(attrs, "status"),
		Label:             strings.TrimSpace(stringAttr(attrs, "label")),
		Tons:              stringAttr(attrs, "tons"),
		TonsPerShift:      stringAttr(attrs, "tons-per-shift"),
		ExternalJobNumber: strings.TrimSpace(stringAttr(attrs, "external-job-number")),
	}

	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	} else if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	} else if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["preceding-commitment"]; ok && rel.Data != nil {
		row.PrecedingCommitmentID = rel.Data.ID
	}

	return row
}

func buildCustomerCommitmentRowFromSingle(resp jsonAPISingleResponse) customerCommitmentRow {
	return buildCustomerCommitmentRow(resp.Data)
}

func renderCustomerCommitmentsTable(cmd *cobra.Command, rows []customerCommitmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer commitments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tCUSTOMER\tBROKER\tTONS\tTONS/SHIFT\tJOB NUMBER\tLABEL")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.CustomerID,
			row.BrokerID,
			row.Tons,
			row.TonsPerShift,
			row.ExternalJobNumber,
			row.Label,
		)
	}
	return writer.Flush()
}
