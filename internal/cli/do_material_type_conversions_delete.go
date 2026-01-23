package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialTypeConversionsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoMaterialTypeConversionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a material type conversion",
		Long: `Delete a material type conversion.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a material type conversion
  xbe do material-type-conversions delete 123 --confirm

  # Get JSON output
  xbe do material-type-conversions delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTypeConversionsDelete,
	}
	initDoMaterialTypeConversionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypeConversionsCmd.AddCommand(newDoMaterialTypeConversionsDeleteCmd())
}

func initDoMaterialTypeConversionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTypeConversionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTypeConversionsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/material-type-conversions/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"id":      opts.ID,
			"deleted": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted material type conversion %s\n", opts.ID)
	return nil
}

func parseDoMaterialTypeConversionsDeleteOptions(cmd *cobra.Command, args []string) (doMaterialTypeConversionsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypeConversionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
