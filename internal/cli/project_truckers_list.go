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

type projectTruckersListOptions struct {
	BaseURL                                                string
	Token                                                  string
	JSON                                                   bool
	NoAuth                                                 bool
	Limit                                                  int
	Offset                                                 int
	Sort                                                   string
	Project                                                string
	Trucker                                                string
	IsExcludedFromTimeCardPayrollCertificationRequirements string
	CreatedAtMin                                           string
	CreatedAtMax                                           string
	UpdatedAtMin                                           string
	UpdatedAtMax                                           string
}

type projectTruckerRow struct {
	ID                                                     string `json:"id"`
	Project                                                string `json:"project_id,omitempty"`
	Trucker                                                string `json:"trucker_id,omitempty"`
	IsExcludedFromTimeCardPayrollCertificationRequirements bool   `json:"is_excluded_from_time_card_payroll_certification_requirements"`
	CreatedAt                                              string `json:"created_at,omitempty"`
	UpdatedAt                                              string `json:"updated_at,omitempty"`
}

func newProjectTruckersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project truckers",
		Long: `List project truckers with filtering and pagination.

Output Columns:
  ID         Project trucker identifier
  PROJECT    Project ID
  TRUCKER    Trucker ID
  EXCLUDED   Excluded from time card payroll certification requirements
  CREATED AT Creation timestamp
  UPDATED AT Last update timestamp

Filters:
  --project                                                   Filter by project ID
  --trucker                                                   Filter by trucker ID
  --is-excluded-from-time-card-payroll-certification-requirements Filter by exclusion flag (true/false)
  --created-at-min                                            Filter by created-at on/after (ISO 8601)
  --created-at-max                                            Filter by created-at on/before (ISO 8601)
  --updated-at-min                                            Filter by updated-at on/after (ISO 8601)
  --updated-at-max                                            Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project truckers
  xbe view project-truckers list

  # Filter by project
  xbe view project-truckers list --project 123

  # Filter by trucker
  xbe view project-truckers list --trucker 456

  # Filter by exclusion flag
  xbe view project-truckers list --is-excluded-from-time-card-payroll-certification-requirements true

  # Output as JSON
  xbe view project-truckers list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTruckersList,
	}
	initProjectTruckersListFlags(cmd)
	return cmd
}

func init() {
	projectTruckersCmd.AddCommand(newProjectTruckersListCmd())
}

func initProjectTruckersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("is-excluded-from-time-card-payroll-certification-requirements", "", "Filter by exclusion flag (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTruckersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTruckersListOptions(cmd)
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
	query.Set("fields[project-truckers]", "created-at,updated-at,project,trucker,is-excluded-from-time-card-payroll-certification-requirements")

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
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[is_excluded_from_time_card_payroll_certification_requirements]", opts.IsExcludedFromTimeCardPayrollCertificationRequirements)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/project-truckers", query)
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

	rows := buildProjectTruckerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTruckersTable(cmd, rows)
}

func parseProjectTruckersListOptions(cmd *cobra.Command) (projectTruckersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	trucker, _ := cmd.Flags().GetString("trucker")
	exclusion, _ := cmd.Flags().GetString("is-excluded-from-time-card-payroll-certification-requirements")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTruckersListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Project: project,
		Trucker: trucker,
		IsExcludedFromTimeCardPayrollCertificationRequirements: exclusion,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildProjectTruckerRows(resp jsonAPIResponse) []projectTruckerRow {
	rows := make([]projectTruckerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTruckerRow{
			ID: resource.ID,
			IsExcludedFromTimeCardPayrollCertificationRequirements: boolAttr(attrs, "is-excluded-from-time-card-payroll-certification-requirements"),
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
		}
		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.Project = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.Trucker = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectTruckersTable(cmd *cobra.Command, rows []projectTruckerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project truckers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tTRUCKER\tEXCLUDED\tCREATED AT\tUPDATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%t\t%s\t%s\n",
			row.ID,
			row.Project,
			row.Trucker,
			row.IsExcludedFromTimeCardPayrollCertificationRequirements,
			row.CreatedAt,
			row.UpdatedAt,
		)
	}
	return writer.Flush()
}
