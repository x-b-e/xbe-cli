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

type actionItemTeamMembersListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	ActionItem          string
	User                string
	IsResponsiblePerson string
}

type actionItemTeamMemberRow struct {
	ID                  string `json:"id"`
	ActionItemID        string `json:"action_item_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
	IsResponsiblePerson bool   `json:"is_responsible_person"`
}

func newActionItemTeamMembersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List action item team members",
		Long: `List action item team members.

Output Columns:
  ID           Team member identifier
  ACTION ITEM  Action item ID
  USER         User ID
  RESPONSIBLE  Whether the user is the responsible person

Filters:
  --action-item            Filter by action item ID (comma-separated for multiple)
  --user                   Filter by user ID (comma-separated for multiple)
  --is-responsible-person  Filter by responsible person status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List team members
  xbe view action-item-team-members list

  # Filter by action item
  xbe view action-item-team-members list --action-item 123

  # Filter by user
  xbe view action-item-team-members list --user 456

  # Filter by responsible person
  xbe view action-item-team-members list --is-responsible-person true

  # Output as JSON
  xbe view action-item-team-members list --json`,
		Args: cobra.NoArgs,
		RunE: runActionItemTeamMembersList,
	}
	initActionItemTeamMembersListFlags(cmd)
	return cmd
}

func init() {
	actionItemTeamMembersCmd.AddCommand(newActionItemTeamMembersListCmd())
}

func initActionItemTeamMembersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("action-item", "", "Filter by action item ID (comma-separated for multiple)")
	cmd.Flags().String("user", "", "Filter by user ID (comma-separated for multiple)")
	cmd.Flags().String("is-responsible-person", "", "Filter by responsible person status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemTeamMembersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseActionItemTeamMembersListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[action-item-team-members]", "is-responsible-person,action-item,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[action-item]", opts.ActionItem)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[is-responsible-person]", opts.IsResponsiblePerson)

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-team-members", query)
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

	rows := buildActionItemTeamMemberRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderActionItemTeamMembersTable(cmd, rows)
}

func parseActionItemTeamMembersListOptions(cmd *cobra.Command) (actionItemTeamMembersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	actionItem, _ := cmd.Flags().GetString("action-item")
	user, _ := cmd.Flags().GetString("user")
	isResponsiblePerson, _ := cmd.Flags().GetString("is-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemTeamMembersListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		ActionItem:          actionItem,
		User:                user,
		IsResponsiblePerson: isResponsiblePerson,
	}, nil
}

func buildActionItemTeamMemberRows(resp jsonAPIResponse) []actionItemTeamMemberRow {
	rows := make([]actionItemTeamMemberRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := actionItemTeamMemberRow{
			ID:                  resource.ID,
			ActionItemID:        relationshipIDFromMap(resource.Relationships, "action-item"),
			UserID:              relationshipIDFromMap(resource.Relationships, "user"),
			IsResponsiblePerson: boolAttr(attrs, "is-responsible-person"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderActionItemTeamMembersTable(cmd *cobra.Command, rows []actionItemTeamMemberRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No action item team members found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tACTION_ITEM\tUSER\tRESPONSIBLE")
	for _, row := range rows {
		responsible := "no"
		if row.IsResponsiblePerson {
			responsible = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.ActionItemID,
			row.UserID,
			responsible,
		)
	}
	return writer.Flush()
}
