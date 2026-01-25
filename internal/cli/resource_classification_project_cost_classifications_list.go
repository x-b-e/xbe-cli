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

type resourceClassificationProjectCostClassificationsListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Sort                       string
	ResourceClassificationType string
	ResourceClassificationID   string
	ProjectCostClassification  string
	Broker                     string
	CreatedAtMin               string
	CreatedAtMax               string
	UpdatedAtMin               string
	UpdatedAtMax               string
	IsCreatedAt                string
	IsUpdatedAt                string
}

type resourceClassificationProjectCostClassificationRow struct {
	ID                         string `json:"id"`
	ResourceClassificationType string `json:"resource_classification_type,omitempty"`
	ResourceClassificationID   string `json:"resource_classification_id,omitempty"`
	ProjectCostClassification  string `json:"project_cost_classification_id,omitempty"`
	Broker                     string `json:"broker_id,omitempty"`
}

func newResourceClassificationProjectCostClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resource classification project cost classifications",
		Long: `List resource classification project cost classifications with filtering and pagination.

Resource classification project cost classifications link labor or equipment
classifications to project cost classifications for a broker.

Output Columns:
  ID               Association identifier
  RESOURCE TYPE    Resource classification type
  RESOURCE ID      Resource classification ID
  PROJECT COST     Project cost classification ID
  BROKER           Broker ID

Filters:
  --resource-classification-type      Filter by resource classification type (LaborClassification, EquipmentClassification)
  --resource-classification-id        Filter by resource classification ID (requires --resource-classification-type)
  --project-cost-classification       Filter by project cost classification ID
  --broker                            Filter by broker ID
  --created-at-min                    Filter by created-at on/after (ISO 8601)
  --created-at-max                    Filter by created-at on/before (ISO 8601)
  --updated-at-min                    Filter by updated-at on/after (ISO 8601)
  --updated-at-max                    Filter by updated-at on/before (ISO 8601)
  --is-created-at                     Filter by presence of created-at (true/false)
  --is-updated-at                     Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List resource classification project cost classifications
  xbe view resource-classification-project-cost-classifications list

  # Filter by resource classification type
  xbe view resource-classification-project-cost-classifications list --resource-classification-type LaborClassification

  # Filter by resource classification and project cost classification
  xbe view resource-classification-project-cost-classifications list \
    --resource-classification-type LaborClassification \
    --resource-classification-id 456 \
    --project-cost-classification 789

  # Filter by broker
  xbe view resource-classification-project-cost-classifications list --broker 123

  # Output as JSON
  xbe view resource-classification-project-cost-classifications list --json`,
		Args: cobra.NoArgs,
		RunE: runResourceClassificationProjectCostClassificationsList,
	}
	initResourceClassificationProjectCostClassificationsListFlags(cmd)
	return cmd
}

func init() {
	resourceClassificationProjectCostClassificationsCmd.AddCommand(newResourceClassificationProjectCostClassificationsListCmd())
}

func initResourceClassificationProjectCostClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("resource-classification-type", "", "Filter by resource classification type (LaborClassification, EquipmentClassification)")
	cmd.Flags().String("resource-classification-id", "", "Filter by resource classification ID (requires --resource-classification-type)")
	cmd.Flags().String("project-cost-classification", "", "Filter by project cost classification ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runResourceClassificationProjectCostClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseResourceClassificationProjectCostClassificationsListOptions(cmd)
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
	query.Set("fields[resource-classification-project-cost-classifications]", "resource-classification,project-cost-classification,broker,created-at,updated-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.ResourceClassificationType != "" && opts.ResourceClassificationID != "" {
		query.Set("filter[resource_classification]", opts.ResourceClassificationType+"|"+opts.ResourceClassificationID)
	} else if opts.ResourceClassificationType != "" {
		query.Set("filter[resource_classification_type]", opts.ResourceClassificationType)
	}

	setFilterIfPresent(query, "filter[project_cost_classification]", opts.ProjectCostClassification)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/resource-classification-project-cost-classifications", query)
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

	rows := buildResourceClassificationProjectCostClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderResourceClassificationProjectCostClassificationsTable(cmd, rows)
}

func parseResourceClassificationProjectCostClassificationsListOptions(cmd *cobra.Command) (resourceClassificationProjectCostClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	resourceClassificationType, _ := cmd.Flags().GetString("resource-classification-type")
	resourceClassificationID, _ := cmd.Flags().GetString("resource-classification-id")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	broker, _ := cmd.Flags().GetString("broker")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return resourceClassificationProjectCostClassificationsListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		Sort:                       sort,
		ResourceClassificationType: resourceClassificationType,
		ResourceClassificationID:   resourceClassificationID,
		ProjectCostClassification:  projectCostClassification,
		Broker:                     broker,
		CreatedAtMin:               createdAtMin,
		CreatedAtMax:               createdAtMax,
		UpdatedAtMin:               updatedAtMin,
		UpdatedAtMax:               updatedAtMax,
		IsCreatedAt:                isCreatedAt,
		IsUpdatedAt:                isUpdatedAt,
	}, nil
}

func buildResourceClassificationProjectCostClassificationRows(resp jsonAPIResponse) []resourceClassificationProjectCostClassificationRow {
	rows := make([]resourceClassificationProjectCostClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := resourceClassificationProjectCostClassificationRow{
			ID:                        resource.ID,
			ProjectCostClassification: relationshipIDFromMap(resource.Relationships, "project-cost-classification"),
			Broker:                    relationshipIDFromMap(resource.Relationships, "broker"),
		}

		if rel, ok := resource.Relationships["resource-classification"]; ok && rel.Data != nil {
			row.ResourceClassificationType = rel.Data.Type
			row.ResourceClassificationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderResourceClassificationProjectCostClassificationsTable(cmd *cobra.Command, rows []resourceClassificationProjectCostClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No resource classification project cost classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRESOURCE TYPE\tRESOURCE ID\tPROJECT COST\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ResourceClassificationType,
			row.ResourceClassificationID,
			row.ProjectCostClassification,
			row.Broker,
		)
	}
	return writer.Flush()
}
