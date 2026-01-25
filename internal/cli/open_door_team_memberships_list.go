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

type openDoorTeamMembershipsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Organization string
	Membership   string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
	IsCreatedAt  string
	IsUpdatedAt  string
}

type openDoorTeamMembershipRow struct {
	ID               string `json:"id"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	MembershipID     string `json:"membership_id,omitempty"`
}

func newOpenDoorTeamMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List open door team memberships",
		Long: `List open door team memberships with filtering and pagination.

Open door team memberships link memberships to organizations for Open Door issue access.

Output Columns:
  ID        Open door team membership identifier
  ORG TYPE  Organization type
  ORG ID    Organization ID
  MEMBERSHIP Membership ID

Filters:
  --organization   Filter by organization (Type|ID, e.g. Broker|123)
  --membership     Filter by membership ID
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)
  --is-created-at  Filter by presence of created-at (true/false)
  --is-updated-at  Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List open door team memberships
  xbe view open-door-team-memberships list

  # Filter by organization
  xbe view open-door-team-memberships list --organization Broker|123

  # Filter by membership
  xbe view open-door-team-memberships list --membership 456

  # Filter by created-at
  xbe view open-door-team-memberships list --created-at-min 2024-01-01T00:00:00Z --created-at-max 2024-12-31T23:59:59Z

  # Output as JSON
  xbe view open-door-team-memberships list --json`,
		Args: cobra.NoArgs,
		RunE: runOpenDoorTeamMembershipsList,
	}
	initOpenDoorTeamMembershipsListFlags(cmd)
	return cmd
}

func init() {
	openDoorTeamMembershipsCmd.AddCommand(newOpenDoorTeamMembershipsListCmd())
}

func initOpenDoorTeamMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Broker|123)")
	cmd.Flags().String("membership", "", "Filter by membership ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenDoorTeamMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOpenDoorTeamMembershipsListOptions(cmd)
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
	query.Set("fields[open-door-team-memberships]", "organization,membership")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[membership]", opts.Membership)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/open-door-team-memberships", query)
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

	rows := buildOpenDoorTeamMembershipRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOpenDoorTeamMembershipsTable(cmd, rows)
}

func parseOpenDoorTeamMembershipsListOptions(cmd *cobra.Command) (openDoorTeamMembershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	membership, _ := cmd.Flags().GetString("membership")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openDoorTeamMembershipsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Organization: organization,
		Membership:   membership,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsCreatedAt:  isCreatedAt,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildOpenDoorTeamMembershipRows(resp jsonAPIResponse) []openDoorTeamMembershipRow {
	rows := make([]openDoorTeamMembershipRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := openDoorTeamMembershipRow{
			ID:           resource.ID,
			MembershipID: relationshipIDFromMap(resource.Relationships, "membership"),
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderOpenDoorTeamMembershipsTable(cmd *cobra.Command, rows []openDoorTeamMembershipRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No open door team memberships found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORG TYPE\tORG ID\tMEMBERSHIP")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.OrganizationType,
			row.OrganizationID,
			row.MembershipID,
		)
	}
	return writer.Flush()
}
