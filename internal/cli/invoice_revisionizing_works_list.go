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

type invoiceRevisionizingWorksListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	Broker           string
	CreatedBy        string
	OrganizationType string
	OrganizationID   string
	JID              string
}

type invoiceRevisionizingWorkRow struct {
	ID               string `json:"id"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	JID              string `json:"jid,omitempty"`
	IsRetry          bool   `json:"is_retry"`
	ScheduledAt      string `json:"scheduled_at,omitempty"`
	ProcessedAt      string `json:"processed_at,omitempty"`
	Comment          string `json:"comment,omitempty"`
}

func newInvoiceRevisionizingWorksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoice revisionizing work",
		Long: `List invoice revisionizing work items with pagination.

Output Columns:
  ID         Work identifier
  ORG TYPE   Organization type
  ORG ID     Organization ID
  BROKER     Broker ID
  CREATED BY Creator user ID
  JID        Background job ID (if scheduled async)
  RETRY      Whether the work is a retry
  SCHEDULED  When work was scheduled
  PROCESSED  When work finished processing
  COMMENT    Comment (truncated)

Filters:
  --broker             Filter by broker ID
  --created-by         Filter by creator user ID
  --organization-type  Filter by organization type (Broker, Trucker, Customer)
  --organization-id    Filter by organization ID
  --jid                Filter by background job ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List invoice revisionizing work
  xbe view invoice-revisionizing-works list

  # Filter by broker
  xbe view invoice-revisionizing-works list --broker 123

  # Filter by organization
  xbe view invoice-revisionizing-works list --organization-type Broker --organization-id 456

  # Filter by creator
  xbe view invoice-revisionizing-works list --created-by 789

  # Filter by job id
  xbe view invoice-revisionizing-works list --jid 12345

  # Output as JSON
  xbe view invoice-revisionizing-works list --json`,
		RunE: runInvoiceRevisionizingWorksList,
	}
	initInvoiceRevisionizingWorksListFlags(cmd)
	return cmd
}

func init() {
	invoiceRevisionizingWorksCmd.AddCommand(newInvoiceRevisionizingWorksListCmd())
}

func initInvoiceRevisionizingWorksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("organization-type", "", "Filter by organization type (Broker, Trucker, Customer)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID")
	cmd.Flags().String("jid", "", "Filter by background job ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceRevisionizingWorksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInvoiceRevisionizingWorksListOptions(cmd)
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
	query.Set("fields[invoice-revisionizing-works]", "comment,is-retry,jid,scheduled-at,processed-at,organization,broker,created-by")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	if opts.OrganizationType != "" && opts.OrganizationID != "" {
		query.Set("filter[organization]", opts.OrganizationType+"|"+opts.OrganizationID)
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[jid]", opts.JID)

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-revisionizing-works", query)
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

	rows := buildInvoiceRevisionizingWorkRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInvoiceRevisionizingWorksTable(cmd, rows)
}

func parseInvoiceRevisionizingWorksListOptions(cmd *cobra.Command) (invoiceRevisionizingWorksListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	jid, _ := cmd.Flags().GetString("jid")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceRevisionizingWorksListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		Broker:           broker,
		CreatedBy:        createdBy,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
		JID:              jid,
	}, nil
}

func buildInvoiceRevisionizingWorkRows(resp jsonAPIResponse) []invoiceRevisionizingWorkRow {
	rows := make([]invoiceRevisionizingWorkRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := invoiceRevisionizingWorkRow{
			ID:          resource.ID,
			JID:         stringAttr(attrs, "jid"),
			IsRetry:     boolAttr(attrs, "is-retry"),
			ScheduledAt: formatDateTime(stringAttr(attrs, "scheduled-at")),
			ProcessedAt: formatDateTime(stringAttr(attrs, "processed-at")),
			Comment:     strings.TrimSpace(stringAttr(attrs, "comment")),
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderInvoiceRevisionizingWorksTable(cmd *cobra.Command, rows []invoiceRevisionizingWorkRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No invoice revisionizing works found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORG TYPE\tORG ID\tBROKER\tCREATED BY\tJID\tRETRY\tSCHEDULED\tPROCESSED\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.OrganizationType,
			row.OrganizationID,
			row.BrokerID,
			row.CreatedByID,
			row.JID,
			formatBool(row.IsRetry),
			row.ScheduledAt,
			row.ProcessedAt,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
