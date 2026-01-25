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

type userCreatorFeedCreatorsListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	UserCreatorFeed string
	User            string
	UserID          string
	Creator         string
	CreatorType     string
	CreatorID       string
	NotCreatorType  string
	Follow          string
}

type userCreatorFeedCreatorRow struct {
	ID                string `json:"id"`
	Order             int    `json:"order,omitempty"`
	CreatorName       string `json:"creator_name,omitempty"`
	CreatorType       string `json:"creator_type,omitempty"`
	CreatorID         string `json:"creator_id,omitempty"`
	CreatorAvatarURL  string `json:"creator_avatar_url,omitempty"`
	UserCreatorFeedID string `json:"user_creator_feed_id,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	FollowID          string `json:"follow_id,omitempty"`
}

func newUserCreatorFeedCreatorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user creator feed creators",
		Long: `List user creator feed creators with filtering and pagination.

User creator feed creators represent the ordered creators shown in a user's
creator feed.

Output Columns:
  ID       User creator feed creator identifier
  ORDER    Position in the user creator feed
  CREATOR  Creator name (falls back to type/id)
  USER     User ID
  FEED     User creator feed ID
  FOLLOW   Follow ID (if present)

Filters:
  --user-creator-feed  Filter by user creator feed ID
  --user               Filter by user ID
  --user-id            Filter by user ID (via feed)
  --creator            Filter by creator (Type|ID, e.g. User|123)
  --creator-type       Filter by creator type (e.g. User, Broker)
  --creator-id         Filter by creator ID (use with --creator-type)
  --not-creator-type   Exclude by creator type
  --follow             Filter by follow ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List user creator feed creators
  xbe view user-creator-feed-creators list

  # Filter by user creator feed
  xbe view user-creator-feed-creators list --user-creator-feed 123

  # Filter by user
  xbe view user-creator-feed-creators list --user 456

  # Filter by creator
  xbe view user-creator-feed-creators list --creator "User|789"

  # Filter by creator type and ID
  xbe view user-creator-feed-creators list --creator-type User --creator-id 789

  # Output as JSON
  xbe view user-creator-feed-creators list --json`,
		Args: cobra.NoArgs,
		RunE: runUserCreatorFeedCreatorsList,
	}
	initUserCreatorFeedCreatorsListFlags(cmd)
	return cmd
}

func init() {
	userCreatorFeedCreatorsCmd.AddCommand(newUserCreatorFeedCreatorsListCmd())
}

func initUserCreatorFeedCreatorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user-creator-feed", "", "Filter by user creator feed ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("user-id", "", "Filter by user ID (via feed)")
	cmd.Flags().String("creator", "", "Filter by creator (Type|ID, e.g. User|123)")
	cmd.Flags().String("creator-type", "", "Filter by creator type (e.g. User, Broker)")
	cmd.Flags().String("creator-id", "", "Filter by creator ID (use with --creator-type)")
	cmd.Flags().String("not-creator-type", "", "Exclude by creator type")
	cmd.Flags().String("follow", "", "Filter by follow ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserCreatorFeedCreatorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserCreatorFeedCreatorsListOptions(cmd)
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
	query.Set("fields[user-creator-feed-creators]", "order,creator-name,creator-avatar-url,creator,user,user-creator-feed,follow")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[user-creator-feed]", opts.UserCreatorFeed)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[user-id]", opts.UserID)
	setFilterIfPresent(query, "filter[follow]", opts.Follow)

	creatorFilter := normalizePolymorphicFilter(opts.Creator)
	if creatorFilter == "" && opts.CreatorType != "" && opts.CreatorID != "" {
		creatorFilter = normalizeResourceTypeForFilter(opts.CreatorType) + "|" + opts.CreatorID
	}
	if creatorFilter != "" {
		query.Set("filter[creator]", creatorFilter)
		if opts.CreatorID != "" {
			query.Set("filter[creator-id]", creatorFilter)
		}
	} else if opts.CreatorID != "" {
		return fmt.Errorf("--creator-id requires --creator-type or --creator")
	}

	creatorType := normalizeResourceTypeForFilter(opts.CreatorType)
	setFilterIfPresent(query, "filter[creator-type]", creatorType)

	body, _, err := client.Get(cmd.Context(), "/v1/user-creator-feed-creators", query)
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

	rows := buildUserCreatorFeedCreatorRows(resp)
	if opts.NotCreatorType != "" {
		rows = filterUserCreatorFeedCreatorsByNotCreatorType(rows, normalizeResourceTypeForFilter(opts.NotCreatorType))
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserCreatorFeedCreatorsTable(cmd, rows)
}

func parseUserCreatorFeedCreatorsListOptions(cmd *cobra.Command) (userCreatorFeedCreatorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	userCreatorFeed, _ := cmd.Flags().GetString("user-creator-feed")
	user, _ := cmd.Flags().GetString("user")
	userID, _ := cmd.Flags().GetString("user-id")
	creator, _ := cmd.Flags().GetString("creator")
	creatorType, _ := cmd.Flags().GetString("creator-type")
	creatorID, _ := cmd.Flags().GetString("creator-id")
	notCreatorType, _ := cmd.Flags().GetString("not-creator-type")
	follow, _ := cmd.Flags().GetString("follow")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userCreatorFeedCreatorsListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		UserCreatorFeed: userCreatorFeed,
		User:            user,
		UserID:          userID,
		Creator:         creator,
		CreatorType:     creatorType,
		CreatorID:       creatorID,
		NotCreatorType:  notCreatorType,
		Follow:          follow,
	}, nil
}

func buildUserCreatorFeedCreatorRows(resp jsonAPIResponse) []userCreatorFeedCreatorRow {
	rows := make([]userCreatorFeedCreatorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := userCreatorFeedCreatorRow{
			ID:               resource.ID,
			Order:            intAttr(resource.Attributes, "order"),
			CreatorName:      strings.TrimSpace(stringAttr(resource.Attributes, "creator-name")),
			CreatorAvatarURL: strings.TrimSpace(stringAttr(resource.Attributes, "creator-avatar-url")),
		}

		if rel, ok := resource.Relationships["creator"]; ok && rel.Data != nil {
			row.CreatorType = rel.Data.Type
			row.CreatorID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["user-creator-feed"]; ok && rel.Data != nil {
			row.UserCreatorFeedID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["follow"]; ok && rel.Data != nil {
			row.FollowID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderUserCreatorFeedCreatorsTable(cmd *cobra.Command, rows []userCreatorFeedCreatorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user creator feed creators found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORDER\tCREATOR\tUSER\tFEED\tFOLLOW")
	for _, row := range rows {
		creatorLabel := formatCreatorLabel(row.CreatorName, row.CreatorType, row.CreatorID)
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Order,
			truncateString(creatorLabel, 36),
			row.UserID,
			row.UserCreatorFeedID,
			row.FollowID,
		)
	}
	return writer.Flush()
}

func formatCreatorLabel(name, creatorType, creatorID string) string {
	reference := formatCreatorReference(creatorType, creatorID)
	if name != "" && reference != "" {
		return fmt.Sprintf("%s (%s)", name, reference)
	}
	if name != "" {
		return name
	}
	return reference
}

func formatCreatorReference(creatorType, creatorID string) string {
	if creatorType != "" && creatorID != "" {
		return creatorType + "/" + creatorID
	}
	if creatorID != "" {
		return creatorID
	}
	return creatorType
}

func normalizePolymorphicFilter(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return value
	}
	typePart := normalizeResourceTypeForFilter(strings.TrimSpace(parts[0]))
	idPart := strings.TrimSpace(parts[1])
	if typePart == "" || idPart == "" {
		return value
	}
	return typePart + "|" + idPart
}

func filterUserCreatorFeedCreatorsByNotCreatorType(rows []userCreatorFeedCreatorRow, notCreatorType string) []userCreatorFeedCreatorRow {
	notCreatorType = strings.TrimSpace(notCreatorType)
	if notCreatorType == "" {
		return rows
	}
	filtered := make([]userCreatorFeedCreatorRow, 0, len(rows))
	for _, row := range rows {
		normalizedType := normalizeResourceTypeForFilter(row.CreatorType)
		if normalizedType != notCreatorType {
			filtered = append(filtered, row)
		}
	}
	return filtered
}
