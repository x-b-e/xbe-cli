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

type doOpenAiVectorStoresUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	Purpose   string
	Scope     string
	ScopeType string
	ScopeID   string
}

func newDoOpenAiVectorStoresUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an OpenAI vector store",
		Long: `Update an existing OpenAI vector store.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The vector store ID (required)

Flags:
  --purpose         Update the purpose
  --scope           Update the scope (Type|ID)
  --scope-type      Update the scope type (use with --scope-id)
  --scope-id        Update the scope ID (use with --scope-type)`,
		Example: `  # Update the purpose
  xbe do open-ai-vector-stores update 123 --purpose platform_content

  # Update the scope
  xbe do open-ai-vector-stores update 123 --scope "UserPostFeed|456"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOpenAiVectorStoresUpdate,
	}
	initDoOpenAiVectorStoresUpdateFlags(cmd)
	return cmd
}

func init() {
	doOpenAiVectorStoresCmd.AddCommand(newDoOpenAiVectorStoresUpdateCmd())
}

func initDoOpenAiVectorStoresUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("purpose", "", "New purpose")
	cmd.Flags().String("scope", "", "Scope in Type|ID format")
	cmd.Flags().String("scope-type", "", "Scope type (use with --scope-id)")
	cmd.Flags().String("scope-id", "", "Scope ID (use with --scope-type)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOpenAiVectorStoresUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOpenAiVectorStoresUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("open ai vector store id is required")
	}

	if opts.Purpose == "" && opts.Scope == "" && opts.ScopeType == "" && opts.ScopeID == "" {
		err := fmt.Errorf("at least one field is required to update")
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

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Purpose) != "" {
		attributes["purpose"] = strings.TrimSpace(opts.Purpose)
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
		"type": "open-ai-vector-stores",
		"id":   id,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Patch(cmd.Context(), "/v1/open-ai-vector-stores/"+id, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated open ai vector store %s\n", row.ID)
	return nil
}

func parseDoOpenAiVectorStoresUpdateOptions(cmd *cobra.Command) (doOpenAiVectorStoresUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	purpose, _ := cmd.Flags().GetString("purpose")
	scope, _ := cmd.Flags().GetString("scope")
	scopeType, _ := cmd.Flags().GetString("scope-type")
	scopeID, _ := cmd.Flags().GetString("scope-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenAiVectorStoresUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		Purpose:   purpose,
		Scope:     scope,
		ScopeType: scopeType,
		ScopeID:   scopeID,
	}, nil
}
