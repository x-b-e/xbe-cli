package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type openAiVectorStoresShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type openAiVectorStoreDetails struct {
	ID                 string `json:"id"`
	OpenAiID           string `json:"open_ai_id,omitempty"`
	Purpose            string `json:"purpose,omitempty"`
	ChunkOverlapTokens int    `json:"chunk_overlap_tokens,omitempty"`
	MaxChunkSizeTokens int    `json:"max_chunk_size_tokens,omitempty"`
	ScopeType          string `json:"scope_type,omitempty"`
	ScopeID            string `json:"scope_id,omitempty"`
}

func newOpenAiVectorStoresShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show OpenAI vector store details",
		Long: `Show the full details of an OpenAI vector store.

Output Fields:
  ID
  Open AI ID
  Purpose
  Chunk Overlap Tokens
  Max Chunk Size Tokens
  Scope (type and ID)

Arguments:
  <id>    The vector store ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show vector store details
  xbe view open-ai-vector-stores show 123

  # Get JSON output
  xbe view open-ai-vector-stores show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOpenAiVectorStoresShow,
	}
	initOpenAiVectorStoresShowFlags(cmd)
	return cmd
}

func init() {
	openAiVectorStoresCmd.AddCommand(newOpenAiVectorStoresShowCmd())
}

func initOpenAiVectorStoresShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenAiVectorStoresShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOpenAiVectorStoresShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("open ai vector store id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/open-ai-vector-stores/"+id, nil)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildOpenAiVectorStoreDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOpenAiVectorStoreDetails(cmd, details)
}

func parseOpenAiVectorStoresShowOptions(cmd *cobra.Command) (openAiVectorStoresShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openAiVectorStoresShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOpenAiVectorStoreDetails(resp jsonAPISingleResponse) openAiVectorStoreDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := openAiVectorStoreDetails{
		ID:                 resource.ID,
		OpenAiID:           stringAttr(attrs, "open-ai-id"),
		Purpose:            stringAttr(attrs, "purpose"),
		ChunkOverlapTokens: intAttr(attrs, "chunk-overlap-tokens"),
		MaxChunkSizeTokens: intAttr(attrs, "max-chunk-size-tokens"),
		ScopeType:          stringAttr(attrs, "scope-type"),
		ScopeID:            stringAttr(attrs, "scope-id"),
	}

	if rel, ok := resource.Relationships["scope"]; ok && rel.Data != nil {
		details.ScopeType = rel.Data.Type
		details.ScopeID = rel.Data.ID
	}

	return details
}

func renderOpenAiVectorStoreDetails(cmd *cobra.Command, details openAiVectorStoreDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OpenAiID != "" {
		fmt.Fprintf(out, "Open AI ID: %s\n", details.OpenAiID)
	}
	if details.Purpose != "" {
		fmt.Fprintf(out, "Purpose: %s\n", details.Purpose)
	}
	if details.ChunkOverlapTokens != 0 {
		fmt.Fprintf(out, "Chunk Overlap Tokens: %d\n", details.ChunkOverlapTokens)
	}
	if details.MaxChunkSizeTokens != 0 {
		fmt.Fprintf(out, "Max Chunk Size Tokens: %d\n", details.MaxChunkSizeTokens)
	}
	if details.ScopeType != "" && details.ScopeID != "" {
		fmt.Fprintf(out, "Scope: %s/%s\n", details.ScopeType, details.ScopeID)
	}

	return nil
}
