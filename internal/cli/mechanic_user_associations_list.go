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

type mechanicUserAssociationsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	User                   string
	MaintenanceRequirement string
}

type mechanicUserAssociationRow struct {
	ID                       string `json:"id"`
	UserID                   string `json:"user_id,omitempty"`
	MaintenanceRequirementID string `json:"maintenance_requirement_id,omitempty"`
}

func newMechanicUserAssociationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mechanic user associations",
		Long: `List mechanic user associations with filtering and pagination.

Mechanic user associations link users to maintenance requirements.

Output Columns:
  ID           Record identifier
  USER         User ID
  REQUIREMENT  Maintenance requirement ID

Filters:
  --user                   Filter by user ID (comma-separated for multiple)
  --maintenance-requirement Filter by maintenance requirement ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List mechanic user associations
  xbe view mechanic-user-associations list

  # Filter by user
  xbe view mechanic-user-associations list --user 123

  # Filter by maintenance requirement
  xbe view mechanic-user-associations list --maintenance-requirement 456

  # Output as JSON
  xbe view mechanic-user-associations list --json`,
		Args: cobra.NoArgs,
		RunE: runMechanicUserAssociationsList,
	}
	initMechanicUserAssociationsListFlags(cmd)
	return cmd
}

func init() {
	mechanicUserAssociationsCmd.AddCommand(newMechanicUserAssociationsListCmd())
}

func initMechanicUserAssociationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID (comma-separated for multiple)")
	cmd.Flags().String("maintenance-requirement", "", "Filter by maintenance requirement ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMechanicUserAssociationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMechanicUserAssociationsListOptions(cmd)
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
	query.Set("fields[mechanic-user-associations]", "user,maintenance-requirement")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[maintenance-requirement]", opts.MaintenanceRequirement)

	body, _, err := client.Get(cmd.Context(), "/v1/mechanic-user-associations", query)
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

	rows := buildMechanicUserAssociationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMechanicUserAssociationsTable(cmd, rows)
}

func parseMechanicUserAssociationsListOptions(cmd *cobra.Command) (mechanicUserAssociationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return mechanicUserAssociationsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		User:                   user,
		MaintenanceRequirement: maintenanceRequirement,
	}, nil
}

func buildMechanicUserAssociationRows(resp jsonAPIResponse) []mechanicUserAssociationRow {
	rows := make([]mechanicUserAssociationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := mechanicUserAssociationRow{ID: resource.ID}

		row.UserID = relationshipIDFromMap(resource.Relationships, "user")
		row.MaintenanceRequirementID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement")

		rows = append(rows, row)
	}
	return rows
}

func renderMechanicUserAssociationsTable(cmd *cobra.Command, rows []mechanicUserAssociationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No mechanic user associations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tREQUIREMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.UserID,
			row.MaintenanceRequirementID,
		)
	}
	return writer.Flush()
}
