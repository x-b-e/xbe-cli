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

type costCodeTruckingCostSummariesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Broker       string
	CreatedBy    string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type costCodeTruckingCostSummaryRow struct {
	ID            string `json:"id"`
	BrokerID      string `json:"broker_id,omitempty"`
	BrokerName    string `json:"broker_name,omitempty"`
	CreatedByID   string `json:"created_by_id,omitempty"`
	CreatedByName string `json:"created_by_name,omitempty"`
	StartOn       string `json:"start_on,omitempty"`
	EndOn         string `json:"end_on,omitempty"`
}

func newCostCodeTruckingCostSummariesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cost code trucking cost summaries",
		Long: `List cost code trucking cost summaries with filtering and pagination.

Output Columns:
  ID         Summary identifier
  BROKER     Broker name
  START ON   Start date for the summary window
  END ON     End date for the summary window
  CREATED BY Summary creator

Filters:
  --broker         Filter by broker ID
  --created-by     Filter by creator user ID
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --is-created-at  Filter by has created-at (true/false)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)
  --is-updated-at  Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List summaries
  xbe view cost-code-trucking-cost-summaries list

  # Filter by broker
  xbe view cost-code-trucking-cost-summaries list --broker 123

  # Filter by creator
  xbe view cost-code-trucking-cost-summaries list --created-by 456

  # Output as JSON
  xbe view cost-code-trucking-cost-summaries list --json`,
		Args: cobra.NoArgs,
		RunE: runCostCodeTruckingCostSummariesList,
	}
	initCostCodeTruckingCostSummariesListFlags(cmd)
	return cmd
}

func init() {
	costCodeTruckingCostSummariesCmd.AddCommand(newCostCodeTruckingCostSummariesListCmd())
}

func initCostCodeTruckingCostSummariesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCostCodeTruckingCostSummariesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCostCodeTruckingCostSummariesListOptions(cmd)
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
	query.Set("fields[cost-code-trucking-cost-summaries]", "start-on,end-on,broker,created-by")
	query.Set("include", "broker,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/cost-code-trucking-cost-summaries", query)
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

	rows := buildCostCodeTruckingCostSummaryRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCostCodeTruckingCostSummariesTable(cmd, rows)
}

func parseCostCodeTruckingCostSummariesListOptions(cmd *cobra.Command) (costCodeTruckingCostSummariesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return costCodeTruckingCostSummariesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Broker:       broker,
		CreatedBy:    createdBy,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildCostCodeTruckingCostSummaryRows(resp jsonAPIResponse) []costCodeTruckingCostSummaryRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]costCodeTruckingCostSummaryRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildCostCodeTruckingCostSummaryRow(resource)

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(inc.Attributes, "company-name")
			}
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CreatedByName = stringAttr(inc.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func buildCostCodeTruckingCostSummaryRow(resource jsonAPIResource) costCodeTruckingCostSummaryRow {
	row := costCodeTruckingCostSummaryRow{
		ID:      resource.ID,
		StartOn: stringAttr(resource.Attributes, "start-on"),
		EndOn:   stringAttr(resource.Attributes, "end-on"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderCostCodeTruckingCostSummariesTable(cmd *cobra.Command, rows []costCodeTruckingCostSummaryRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No cost code trucking cost summaries found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBROKER\tSTART ON\tEND ON\tCREATED BY")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" {
			broker = row.BrokerID
		}
		createdBy := row.CreatedByName
		if createdBy == "" {
			createdBy = row.CreatedByID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			broker,
			row.StartOn,
			row.EndOn,
			createdBy,
		)
	}

	return writer.Flush()
}
