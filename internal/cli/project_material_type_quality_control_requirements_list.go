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

type projectMaterialTypeQualityControlRequirementsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Sort                         string
	ProjectMaterialType          string
	QualityControlClassification string
}

type projectMaterialTypeQualityControlRequirementRow struct {
	ID                               string `json:"id"`
	ProjectMaterialTypeID            string `json:"project_material_type_id,omitempty"`
	ProjectMaterialTypeName          string `json:"project_material_type_name,omitempty"`
	QualityControlClassificationID   string `json:"quality_control_classification_id,omitempty"`
	QualityControlClassificationName string `json:"quality_control_classification_name,omitempty"`
	Note                             string `json:"note,omitempty"`
}

func newProjectMaterialTypeQualityControlRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project material type quality control requirements",
		Long: `List project material type quality control requirements with filtering and pagination.

Project material type quality control requirements specify which quality control
classifications are required for a project material type.

Output Columns:
  ID                             Requirement identifier
  PROJECT MATERIAL TYPE          Project material type name or ID
  QUALITY CONTROL CLASSIFICATION Quality control classification name or ID
  NOTE                           Optional note

Filters:
  --project-material-type          Filter by project material type ID
  --quality-control-classification Filter by quality control classification ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List requirements
  xbe view project-material-type-quality-control-requirements list

  # Filter by project material type
  xbe view project-material-type-quality-control-requirements list --project-material-type 123

  # Filter by quality control classification
  xbe view project-material-type-quality-control-requirements list --quality-control-classification 456

  # Output as JSON
  xbe view project-material-type-quality-control-requirements list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectMaterialTypeQualityControlRequirementsList,
	}
	initProjectMaterialTypeQualityControlRequirementsListFlags(cmd)
	return cmd
}

func init() {
	projectMaterialTypeQualityControlRequirementsCmd.AddCommand(newProjectMaterialTypeQualityControlRequirementsListCmd())
}

func initProjectMaterialTypeQualityControlRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-material-type", "", "Filter by project material type ID")
	cmd.Flags().String("quality-control-classification", "", "Filter by quality control classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectMaterialTypeQualityControlRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectMaterialTypeQualityControlRequirementsListOptions(cmd)
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
	query.Set("fields[project-material-type-quality-control-requirements]", "project-material-type,quality-control-classification,note")
	query.Set("include", "project-material-type,quality-control-classification")
	query.Set("fields[project-material-types]", "display-name,explicit-display-name")
	query.Set("fields[quality-control-classifications]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "id")
	}
	setFilterIfPresent(query, "filter[project-material-type]", opts.ProjectMaterialType)
	setFilterIfPresent(query, "filter[quality-control-classification]", opts.QualityControlClassification)

	body, _, err := client.Get(cmd.Context(), "/v1/project-material-type-quality-control-requirements", query)
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

	rows := buildProjectMaterialTypeQualityControlRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectMaterialTypeQualityControlRequirementsTable(cmd, rows)
}

func parseProjectMaterialTypeQualityControlRequirementsListOptions(cmd *cobra.Command) (projectMaterialTypeQualityControlRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectMaterialType, _ := cmd.Flags().GetString("project-material-type")
	qualityControlClassification, _ := cmd.Flags().GetString("quality-control-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectMaterialTypeQualityControlRequirementsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Sort:                         sort,
		ProjectMaterialType:          projectMaterialType,
		QualityControlClassification: qualityControlClassification,
	}, nil
}

func buildProjectMaterialTypeQualityControlRequirementRows(resp jsonAPIResponse) []projectMaterialTypeQualityControlRequirementRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectMaterialTypeQualityControlRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectMaterialTypeQualityControlRequirementRow{
			ID:   resource.ID,
			Note: stringAttr(resource.Attributes, "note"),
		}

		if rel, ok := resource.Relationships["project-material-type"]; ok && rel.Data != nil {
			row.ProjectMaterialTypeID = rel.Data.ID
			if projectMaterialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectMaterialTypeName = projectMaterialTypeLabel(projectMaterialType.Attributes)
			}
		}

		if rel, ok := resource.Relationships["quality-control-classification"]; ok && rel.Data != nil {
			row.QualityControlClassificationID = rel.Data.ID
			if classification, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.QualityControlClassificationName = stringAttr(classification.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderProjectMaterialTypeQualityControlRequirementsTable(cmd *cobra.Command, rows []projectMaterialTypeQualityControlRequirementRow) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT MATERIAL TYPE\tQUALITY CONTROL CLASSIFICATION\tNOTE")
	for _, row := range rows {
		projectMaterialType := firstNonEmpty(row.ProjectMaterialTypeName, row.ProjectMaterialTypeID)
		classification := firstNonEmpty(row.QualityControlClassificationName, row.QualityControlClassificationID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(projectMaterialType, 28),
			truncateString(classification, 28),
			truncateString(row.Note, 40),
		)
	}
	return writer.Flush()
}

func projectMaterialTypeLabel(attrs map[string]any) string {
	explicit := stringAttr(attrs, "explicit-display-name")
	if explicit != "" {
		return explicit
	}

	return stringAttr(attrs, "display-name")
}
