package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTimeSheetLineItemsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoTimeSheetLineItemsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a time sheet line item",
		Long: `Delete a time sheet line item.

Provide the time sheet line item ID as an argument. The --confirm flag is required
to delete the line item.

Arguments:
  <id>    The time sheet line item ID (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Delete a time sheet line item
  xbe do time-sheet-line-items delete 123 --confirm

  # Delete and output JSON
  xbe do time-sheet-line-items delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetLineItemsDelete,
	}
	initDoTimeSheetLineItemsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetLineItemsCmd.AddCommand(newDoTimeSheetLineItemsDeleteCmd())
}

func initDoTimeSheetLineItemsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetLineItemsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetLineItemsDeleteOptions(cmd, args)
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

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a time sheet line item")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/time-sheet-line-items/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		result := map[string]any{
			"id":      opts.ID,
			"deleted": true,
		}
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted time sheet line item %s\n", opts.ID)
	return nil
}

func parseDoTimeSheetLineItemsDeleteOptions(cmd *cobra.Command, args []string) (doTimeSheetLineItemsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	id := strings.TrimSpace(args[0])
	if id == "" {
		return doTimeSheetLineItemsDeleteOptions{}, fmt.Errorf("time sheet line item id is required")
	}

	return doTimeSheetLineItemsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      id,
		Confirm: confirm,
	}, nil
}
