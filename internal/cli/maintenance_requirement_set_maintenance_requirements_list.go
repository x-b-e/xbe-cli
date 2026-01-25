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

type maintenanceRequirementSetMaintenanceRequirementsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	MaintenanceRequirementSet string
	MaintenanceRequirement    string
}

type maintenanceRequirementSetMaintenanceRequirementRow struct {
	ID                          string `json:"id"`
	MaintenanceRequirementSetID string `json:"maintenance_requirement_set_id,omitempty"`
	MaintenanceRequirementID    string `json:"maintenance_requirement_id,omitempty"`
}

func newMaintenanceRequirementSetMaintenanceRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement set maintenance requirements",
		Long: `List maintenance requirement set maintenance requirements with filtering and pagination.

Output Columns:
  ID          Record identifier
  SET         Maintenance requirement set ID
  REQUIREMENT Maintenance requirement ID

Filters:
  --maintenance-requirement-set  Filter by maintenance requirement set ID
  --maintenance-requirement      Filter by maintenance requirement ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List records
  xbe view maintenance-requirement-set-maintenance-requirements list

  # Filter by maintenance requirement set
  xbe view maintenance-requirement-set-maintenance-requirements list --maintenance-requirement-set 123

  # Filter by maintenance requirement
  xbe view maintenance-requirement-set-maintenance-requirements list --maintenance-requirement 456

  # Output as JSON
  xbe view maintenance-requirement-set-maintenance-requirements list --json`,
		Args: cobra.NoArgs,
		RunE: runMaintenanceRequirementSetMaintenanceRequirementsList,
	}
	initMaintenanceRequirementSetMaintenanceRequirementsListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementSetMaintenanceRequirementsCmd.AddCommand(newMaintenanceRequirementSetMaintenanceRequirementsListCmd())
}

func initMaintenanceRequirementSetMaintenanceRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("maintenance-requirement-set", "", "Filter by maintenance requirement set ID")
	cmd.Flags().String("maintenance-requirement", "", "Filter by maintenance requirement ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementSetMaintenanceRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRequirementSetMaintenanceRequirementsListOptions(cmd)
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
	query.Set("fields[maintenance-requirement-set-maintenance-requirements]", "maintenance-requirement-set,maintenance-requirement")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[maintenance-requirement-set]", opts.MaintenanceRequirementSet)
	setFilterIfPresent(query, "filter[maintenance-requirement]", opts.MaintenanceRequirement)

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-set-maintenance-requirements", query)
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

	rows := buildMaintenanceRequirementSetMaintenanceRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaintenanceRequirementSetMaintenanceRequirementsTable(cmd, rows)
}

func parseMaintenanceRequirementSetMaintenanceRequirementsListOptions(cmd *cobra.Command) (maintenanceRequirementSetMaintenanceRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	maintenanceRequirementSet, _ := cmd.Flags().GetString("maintenance-requirement-set")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementSetMaintenanceRequirementsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		MaintenanceRequirementSet: maintenanceRequirementSet,
		MaintenanceRequirement:    maintenanceRequirement,
	}, nil
}

func buildMaintenanceRequirementSetMaintenanceRequirementRows(resp jsonAPIResponse) []maintenanceRequirementSetMaintenanceRequirementRow {
	rows := make([]maintenanceRequirementSetMaintenanceRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := maintenanceRequirementSetMaintenanceRequirementRow{ID: resource.ID}

		row.MaintenanceRequirementSetID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement-set")
		row.MaintenanceRequirementID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement")

		rows = append(rows, row)
	}
	return rows
}

func renderMaintenanceRequirementSetMaintenanceRequirementsTable(cmd *cobra.Command, rows []maintenanceRequirementSetMaintenanceRequirementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No maintenance requirement set maintenance requirements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSET\tREQUIREMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.MaintenanceRequirementSetID,
			row.MaintenanceRequirementID,
		)
	}
	return writer.Flush()
}
