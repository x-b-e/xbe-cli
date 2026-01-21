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

type businessUnitMembershipsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	UserID         string
	MembershipID   string
	BusinessUnitID string
	Kind           string
	Sort           string
	Me             bool
}

type buMembershipRow struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	MembershipID     string `json:"membership_id,omitempty"`
	UserID           string `json:"user_id,omitempty"`
	UserName         string `json:"user_name,omitempty"`
	BusinessUnitID   string `json:"business_unit_id"`
	BusinessUnitName string `json:"business_unit_name"`
}

func newBusinessUnitMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List business unit memberships",
		Long: `List business unit memberships with filtering.

Shows which business units users have access to and their role in each.

Output Columns (table format):
  ID              Unique BU membership identifier
  USER            User name
  BUSINESS_UNIT   Business unit name
  KIND            Role (manager, technician, general)

Filtering:
  --me              Show only my memberships (current user)
  --user-id         Filter by user ID
  --membership-id   Filter by membership ID`,
		Example: `  # List all BU memberships you can see
  xbe view business-unit-memberships list

  # Show only my BU memberships
  xbe view business-unit-memberships list --me

  # Filter by user
  xbe view business-unit-memberships list --user-id 5724

  # Filter by membership
  xbe view business-unit-memberships list --membership-id 7627

  # Output as JSON
  xbe view business-unit-memberships list --json`,
		RunE: runBusinessUnitMembershipsList,
	}
	initBusinessUnitMembershipsListFlags(cmd)
	return cmd
}

func init() {
	businessUnitMembershipsCmd.AddCommand(newBusinessUnitMembershipsListCmd())
}

func initBusinessUnitMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("me", false, "Show only my memberships (current user)")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("user-id", "", "Filter by user ID")
	cmd.Flags().String("membership-id", "", "Filter by broker membership ID")
	cmd.Flags().String("bu-id", "", "Filter by business unit ID")
	cmd.Flags().String("kind", "", "Filter by role (manager, technician, general)")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBusinessUnitMembershipsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// If --me or --user-id is set, fetch user's membership IDs first
	var membershipIDs []string
	if opts.Me {
		if opts.UserID != "" {
			return fmt.Errorf("cannot use both --me and --user-id")
		}
		ids, err := getCurrentUserMembershipIDs(cmd, client)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		membershipIDs = ids
	} else if opts.UserID != "" {
		// Fetch membership IDs for the specified user
		ids, err := getMembershipIDsForUser(cmd, client, opts.UserID)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		membershipIDs = ids
	}

	query := url.Values{}
	query.Set("include", "business-unit,membership,membership.user")
	query.Set("fields[business-units]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Apply filters
	if len(membershipIDs) > 0 {
		query.Set("filter[membership]", strings.Join(membershipIDs, ","))
	} else {
		setFilterIfPresent(query, "filter[membership]", opts.MembershipID)
	}
	setFilterIfPresent(query, "filter[business_unit]", opts.BusinessUnitID)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-memberships", query)
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

	if opts.JSON {
		rows := buildBUMembershipRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBUMembershipsList(cmd, resp)
}

func parseBusinessUnitMembershipsListOptions(cmd *cobra.Command) (businessUnitMembershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	me, _ := cmd.Flags().GetBool("me")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	userID, _ := cmd.Flags().GetString("user-id")
	membershipID, _ := cmd.Flags().GetString("membership-id")
	businessUnitID, _ := cmd.Flags().GetString("bu-id")
	kind, _ := cmd.Flags().GetString("kind")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitMembershipsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		UserID:         userID,
		MembershipID:   membershipID,
		BusinessUnitID: businessUnitID,
		Kind:           kind,
		Sort:           sort,
		Me:             me,
	}, nil
}

func buildBUMembershipRows(resp jsonAPIResponse) []buMembershipRow {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]buMembershipRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes

		row := buMembershipRow{
			ID:   resource.ID,
			Kind: stringAttr(attrs, "kind"),
		}

		// Get business unit info
		if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
			row.BusinessUnitID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.BusinessUnitName = stringAttr(inc.Attributes, "company-name")
			}
		}

		// Get membership and user info
		if rel, ok := resource.Relationships["membership"]; ok && rel.Data != nil {
			row.MembershipID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				// Get user from membership
				if userRel, ok := inc.Relationships["user"]; ok && userRel.Data != nil {
					row.UserID = userRel.Data.ID
					userKey := resourceKey(userRel.Data.Type, userRel.Data.ID)
					if userInc, ok := included[userKey]; ok {
						row.UserName = stringAttr(userInc.Attributes, "name")
					}
				}
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderBUMembershipsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildBUMembershipRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No business unit memberships found.")
		return nil
	}

	const userMax = 25
	const buMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tBUSINESS_UNIT\tKIND")
	for _, row := range rows {
		kind := row.Kind
		if kind == "" {
			kind = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.UserName, userMax),
			truncateString(row.BusinessUnitName, buMax),
			kind,
		)
	}
	return writer.Flush()
}

func getCurrentUserMembershipIDs(cmd *cobra.Command, client *api.Client) ([]string, error) {
	// First get current user ID
	userBody, _, err := client.Get(cmd.Context(), "/v1/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	var userResp jsonAPISingleResponse
	if err := json.Unmarshal(userBody, &userResp); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	if userResp.Data.ID == "" {
		return nil, fmt.Errorf("no user ID in response")
	}

	return getMembershipIDsForUser(cmd, client, userResp.Data.ID)
}

func getMembershipIDsForUser(cmd *cobra.Command, client *api.Client, userID string) ([]string, error) {
	query := url.Values{}
	query.Set("filter[user]", userID)
	query.Set("fields[memberships]", "id")
	query.Set("page[limit]", "500")

	membershipsBody, _, err := client.Get(cmd.Context(), "/v1/memberships", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	var membershipsResp jsonAPIResponse
	if err := json.Unmarshal(membershipsBody, &membershipsResp); err != nil {
		return nil, fmt.Errorf("failed to parse memberships response: %w", err)
	}

	ids := make([]string, 0, len(membershipsResp.Data))
	for _, m := range membershipsResp.Data {
		ids = append(ids, m.ID)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no memberships found for user %s", userID)
	}

	return ids, nil
}
