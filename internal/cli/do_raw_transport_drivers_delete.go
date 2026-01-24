package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawTransportDriversDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
	ID      string
}

func newDoRawTransportDriversDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a raw transport driver",
		Long: `Delete a raw transport driver.

Requires the --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a raw transport driver
  xbe do raw-transport-drivers delete 123 --confirm

  # Output JSON
  xbe do raw-transport-drivers delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawTransportDriversDelete,
	}
	initDoRawTransportDriversDeleteFlags(cmd)
	return cmd
}

func init() {
	doRawTransportDriversCmd.AddCommand(newDoRawTransportDriversDeleteCmd())
}

func initDoRawTransportDriversDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawTransportDriversDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawTransportDriversDeleteOptions(cmd, args[0])
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/raw-transport-drivers/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	result := map[string]any{
		"id":      opts.ID,
		"deleted": true,
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted raw transport driver %s\n", opts.ID)
	return nil
}

func parseDoRawTransportDriversDeleteOptions(cmd *cobra.Command, id string) (doRawTransportDriversDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportDriversDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
		ID:      id,
	}, nil
}
