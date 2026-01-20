package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doActionItemsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoActionItemsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an action item",
		Long: `Delete an action item (soft delete).

This performs a soft delete by setting the deleted_at timestamp. The action item
will no longer appear in normal listings but remains in the database.

The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The action item ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete an action item
  xbe do action-items delete 123 --confirm

  # Get JSON output of deleted record
  xbe do action-items delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoActionItemsDelete,
	}
	initDoActionItemsDeleteFlags(cmd)
	return cmd
}

func init() {
	doActionItemsCmd.AddCommand(newDoActionItemsDeleteCmd())
}

func initDoActionItemsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemsDeleteOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require --confirm flag
	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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
		return fmt.Errorf("action item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// First, fetch the record so we can show what was deleted
	query := url.Values{}
	query.Set("fields[action-items]", "title,status,kind")

	getBody, _, err := client.Get(cmd.Context(), "/v1/action-items/"+id, query)
	if err != nil {
		if len(getBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(getBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var getResp jsonAPISingleResponse
	if err := json.Unmarshal(getBody, &getResp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Store title for confirmation message
	title := stringAttr(getResp.Data.Attributes, "title")

	// Soft delete by setting deleted_at
	requestBody := map[string]any{
		"data": map[string]any{
			"id":   id,
			"type": "action-items",
			"attributes": map[string]any{
				"deleted-at": time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	patchBody, _, err := client.Patch(cmd.Context(), "/v1/action-items/"+id, jsonBody)
	if err != nil {
		if len(patchBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(patchBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(patchBody, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildActionItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted action item %s (%s)\n", id, title)
	return nil
}

func parseDoActionItemsDeleteOptions(cmd *cobra.Command) (doActionItemsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
