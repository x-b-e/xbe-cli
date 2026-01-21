package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCostIndexEntriesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoCostIndexEntriesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a cost index entry",
		Long: `Delete a cost index entry.

Provide the entry ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Note: Entries that have been used in adjustment calculations cannot be deleted.`,
		Example: `  # Delete a cost index entry
  xbe do cost-index-entries delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCostIndexEntriesDelete,
	}
	initDoCostIndexEntriesDeleteFlags(cmd)
	return cmd
}

func init() {
	doCostIndexEntriesCmd.AddCommand(newDoCostIndexEntriesDeleteCmd())
}

func initDoCostIndexEntriesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostIndexEntriesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostIndexEntriesDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a cost index entry")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/cost-index-entries/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted cost index entry %s\n", opts.ID)
	return nil
}

func parseDoCostIndexEntriesDeleteOptions(cmd *cobra.Command, args []string) (doCostIndexEntriesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostIndexEntriesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
