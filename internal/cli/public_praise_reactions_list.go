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

type publicPraiseReactionsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type publicPraiseReactionRow struct {
	ID                       string `json:"id"`
	PublicPraiseID           string `json:"public_praise_id,omitempty"`
	ReactionClassificationID string `json:"reaction_classification_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
}

func newPublicPraiseReactionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List public praise reactions",
		Long: `List public praise reactions.

Output Columns:
  ID                   Reaction identifier
  PUBLIC PRAISE        Public praise ID
  REACTION             Reaction classification ID
  CREATED BY           User ID who created the reaction
  CREATED AT           Creation timestamp

Filters:
  --created-at-min     Filter by created-at on/after (ISO 8601)
  --created-at-max     Filter by created-at on/before (ISO 8601)
  --updated-at-min     Filter by updated-at on/after (ISO 8601)
  --updated-at-max     Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List public praise reactions
  xbe view public-praise-reactions list

  # Output as JSON
  xbe view public-praise-reactions list --json`,
		Args: cobra.NoArgs,
		RunE: runPublicPraiseReactionsList,
	}
	initPublicPraiseReactionsListFlags(cmd)
	return cmd
}

func init() {
	publicPraiseReactionsCmd.AddCommand(newPublicPraiseReactionsListCmd())
}

func initPublicPraiseReactionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPublicPraiseReactionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePublicPraiseReactionsListOptions(cmd)
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
	query.Set("fields[public-praise-reactions]", "public-praise,created-by,reaction-classification,created-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/public-praise-reactions", query)
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

	rows := buildPublicPraiseReactionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPublicPraiseReactionsTable(cmd, rows)
}

func parsePublicPraiseReactionsListOptions(cmd *cobra.Command) (publicPraiseReactionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return publicPraiseReactionsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildPublicPraiseReactionRows(resp jsonAPIResponse) []publicPraiseReactionRow {
	rows := make([]publicPraiseReactionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildPublicPraiseReactionRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildPublicPraiseReactionRow(resource jsonAPIResource) publicPraiseReactionRow {
	row := publicPraiseReactionRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(resource.Attributes, "created-at")),
	}

	if rel, ok := resource.Relationships["public-praise"]; ok && rel.Data != nil {
		row.PublicPraiseID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["reaction-classification"]; ok && rel.Data != nil {
		row.ReactionClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderPublicPraiseReactionsTable(cmd *cobra.Command, rows []publicPraiseReactionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No public praise reactions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPUBLIC PRAISE\tREACTION\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.PublicPraiseID,
			row.ReactionClassificationID,
			row.CreatedByID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
