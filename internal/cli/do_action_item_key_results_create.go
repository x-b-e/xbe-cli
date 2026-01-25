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

type doActionItemKeyResultsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ActionItem string
	KeyResult  string
}

func newDoActionItemKeyResultsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an action item key result link",
		Long: `Create an action item key result link.

Required flags:
  --action-item  Action item ID (required)
  --key-result   Key result ID (required)`,
		Example: `  # Link an action item to a key result
  xbe do action-item-key-results create \
    --action-item 123 \
    --key-result 456

  # JSON output
  xbe do action-item-key-results create \
    --action-item 123 \
    --key-result 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoActionItemKeyResultsCreate,
	}
	initDoActionItemKeyResultsCreateFlags(cmd)
	return cmd
}

func init() {
	doActionItemKeyResultsCmd.AddCommand(newDoActionItemKeyResultsCreateCmd())
}

func initDoActionItemKeyResultsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("action-item", "", "Action item ID (required)")
	cmd.Flags().String("key-result", "", "Key result ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemKeyResultsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoActionItemKeyResultsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ActionItem) == "" {
		err := fmt.Errorf("--action-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.KeyResult) == "" {
		err := fmt.Errorf("--key-result is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"action-item": map[string]any{
			"data": map[string]any{
				"type": "action-items",
				"id":   opts.ActionItem,
			},
		},
		"key-result": map[string]any{
			"data": map[string]any{
				"type": "key-results",
				"id":   opts.KeyResult,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "action-item-key-results",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/action-item-key-results", jsonBody)
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

	row := buildActionItemKeyResultRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created action item key result %s\n", row.ID)
	return nil
}

func parseDoActionItemKeyResultsCreateOptions(cmd *cobra.Command) (doActionItemKeyResultsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	actionItem, _ := cmd.Flags().GetString("action-item")
	keyResult, _ := cmd.Flags().GetString("key-result")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemKeyResultsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ActionItem: actionItem,
		KeyResult:  keyResult,
	}, nil
}

func buildActionItemKeyResultRowFromSingle(resp jsonAPISingleResponse) actionItemKeyResultRow {
	row := actionItemKeyResultRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["action-item"]; ok && rel.Data != nil {
		row.ActionItemID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["key-result"]; ok && rel.Data != nil {
		row.KeyResultID = rel.Data.ID
	}

	return row
}
