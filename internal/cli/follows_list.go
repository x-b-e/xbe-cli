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

type followsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Follower     string
	Creator      string
	CreatorType  string
	CreatorID    string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type followRow struct {
	ID            string `json:"id"`
	FollowerID    string `json:"follower_id,omitempty"`
	FollowerName  string `json:"follower_name,omitempty"`
	FollowerEmail string `json:"follower_email,omitempty"`
	CreatorType   string `json:"creator_type,omitempty"`
	CreatorID     string `json:"creator_id,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

func newFollowsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List follows",
		Long: `List follow relationships with filtering and pagination.

Output Columns:
  ID        Follow identifier
  FOLLOWER  User who follows
  CREATOR   Creator type and ID
  CREATED   Follow creation time

Filters:
  --follower          Filter by follower user ID
  --creator           Filter by creator (Type|ID, e.g., Project|123)
  --creator-type      Filter by creator type (e.g., Project)
  --creator-id        Filter by creator ID (requires --creator-type)
  --created-at-min    Filter by created-at on/after (ISO 8601)
  --created-at-max    Filter by created-at on/before (ISO 8601)
  --is-created-at     Filter by has created-at (true/false)
  --updated-at-min    Filter by updated-at on/after (ISO 8601)
  --updated-at-max    Filter by updated-at on/before (ISO 8601)
  --is-updated-at     Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List follows
  xbe view follows list

  # Filter by follower
  xbe view follows list --follower 123

  # Filter by creator
  xbe view follows list --creator Project|456

  # Output as JSON
  xbe view follows list --json`,
		Args: cobra.NoArgs,
		RunE: runFollowsList,
	}
	initFollowsListFlags(cmd)
	return cmd
}

func init() {
	followsCmd.AddCommand(newFollowsListCmd())
}

func initFollowsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("follower", "", "Filter by follower user ID")
	cmd.Flags().String("creator", "", "Filter by creator (Type|ID, e.g., Project|123)")
	cmd.Flags().String("creator-type", "", "Filter by creator type (e.g., Project)")
	cmd.Flags().String("creator-id", "", "Filter by creator ID (requires --creator-type)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFollowsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseFollowsListOptions(cmd)
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

	if opts.Creator != "" && (opts.CreatorType != "" || opts.CreatorID != "") {
		return fmt.Errorf("--creator cannot be combined with --creator-type or --creator-id")
	}
	if opts.Creator == "" && opts.CreatorID != "" && opts.CreatorType == "" {
		return fmt.Errorf("--creator-type is required when --creator-id is set")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[follows]", "follower,creator,created-at,updated-at")
	query.Set("include", "follower")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[follower]", opts.Follower)

	if opts.Creator != "" {
		creatorFilter := strings.TrimSpace(opts.Creator)
		parts := strings.SplitN(creatorFilter, "|", 2)
		if len(parts) == 2 {
			normalizedType := normalizeResourceTypeForFilter(parts[0])
			if normalizedType != "" {
				creatorFilter = normalizedType + "|" + strings.TrimSpace(parts[1])
			}
		}
		query.Set("filter[creator]", creatorFilter)
	} else {
		creatorType := normalizeResourceTypeForFilter(opts.CreatorType)
		if creatorType != "" && opts.CreatorID != "" {
			query.Set("filter[creator]", creatorType+"|"+opts.CreatorID)
		} else {
			setFilterIfPresent(query, "filter[creator-type]", creatorType)
		}
	}

	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/follows", query)
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

	rows := buildFollowRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderFollowsTable(cmd, rows)
}

func parseFollowsListOptions(cmd *cobra.Command) (followsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	follower, _ := cmd.Flags().GetString("follower")
	creator, _ := cmd.Flags().GetString("creator")
	creatorType, _ := cmd.Flags().GetString("creator-type")
	creatorID, _ := cmd.Flags().GetString("creator-id")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return followsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Follower:     follower,
		Creator:      creator,
		CreatorType:  creatorType,
		CreatorID:    creatorID,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildFollowRows(resp jsonAPIResponse) []followRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]followRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := followRow{
			ID:        resource.ID,
			CreatedAt: formatDateTime(stringAttr(resource.Attributes, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(resource.Attributes, "updated-at")),
		}

		if rel, ok := resource.Relationships["follower"]; ok && rel.Data != nil {
			row.FollowerID = rel.Data.ID
			if follower, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.FollowerName = stringAttr(follower.Attributes, "name")
				row.FollowerEmail = stringAttr(follower.Attributes, "email-address")
			}
		}

		if rel, ok := resource.Relationships["creator"]; ok && rel.Data != nil {
			row.CreatorType = rel.Data.Type
			row.CreatorID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func followRowFromSingle(resp jsonAPISingleResponse) followRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	resource := resp.Data
	row := followRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(resource.Attributes, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(resource.Attributes, "updated-at")),
	}

	if rel, ok := resource.Relationships["follower"]; ok && rel.Data != nil {
		row.FollowerID = rel.Data.ID
		if follower, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.FollowerName = stringAttr(follower.Attributes, "name")
			row.FollowerEmail = stringAttr(follower.Attributes, "email-address")
		}
	}

	if rel, ok := resource.Relationships["creator"]; ok && rel.Data != nil {
		row.CreatorType = rel.Data.Type
		row.CreatorID = rel.Data.ID
	}

	return row
}

func renderFollowsTable(cmd *cobra.Command, rows []followRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No follows found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tFOLLOWER\tCREATOR\tCREATED")
	for _, row := range rows {
		follower := firstNonEmpty(row.FollowerName, row.FollowerEmail, row.FollowerID)
		creator := formatPolymorphic(row.CreatorType, row.CreatorID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			follower,
			creator,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
