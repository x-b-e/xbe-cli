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

type projectTransportPlanDriverConfirmationsListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Sort                       string
	Status                     string
	ProjectTransportPlan       string
	ProjectTransportPlanDriver string
	Driver                     string
}

type projectTransportPlanDriverConfirmationRow struct {
	ID                           string `json:"id"`
	Status                       string `json:"status,omitempty"`
	ProjectTransportPlanID       string `json:"project_transport_plan_id,omitempty"`
	ProjectTransportPlanDriverID string `json:"project_transport_plan_driver_id,omitempty"`
	DriverID                     string `json:"driver_id,omitempty"`
	ConfirmedByID                string `json:"confirmed_by_id,omitempty"`
	ConfirmAtMax                 string `json:"confirm_at_max,omitempty"`
	ConfirmedAt                  string `json:"confirmed_at,omitempty"`
	Note                         string `json:"note,omitempty"`
}

func newProjectTransportPlanDriverConfirmationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan driver confirmations",
		Long: `List project transport plan driver confirmations.

Output Columns:
  ID            Confirmation identifier
  STATUS        Confirmation status
  PLAN          Project transport plan ID
  PLAN DRIVER   Project transport plan driver ID
  DRIVER        Driver (user) ID
  CONFIRM BY    Confirmed-by user ID
  CONFIRM MAX   Confirm-at max timestamp
  CONFIRMED AT  Confirmation timestamp

Filters:
  --status                       Filter by status (pending, confirmed, rejected, expired, superseded)
  --project-transport-plan        Filter by project transport plan ID
  --project-transport-plan-driver Filter by project transport plan driver ID
  --driver                        Filter by driver (user) ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List confirmations
  xbe view project-transport-plan-driver-confirmations list

  # Filter by status
  xbe view project-transport-plan-driver-confirmations list --status pending

  # Filter by project transport plan driver
  xbe view project-transport-plan-driver-confirmations list --project-transport-plan-driver 123

  # JSON output
  xbe view project-transport-plan-driver-confirmations list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanDriverConfirmationsList,
	}
	initProjectTransportPlanDriverConfirmationsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanDriverConfirmationsCmd.AddCommand(newProjectTransportPlanDriverConfirmationsListCmd())
}

func initProjectTransportPlanDriverConfirmationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("project-transport-plan-driver", "", "Filter by project transport plan driver ID")
	cmd.Flags().String("driver", "", "Filter by driver (user) ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanDriverConfirmationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanDriverConfirmationsListOptions(cmd)
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
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[project-transport-plan]", opts.ProjectTransportPlan)
	setFilterIfPresent(query, "filter[project-transport-plan-driver]", opts.ProjectTransportPlanDriver)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-driver-confirmations", query)
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

	rows := buildProjectTransportPlanDriverConfirmationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanDriverConfirmationsTable(cmd, rows)
}

func parseProjectTransportPlanDriverConfirmationsListOptions(cmd *cobra.Command) (projectTransportPlanDriverConfirmationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportPlanDriver, _ := cmd.Flags().GetString("project-transport-plan-driver")
	driver, _ := cmd.Flags().GetString("driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanDriverConfirmationsListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		Sort:                       sort,
		Status:                     status,
		ProjectTransportPlan:       projectTransportPlan,
		ProjectTransportPlanDriver: projectTransportPlanDriver,
		Driver:                     driver,
	}, nil
}

func buildProjectTransportPlanDriverConfirmationRows(resp jsonAPIResponse) []projectTransportPlanDriverConfirmationRow {
	rows := make([]projectTransportPlanDriverConfirmationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanDriverConfirmationRow{
			ID:           resource.ID,
			Status:       stringAttr(attrs, "status"),
			ConfirmAtMax: formatDateTime(stringAttr(attrs, "confirm-at-max")),
			ConfirmedAt:  formatDateTime(stringAttr(attrs, "confirmed-at")),
			Note:         stringAttr(attrs, "note"),
		}

		if rel, ok := resource.Relationships["project-transport-plan-driver"]; ok && rel.Data != nil {
			row.ProjectTransportPlanDriverID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
			row.ProjectTransportPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["confirmed-by"]; ok && rel.Data != nil {
			row.ConfirmedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildProjectTransportPlanDriverConfirmationRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanDriverConfirmationRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := projectTransportPlanDriverConfirmationRow{
		ID:           resource.ID,
		Status:       stringAttr(attrs, "status"),
		ConfirmAtMax: formatDateTime(stringAttr(attrs, "confirm-at-max")),
		ConfirmedAt:  formatDateTime(stringAttr(attrs, "confirmed-at")),
		Note:         stringAttr(attrs, "note"),
	}

	if rel, ok := resource.Relationships["project-transport-plan-driver"]; ok && rel.Data != nil {
		row.ProjectTransportPlanDriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		row.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["confirmed-by"]; ok && rel.Data != nil {
		row.ConfirmedByID = rel.Data.ID
	}

	return row
}

func renderProjectTransportPlanDriverConfirmationsTable(cmd *cobra.Command, rows []projectTransportPlanDriverConfirmationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan driver confirmations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tPLAN\tPLAN DRIVER\tDRIVER\tCONFIRM BY\tCONFIRM MAX\tCONFIRMED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.ProjectTransportPlanID,
			row.ProjectTransportPlanDriverID,
			row.DriverID,
			row.ConfirmedByID,
			row.ConfirmAtMax,
			row.ConfirmedAt,
		)
	}
	return writer.Flush()
}
