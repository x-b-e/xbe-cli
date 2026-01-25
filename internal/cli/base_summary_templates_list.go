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

type baseSummaryTemplatesListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	Label     string
	Broker    string
	CreatedBy string
}

type baseSummaryTemplateRow struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	BrokerID    string `json:"broker_id,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

func newBaseSummaryTemplatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List base summary templates",
		Long: `List base summary templates with filtering and pagination.

Output Columns:
  ID          Template identifier
  LABEL       Template label
  BROKER      Broker ID (if scoped)
  CREATED BY  Creator user ID
  START DATE  Optional start date
  END DATE    Optional end date

Filters:
  --label       Filter by label
  --broker      Filter by broker ID
  --created-by  Filter by creator user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List base summary templates
  xbe view base-summary-templates list

  # Filter by broker
  xbe view base-summary-templates list --broker 123

  # Filter by label
  xbe view base-summary-templates list --label "Daily Summary"

  # JSON output
  xbe view base-summary-templates list --json`,
		Args: cobra.NoArgs,
		RunE: runBaseSummaryTemplatesList,
	}
	initBaseSummaryTemplatesListFlags(cmd)
	return cmd
}

func init() {
	baseSummaryTemplatesCmd.AddCommand(newBaseSummaryTemplatesListCmd())
}

func initBaseSummaryTemplatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("label", "", "Filter by label")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBaseSummaryTemplatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBaseSummaryTemplatesListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[base-summary-templates]", "label,broker,created-by,start-date,end-date")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[label]", opts.Label)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/base-summary-templates", query)
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

	rows := buildBaseSummaryTemplateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBaseSummaryTemplatesTable(cmd, rows)
}

func parseBaseSummaryTemplatesListOptions(cmd *cobra.Command) (baseSummaryTemplatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	label, _ := cmd.Flags().GetString("label")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return baseSummaryTemplatesListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		Label:     label,
		Broker:    broker,
		CreatedBy: createdBy,
	}, nil
}

func buildBaseSummaryTemplateRows(resp jsonAPIResponse) []baseSummaryTemplateRow {
	rows := make([]baseSummaryTemplateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildBaseSummaryTemplateRow(resource))
	}
	return rows
}

func buildBaseSummaryTemplateRowFromSingle(resp jsonAPISingleResponse) baseSummaryTemplateRow {
	return buildBaseSummaryTemplateRow(resp.Data)
}

func buildBaseSummaryTemplateRow(resource jsonAPIResource) baseSummaryTemplateRow {
	attrs := resource.Attributes
	row := baseSummaryTemplateRow{
		ID:        resource.ID,
		Label:     strings.TrimSpace(stringAttr(attrs, "label")),
		StartDate: formatDateTime(stringAttr(attrs, "start-date")),
		EndDate:   formatDateTime(stringAttr(attrs, "end-date")),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderBaseSummaryTemplatesTable(cmd *cobra.Command, rows []baseSummaryTemplateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No base summary templates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tLABEL\tBROKER\tCREATED BY\tSTART DATE\tEND DATE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", row.ID, row.Label, row.BrokerID, row.CreatedByID, row.StartDate, row.EndDate)
	}
	return writer.Flush()
}
