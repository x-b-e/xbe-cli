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

type driverManagersListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	Trucker           string
	ManagerMembership string
	ManagedMembership string
	Broker            string
	ManagerUser       string
	ManagedUser       string
	CreatedAtMin      string
	CreatedAtMax      string
	UpdatedAtMin      string
	UpdatedAtMax      string
}

type driverManagerRow struct {
	ID                  string `json:"id"`
	TruckerID           string `json:"trucker_id,omitempty"`
	ManagerMembershipID string `json:"manager_membership_id,omitempty"`
	ManagedMembershipID string `json:"managed_membership_id,omitempty"`
}

func newDriverManagersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver managers",
		Long: `List driver managers.

Output Columns:
  ID           Driver manager identifier
  TRUCKER      Trucker ID
  MANAGER MBR  Manager membership ID
  MANAGED MBR  Managed membership ID

Filters:
  --trucker             Filter by trucker ID
  --manager-membership  Filter by manager membership ID
  --managed-membership  Filter by managed membership ID
  --broker              Filter by broker ID
  --manager-user        Filter by manager user ID
  --managed-user        Filter by managed user ID
  --created-at-min      Filter by created-at on/after (ISO 8601)
  --created-at-max      Filter by created-at on/before (ISO 8601)
  --updated-at-min      Filter by updated-at on/after (ISO 8601)
  --updated-at-max      Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List driver managers
  xbe view driver-managers list

  # Filter by trucker
  xbe view driver-managers list --trucker 123

  # Filter by manager membership
  xbe view driver-managers list --manager-membership 456

  # Filter by managed membership
  xbe view driver-managers list --managed-membership 789

  # Filter by broker
  xbe view driver-managers list --broker 321

  # Filter by manager user
  xbe view driver-managers list --manager-user 654

  # Output as JSON
  xbe view driver-managers list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverManagersList,
	}
	initDriverManagersListFlags(cmd)
	return cmd
}

func init() {
	driverManagersCmd.AddCommand(newDriverManagersListCmd())
}

func initDriverManagersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("manager-membership", "", "Filter by manager membership ID")
	cmd.Flags().String("managed-membership", "", "Filter by managed membership ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("manager-user", "", "Filter by manager user ID")
	cmd.Flags().String("managed-user", "", "Filter by managed user ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverManagersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverManagersListOptions(cmd)
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

	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[manager-membership]", opts.ManagerMembership)
	setFilterIfPresent(query, "filter[managed-membership]", opts.ManagedMembership)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[manager-user]", opts.ManagerUser)
	setFilterIfPresent(query, "filter[managed-user]", opts.ManagedUser)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-managers", query)
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

	rows := buildDriverManagerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverManagersTable(cmd, rows)
}

func parseDriverManagersListOptions(cmd *cobra.Command) (driverManagersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trucker, _ := cmd.Flags().GetString("trucker")
	managerMembership, _ := cmd.Flags().GetString("manager-membership")
	managedMembership, _ := cmd.Flags().GetString("managed-membership")
	broker, _ := cmd.Flags().GetString("broker")
	managerUser, _ := cmd.Flags().GetString("manager-user")
	managedUser, _ := cmd.Flags().GetString("managed-user")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverManagersListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		Trucker:           trucker,
		ManagerMembership: managerMembership,
		ManagedMembership: managedMembership,
		Broker:            broker,
		ManagerUser:       managerUser,
		ManagedUser:       managedUser,
		CreatedAtMin:      createdAtMin,
		CreatedAtMax:      createdAtMax,
		UpdatedAtMin:      updatedAtMin,
		UpdatedAtMax:      updatedAtMax,
	}, nil
}

func buildDriverManagerRows(resp jsonAPIResponse) []driverManagerRow {
	rows := make([]driverManagerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := driverManagerRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["manager-membership"]; ok && rel.Data != nil {
			row.ManagerMembershipID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["managed-membership"]; ok && rel.Data != nil {
			row.ManagedMembershipID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildDriverManagerRowFromSingle(resp jsonAPISingleResponse) driverManagerRow {
	resource := resp.Data
	row := driverManagerRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["manager-membership"]; ok && rel.Data != nil {
		row.ManagerMembershipID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["managed-membership"]; ok && rel.Data != nil {
		row.ManagedMembershipID = rel.Data.ID
	}

	return row
}

func renderDriverManagersTable(cmd *cobra.Command, rows []driverManagerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver managers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRUCKER\tMANAGER MBR\tMANAGED MBR")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.TruckerID,
			row.ManagerMembershipID,
			row.ManagedMembershipID,
		)
	}
	return writer.Flush()
}
