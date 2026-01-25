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

type projectLaborClassificationsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Project             string
	LaborClassification string
}

type projectLaborClassificationRow struct {
	ID                      string `json:"id"`
	ProjectID               string `json:"project_id,omitempty"`
	ProjectName             string `json:"project_name,omitempty"`
	LaborClassificationID   string `json:"labor_classification_id,omitempty"`
	LaborClassificationName string `json:"labor_classification_name,omitempty"`
	BasicHourlyRate         string `json:"basic_hourly_rate,omitempty"`
	FringeHourlyRate        string `json:"fringe_hourly_rate,omitempty"`
	PrevailingHourlyRate    string `json:"prevailing_hourly_rate,omitempty"`
}

func newProjectLaborClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project labor classifications",
		Long: `List project labor classifications with filtering and pagination.

Project labor classifications link a project to a labor classification and
capture hourly rates.

Output Columns:
  ID          Classification identifier
  PROJECT     Project name or ID
  LABOR       Labor classification name or ID
  BASIC       Basic hourly rate
  FRINGE      Fringe hourly rate
  PREVAILING  Prevailing hourly rate

Filters:
  --project               Filter by project ID
  --labor-classification  Filter by labor classification ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project labor classifications
  xbe view project-labor-classifications list

  # Filter by project
  xbe view project-labor-classifications list --project 123

  # Filter by labor classification
  xbe view project-labor-classifications list --labor-classification 456

  # Output as JSON
  xbe view project-labor-classifications list --json`,
		RunE: runProjectLaborClassificationsList,
	}
	initProjectLaborClassificationsListFlags(cmd)
	return cmd
}

func init() {
	projectLaborClassificationsCmd.AddCommand(newProjectLaborClassificationsListCmd())
}

func initProjectLaborClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("labor-classification", "", "Filter by labor classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectLaborClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectLaborClassificationsListOptions(cmd)
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
	query.Set("fields[project-labor-classifications]", "project,labor-classification,basic-hourly-rate,fringe-hourly-rate,prevailing-hourly-rate")
	query.Set("include", "project,labor-classification")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[labor-classifications]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[labor_classification]", opts.LaborClassification)

	body, _, err := client.Get(cmd.Context(), "/v1/project-labor-classifications", query)
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

	rows := buildProjectLaborClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectLaborClassificationsTable(cmd, rows)
}

func parseProjectLaborClassificationsListOptions(cmd *cobra.Command) (projectLaborClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	laborClassification, _ := cmd.Flags().GetString("labor-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectLaborClassificationsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Project:             project,
		LaborClassification: laborClassification,
	}, nil
}

func buildProjectLaborClassificationRows(resp jsonAPIResponse) []projectLaborClassificationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectLaborClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectLaborClassificationRow{
			ID:                   resource.ID,
			BasicHourlyRate:      stringAttr(resource.Attributes, "basic-hourly-rate"),
			FringeHourlyRate:     stringAttr(resource.Attributes, "fringe-hourly-rate"),
			PrevailingHourlyRate: stringAttr(resource.Attributes, "prevailing-hourly-rate"),
		}

		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
			if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectName = firstNonEmpty(
					stringAttr(project.Attributes, "name"),
					stringAttr(project.Attributes, "number"),
				)
			}
		}

		if rel, ok := resource.Relationships["labor-classification"]; ok && rel.Data != nil {
			row.LaborClassificationID = rel.Data.ID
			if labor, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.LaborClassificationName = firstNonEmpty(
					stringAttr(labor.Attributes, "name"),
					stringAttr(labor.Attributes, "abbreviation"),
				)
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func buildProjectLaborClassificationRowFromSingle(resp jsonAPISingleResponse) projectLaborClassificationRow {
	attrs := resp.Data.Attributes
	row := projectLaborClassificationRow{
		ID:                   resp.Data.ID,
		BasicHourlyRate:      stringAttr(attrs, "basic-hourly-rate"),
		FringeHourlyRate:     stringAttr(attrs, "fringe-hourly-rate"),
		PrevailingHourlyRate: stringAttr(attrs, "prevailing-hourly-rate"),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["labor-classification"]; ok && rel.Data != nil {
		row.LaborClassificationID = rel.Data.ID
	}

	return row
}

func renderProjectLaborClassificationsTable(cmd *cobra.Command, rows []projectLaborClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project labor classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tLABOR\tBASIC\tFRINGE\tPREVAILING")
	for _, row := range rows {
		project := row.ProjectName
		if project == "" {
			project = row.ProjectID
		}
		labor := row.LaborClassificationName
		if labor == "" {
			labor = row.LaborClassificationID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(project, 30),
			truncateString(labor, 30),
			truncateString(row.BasicHourlyRate, 12),
			truncateString(row.FringeHourlyRate, 12),
			truncateString(row.PrevailingHourlyRate, 12),
		)
	}
	return writer.Flush()
}
