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

type jobProductionPlanChangeSetsListOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	NoAuth                         bool
	Limit                          int
	Offset                         int
	Sort                           string
	Broker                         string
	Customer                       string
	CreatedBy                      string
	ChangeOldMaterialType          string
	ChangeNewMaterialType          string
	ChangeNewPlanner               string
	ChangeNewPlannerNullify        string
	ChangeNewProjectManager        string
	ChangeNewProjectManagerNullify string
	ShouldPersist                  string
}

type jobProductionPlanChangeSetRow struct {
	ID            string `json:"id"`
	BrokerID      string `json:"broker_id,omitempty"`
	BrokerName    string `json:"broker_name,omitempty"`
	CustomerID    string `json:"customer_id,omitempty"`
	CustomerName  string `json:"customer_name,omitempty"`
	CreatedByID   string `json:"created_by_id,omitempty"`
	CreatedBy     string `json:"created_by_name,omitempty"`
	ShouldPersist bool   `json:"should_persist"`
	ProcessedAt   string `json:"processed_at,omitempty"`
	MatchCount    int    `json:"match_count,omitempty"`
}

func newJobProductionPlanChangeSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan change sets",
		Long: `List job production plan change sets with filtering and pagination.

Output Columns:
  ID           Change set identifier
  BROKER       Broker name
  CUSTOMER     Customer name
  CREATED BY   User who created the change set
  PERSIST      Whether changes should persist
  PROCESSED AT When results were processed
  MATCHES      Number of matched job production plans

Filters:
  --broker                          Filter by broker ID
  --customer                        Filter by customer ID
  --created-by                      Filter by creator user ID
  --change-old-material-type        Filter by old material type ID
  --change-new-material-type        Filter by new material type ID
  --change-new-planner              Filter by new planner user ID
  --change-new-planner-nullify      Filter by new planner nullify (true/false)
  --change-new-project-manager      Filter by new project manager user ID
  --change-new-project-manager-nullify Filter by new project manager nullify (true/false)
  --should-persist                  Filter by should persist (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List change sets
  xbe view job-production-plan-change-sets list

  # Filter by customer
  xbe view job-production-plan-change-sets list --customer 123

  # Filter by planner
  xbe view job-production-plan-change-sets list --change-new-planner 456

  # Output as JSON
  xbe view job-production-plan-change-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanChangeSetsList,
	}
	initJobProductionPlanChangeSetsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanChangeSetsCmd.AddCommand(newJobProductionPlanChangeSetsListCmd())
}

func initJobProductionPlanChangeSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("change-old-material-type", "", "Filter by old material type ID")
	cmd.Flags().String("change-new-material-type", "", "Filter by new material type ID")
	cmd.Flags().String("change-new-planner", "", "Filter by new planner user ID")
	cmd.Flags().String("change-new-planner-nullify", "", "Filter by new planner nullify (true/false)")
	cmd.Flags().String("change-new-project-manager", "", "Filter by new project manager user ID")
	cmd.Flags().String("change-new-project-manager-nullify", "", "Filter by new project manager nullify (true/false)")
	cmd.Flags().String("should-persist", "", "Filter by should persist (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanChangeSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanChangeSetsListOptions(cmd)
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
	query.Set("fields[job-production-plan-change-sets]", "broker,customer,created-by,should-persist,processed-at,match-ids")
	query.Set("include", "broker,customer,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
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
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[change-old-material-type]", opts.ChangeOldMaterialType)
	setFilterIfPresent(query, "filter[change-new-material-type]", opts.ChangeNewMaterialType)
	setFilterIfPresent(query, "filter[change-new-planner]", opts.ChangeNewPlanner)
	setFilterIfPresent(query, "filter[change-new-planner-nullify]", opts.ChangeNewPlannerNullify)
	setFilterIfPresent(query, "filter[change-new-project-manager]", opts.ChangeNewProjectManager)
	setFilterIfPresent(query, "filter[change-new-project-manager-nullify]", opts.ChangeNewProjectManagerNullify)
	setFilterIfPresent(query, "filter[should-persist]", opts.ShouldPersist)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-change-sets", query)
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

	rows := buildJobProductionPlanChangeSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanChangeSetsTable(cmd, rows)
}

func parseJobProductionPlanChangeSetsListOptions(cmd *cobra.Command) (jobProductionPlanChangeSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	createdBy, _ := cmd.Flags().GetString("created-by")
	changeOldMaterialType, _ := cmd.Flags().GetString("change-old-material-type")
	changeNewMaterialType, _ := cmd.Flags().GetString("change-new-material-type")
	changeNewPlanner, _ := cmd.Flags().GetString("change-new-planner")
	changeNewPlannerNullify, _ := cmd.Flags().GetString("change-new-planner-nullify")
	changeNewProjectManager, _ := cmd.Flags().GetString("change-new-project-manager")
	changeNewProjectManagerNullify, _ := cmd.Flags().GetString("change-new-project-manager-nullify")
	shouldPersist, _ := cmd.Flags().GetString("should-persist")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanChangeSetsListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Limit:                          limit,
		Offset:                         offset,
		Sort:                           sort,
		Broker:                         broker,
		Customer:                       customer,
		CreatedBy:                      createdBy,
		ChangeOldMaterialType:          changeOldMaterialType,
		ChangeNewMaterialType:          changeNewMaterialType,
		ChangeNewPlanner:               changeNewPlanner,
		ChangeNewPlannerNullify:        changeNewPlannerNullify,
		ChangeNewProjectManager:        changeNewProjectManager,
		ChangeNewProjectManagerNullify: changeNewProjectManagerNullify,
		ShouldPersist:                  shouldPersist,
	}, nil
}

func buildJobProductionPlanChangeSetRows(resp jsonAPIResponse) []jobProductionPlanChangeSetRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanChangeSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanChangeSetRow(resource, included))
	}
	return rows
}

func jobProductionPlanChangeSetRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanChangeSetRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildJobProductionPlanChangeSetRow(resp.Data, included)
}

func buildJobProductionPlanChangeSetRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanChangeSetRow {
	attrs := resource.Attributes
	matchIDs := stringSliceAttr(attrs, "match-ids")

	row := jobProductionPlanChangeSetRow{
		ID:            resource.ID,
		ShouldPersist: boolAttr(attrs, "should-persist"),
		ProcessedAt:   formatDateTime(stringAttr(attrs, "processed-at")),
		MatchCount:    len(matchIDs),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CustomerName = stringAttr(customer.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CreatedBy = stringAttr(user.Attributes, "name")
		}
	}

	return row
}

func renderJobProductionPlanChangeSetsTable(cmd *cobra.Command, rows []jobProductionPlanChangeSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan change sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBROKER\tCUSTOMER\tCREATED BY\tPERSIST\tPROCESSED AT\tMATCHES")

	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" {
			broker = row.BrokerID
		}
		customer := row.CustomerName
		if customer == "" {
			customer = row.CustomerID
		}
		createdBy := row.CreatedBy
		if createdBy == "" {
			createdBy = row.CreatedByID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%t\t%s\t%d\n",
			row.ID,
			broker,
			customer,
			createdBy,
			row.ShouldPersist,
			row.ProcessedAt,
			row.MatchCount,
		)
	}

	return writer.Flush()
}
