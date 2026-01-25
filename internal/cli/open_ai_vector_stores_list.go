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

type openAiVectorStoresListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Purpose      string
	Scope        string
	ScopeType    string
	ScopeID      string
	NotScopeType string
}

type openAiVectorStoreRow struct {
	ID                 string `json:"id"`
	OpenAiID           string `json:"open_ai_id,omitempty"`
	Purpose            string `json:"purpose,omitempty"`
	ChunkOverlapTokens int    `json:"chunk_overlap_tokens,omitempty"`
	MaxChunkSizeTokens int    `json:"max_chunk_size_tokens,omitempty"`
	ScopeType          string `json:"scope_type,omitempty"`
	ScopeID            string `json:"scope_id,omitempty"`
}

func newOpenAiVectorStoresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OpenAI vector stores",
		Long: `List OpenAI vector stores with filtering and pagination.

Output Columns:
  ID               Vector store identifier
  PURPOSE          Vector store purpose
  OPEN AI ID       OpenAI vector store ID
  CHUNK OVERLAP    Chunk overlap tokens (if configured)
  MAX CHUNK SIZE   Max chunk size tokens (if configured)
  SCOPE            Scope type and ID

Filters:
  --purpose          Filter by purpose (platform_content, user_post_feed, prediction_subject_lowest_losing_bid_recap)
  --scope            Filter by scope (Type|ID)
  --scope-type       Filter by scope type (e.g., Broker, UserPostFeed)
  --scope-id         Filter by scope ID (use with --scope-type)
  --not-scope-type   Exclude scope types

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List vector stores
  xbe view open-ai-vector-stores list

  # Filter by purpose
  xbe view open-ai-vector-stores list --purpose user_post_feed

  # Filter by scope
  xbe view open-ai-vector-stores list --scope "UserPostFeed|123"

  # Output as JSON
  xbe view open-ai-vector-stores list --json`,
		Args: cobra.NoArgs,
		RunE: runOpenAiVectorStoresList,
	}
	initOpenAiVectorStoresListFlags(cmd)
	return cmd
}

func init() {
	openAiVectorStoresCmd.AddCommand(newOpenAiVectorStoresListCmd())
}

func initOpenAiVectorStoresListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("purpose", "", "Filter by purpose")
	cmd.Flags().String("scope", "", "Filter by scope (Type|ID)")
	cmd.Flags().String("scope-type", "", "Filter by scope type")
	cmd.Flags().String("scope-id", "", "Filter by scope ID (use with --scope-type)")
	cmd.Flags().String("not-scope-type", "", "Exclude scope types")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenAiVectorStoresList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOpenAiVectorStoresListOptions(cmd)
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
	query.Set("fields[open-ai-vector-stores]", "open-ai-id,purpose,chunk-overlap-tokens,max-chunk-size-tokens,scope,scope-type,scope-id")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[purpose]", opts.Purpose)

	scopeFilter := normalizePolymorphicFilterValue(opts.Scope)
	if scopeFilter == "" && opts.ScopeType != "" && opts.ScopeID != "" {
		resourceType := normalizeResourceTypeForFilter(opts.ScopeType)
		if resourceType == "" {
			resourceType = strings.TrimSpace(opts.ScopeType)
		}
		scopeFilter = resourceType + "|" + strings.TrimSpace(opts.ScopeID)
	}
	if scopeFilter != "" {
		query.Set("filter[scope]", scopeFilter)
	} else if opts.ScopeID != "" {
		return fmt.Errorf("--scope-id requires --scope-type or --scope")
	}

	if scopeFilter == "" {
		setFilterIfPresent(query, "filter[scope_type]", normalizeResourceTypeForFilter(opts.ScopeType))
	}
	if opts.NotScopeType != "" {
		setFilterIfPresent(query, "filter[not_scope_type]", normalizeResourceTypeForFilter(opts.NotScopeType))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/open-ai-vector-stores", query)
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

	rows := buildOpenAiVectorStoreRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOpenAiVectorStoresTable(cmd, rows)
}

func parseOpenAiVectorStoresListOptions(cmd *cobra.Command) (openAiVectorStoresListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	purpose, _ := cmd.Flags().GetString("purpose")
	scope, _ := cmd.Flags().GetString("scope")
	scopeType, _ := cmd.Flags().GetString("scope-type")
	scopeID, _ := cmd.Flags().GetString("scope-id")
	notScopeType, _ := cmd.Flags().GetString("not-scope-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openAiVectorStoresListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Purpose:      purpose,
		Scope:        scope,
		ScopeType:    scopeType,
		ScopeID:      scopeID,
		NotScopeType: notScopeType,
	}, nil
}

func buildOpenAiVectorStoreRows(resp jsonAPIResponse) []openAiVectorStoreRow {
	rows := make([]openAiVectorStoreRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, openAiVectorStoreRowFromResource(resource))
	}
	return rows
}

func openAiVectorStoreRowFromResource(resource jsonAPIResource) openAiVectorStoreRow {
	attrs := resource.Attributes
	row := openAiVectorStoreRow{
		ID:                 resource.ID,
		OpenAiID:           stringAttr(attrs, "open-ai-id"),
		Purpose:            stringAttr(attrs, "purpose"),
		ChunkOverlapTokens: intAttr(attrs, "chunk-overlap-tokens"),
		MaxChunkSizeTokens: intAttr(attrs, "max-chunk-size-tokens"),
		ScopeType:          stringAttr(attrs, "scope-type"),
		ScopeID:            stringAttr(attrs, "scope-id"),
	}

	if rel, ok := resource.Relationships["scope"]; ok && rel.Data != nil {
		row.ScopeType = rel.Data.Type
		row.ScopeID = rel.Data.ID
	}

	return row
}

func renderOpenAiVectorStoresTable(cmd *cobra.Command, rows []openAiVectorStoreRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No open ai vector stores found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPURPOSE\tOPEN AI ID\tCHUNK OVERLAP\tMAX CHUNK SIZE\tSCOPE")
	for _, row := range rows {
		scope := ""
		if row.ScopeType != "" && row.ScopeID != "" {
			scope = row.ScopeType + "/" + row.ScopeID
		}
		chunkOverlap := ""
		if row.ChunkOverlapTokens != 0 {
			chunkOverlap = strconv.Itoa(row.ChunkOverlapTokens)
		}
		maxChunkSize := ""
		if row.MaxChunkSizeTokens != 0 {
			maxChunkSize = strconv.Itoa(row.MaxChunkSizeTokens)
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Purpose, 28),
			truncateString(row.OpenAiID, 32),
			chunkOverlap,
			maxChunkSize,
			truncateString(scope, 36),
		)
	}
	return writer.Flush()
}
