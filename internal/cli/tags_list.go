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

type tagsListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Name            string
	TagCategoryID   string
	TagCategorySlug string
	ForTaggableType string
}

func newTagsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tags",
		Long: `List tags with filtering and pagination.

Tags are labels that can be applied to various entities and are organized
into tag categories.

Output Columns:
  ID            Tag identifier
  NAME          Tag name
  TAG CATEGORY  Category the tag belongs to

Filters:
  --name               Filter by name (partial match, case-insensitive)
  --tag-category-id    Filter by tag category ID
  --tag-category-slug  Filter by tag category slug
  --for-taggable-type  Filter by entity type tags can apply to`,
		Example: `  # List all tags
  xbe view tags list

  # Filter by name
  xbe view tags list --name "urgent"

  # Filter by tag category ID
  xbe view tags list --tag-category-id 123

  # Filter by tag category slug
  xbe view tags list --tag-category-slug "priority"

  # Filter by taggable type
  xbe view tags list --for-taggable-type PredictionSubject

  # Output as JSON
  xbe view tags list --json`,
		RunE: runTagsList,
	}
	initTagsListFlags(cmd)
	return cmd
}

func init() {
	tagsCmd.AddCommand(newTagsListCmd())
}

func initTagsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("tag-category-id", "", "Filter by tag category ID")
	cmd.Flags().String("tag-category-slug", "", "Filter by tag category slug")
	cmd.Flags().String("for-taggable-type", "", "Filter by taggable entity type")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTagsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTagsListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[tags]", "name,tag-category")
	query.Set("include", "tag-category")
	query.Set("fields[tag-categories]", "name,slug")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[tag-category-id]", opts.TagCategoryID)
	setFilterIfPresent(query, "filter[tag-category-slug]", opts.TagCategorySlug)
	setFilterIfPresent(query, "filter[for-taggable-type]", opts.ForTaggableType)

	body, _, err := client.Get(cmd.Context(), "/v1/tags", query)
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

	rows := buildTagRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTagsTable(cmd, rows)
}

func parseTagsListOptions(cmd *cobra.Command) (tagsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	tagCategoryID, _ := cmd.Flags().GetString("tag-category-id")
	tagCategorySlug, _ := cmd.Flags().GetString("tag-category-slug")
	forTaggableType, _ := cmd.Flags().GetString("for-taggable-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tagsListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Name:            name,
		TagCategoryID:   tagCategoryID,
		TagCategorySlug: tagCategorySlug,
		ForTaggableType: forTaggableType,
	}, nil
}

type tagRow struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	TagCategoryID   string `json:"tag_category_id,omitempty"`
	TagCategoryName string `json:"tag_category_name,omitempty"`
	TagCategorySlug string `json:"tag_category_slug,omitempty"`
}

func buildTagRows(resp jsonAPIResponse) []tagRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]tagRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tagRow{
			ID:   resource.ID,
			Name: stringAttr(resource.Attributes, "name"),
		}

		if rel, ok := resource.Relationships["tag-category"]; ok && rel.Data != nil {
			row.TagCategoryID = rel.Data.ID
			if category, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.TagCategoryName = stringAttr(category.Attributes, "name")
				row.TagCategorySlug = stringAttr(category.Attributes, "slug")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTagsTable(cmd *cobra.Command, rows []tagRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tags found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tTAG CATEGORY")
	for _, row := range rows {
		category := row.TagCategoryName
		if category == "" && row.TagCategoryID != "" {
			category = row.TagCategoryID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(category, 30),
		)
	}
	return writer.Flush()
}
