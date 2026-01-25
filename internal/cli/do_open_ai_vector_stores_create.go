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

type doOpenAiVectorStoresCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	Purpose   string
	Scope     string
	ScopeType string
	ScopeID   string
}

func newDoOpenAiVectorStoresCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an OpenAI vector store",
		Long: `Create an OpenAI vector store.

Required flags:
  --purpose         Vector store purpose (platform_content, user_post_feed, prediction_subject_lowest_losing_bid_recap)

Optional flags:
  --scope           Scope in Type|ID format (e.g., UserPostFeed|123)
  --scope-type      Scope type (use with --scope-id)
  --scope-id        Scope ID (use with --scope-type)`,
		Example: `  # Create a vector store for a user post feed
  xbe do open-ai-vector-stores create \
    --purpose user_post_feed \
    --scope-type UserPostFeed \
    --scope-id 123`,
		Args: cobra.NoArgs,
		RunE: runDoOpenAiVectorStoresCreate,
	}
	initDoOpenAiVectorStoresCreateFlags(cmd)
	return cmd
}

func init() {
	doOpenAiVectorStoresCmd.AddCommand(newDoOpenAiVectorStoresCreateCmd())
}

func initDoOpenAiVectorStoresCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("purpose", "", "Vector store purpose (required)")
	cmd.Flags().String("scope", "", "Scope in Type|ID format")
	cmd.Flags().String("scope-type", "", "Scope type (use with --scope-id)")
	cmd.Flags().String("scope-id", "", "Scope ID (use with --scope-type)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOpenAiVectorStoresCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOpenAiVectorStoresCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.Purpose) == "" {
		err := fmt.Errorf("--purpose is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	scopeType := strings.TrimSpace(opts.ScopeType)
	scopeID := strings.TrimSpace(opts.ScopeID)

	if opts.Scope != "" {
		parsedType, parsedID, err := parseOpenAiVectorStoreScope(opts.Scope)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		scopeType = parsedType
		scopeID = parsedID
	}

	if scopeType != "" || scopeID != "" {
		if scopeType == "" || scopeID == "" {
			err := fmt.Errorf("--scope-type and --scope-id must be provided together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{
		"purpose": strings.TrimSpace(opts.Purpose),
	}

	relationships := map[string]any{}
	if scopeType != "" {
		relationshipType := normalizeOpenAiVectorStoreScopeTypeForRelationship(scopeType)
		if relationshipType == "" {
			relationshipType = scopeType
		}
		relationships["scope"] = map[string]any{
			"data": map[string]any{
				"type": relationshipType,
				"id":   scopeID,
			},
		}
	}

	data := map[string]any{
		"type":       "open-ai-vector-stores",
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/open-ai-vector-stores", jsonBody)
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

	row := openAiVectorStoreRowFromResource(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created open ai vector store %s\n", row.ID)
	return nil
}

func parseDoOpenAiVectorStoresCreateOptions(cmd *cobra.Command) (doOpenAiVectorStoresCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	purpose, _ := cmd.Flags().GetString("purpose")
	scope, _ := cmd.Flags().GetString("scope")
	scopeType, _ := cmd.Flags().GetString("scope-type")
	scopeID, _ := cmd.Flags().GetString("scope-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenAiVectorStoresCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		Purpose:   purpose,
		Scope:     scope,
		ScopeType: scopeType,
		ScopeID:   scopeID,
	}, nil
}
