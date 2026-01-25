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

type taggingsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	TagID        string
	TaggableType string
	TaggableID   string
}

type taggingRow struct {
	ID           string `json:"id"`
	TagID        string `json:"tag_id,omitempty"`
	TagName      string `json:"tag_name,omitempty"`
	TaggableType string `json:"taggable_type,omitempty"`
	TaggableID   string `json:"taggable_id,omitempty"`
}

func newTaggingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List taggings",
		Long: `List taggings with filtering and pagination.

Taggings associate tags with taggable resources.

Output Columns:
  ID        Tagging identifier
  TAG       Tag name and ID
  TAGGABLE  Taggable type and ID

Filters:
  --tag-id         Filter by tag ID
  --taggable-type  Filter by taggable type (class name, e.g., PredictionSubject)
  --taggable-id    Filter by taggable ID (used with --taggable-type)

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List taggings
  xbe view taggings list

  # Filter by tag
  xbe view taggings list --tag-id 123

  # Filter by taggable
  xbe view taggings list --taggable-type PredictionSubject --taggable-id 456

  # Output as JSON
  xbe view taggings list --json`,
		Args: cobra.NoArgs,
		RunE: runTaggingsList,
	}
	initTaggingsListFlags(cmd)
	return cmd
}

func init() {
	taggingsCmd.AddCommand(newTaggingsListCmd())
}

func initTaggingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("tag-id", "", "Filter by tag ID")
	cmd.Flags().String("taggable-type", "", "Filter by taggable type (class name)")
	cmd.Flags().String("taggable-id", "", "Filter by taggable ID (use with --taggable-type)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTaggingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTaggingsListOptions(cmd)
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
	query.Set("fields[taggings]", "tag,taggable")
	query.Set("include", "tag")
	query.Set("fields[tags]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[tag-id]", opts.TagID)
	if opts.TaggableType != "" && opts.TaggableID != "" {
		normalizedType := normalizeResourceTypeForFilter(opts.TaggableType)
		query.Set("filter[taggable]", normalizedType+"|"+opts.TaggableID)
	} else if opts.TaggableType != "" {
		normalizedType := normalizeResourceTypeForFilter(opts.TaggableType)
		query.Set("filter[taggable-type]", normalizedType)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/taggings", query)
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

	rows := buildTaggingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTaggingsTable(cmd, rows)
}

func parseTaggingsListOptions(cmd *cobra.Command) (taggingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	tagID, _ := cmd.Flags().GetString("tag-id")
	taggableType, _ := cmd.Flags().GetString("taggable-type")
	taggableID, _ := cmd.Flags().GetString("taggable-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return taggingsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		TagID:        tagID,
		TaggableType: taggableType,
		TaggableID:   taggableID,
	}, nil
}

func buildTaggingRows(resp jsonAPIResponse) []taggingRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]taggingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := taggingRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["tag"]; ok && rel.Data != nil {
			row.TagID = rel.Data.ID
			if tag, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.TagName = strings.TrimSpace(stringAttr(tag.Attributes, "name"))
			}
		}

		if rel, ok := resource.Relationships["taggable"]; ok && rel.Data != nil {
			row.TaggableType = rel.Data.Type
			row.TaggableID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTaggingsTable(cmd *cobra.Command, rows []taggingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No taggings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTAG\tTAGGABLE")
	for _, row := range rows {
		tagDisplay := row.TagID
		if row.TagName != "" && row.TagID != "" {
			tagDisplay = fmt.Sprintf("%s (%s)", row.TagName, row.TagID)
		} else if row.TagName != "" {
			tagDisplay = row.TagName
		}

		taggableDisplay := row.TaggableID
		if row.TaggableType != "" && row.TaggableID != "" {
			taggableDisplay = row.TaggableType + "/" + row.TaggableID
		} else if row.TaggableType != "" {
			taggableDisplay = row.TaggableType
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(tagDisplay, 36),
			truncateString(taggableDisplay, 36),
		)
	}
	return writer.Flush()
}
