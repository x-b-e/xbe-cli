package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMechanicUserAssociationsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
	JSON    bool
}

func newDoMechanicUserAssociationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a mechanic user association",
		Long: `Delete a mechanic user association.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a mechanic user association
  xbe do mechanic-user-associations delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMechanicUserAssociationsDelete,
	}
	initDoMechanicUserAssociationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doMechanicUserAssociationsCmd.AddCommand(newDoMechanicUserAssociationsDeleteCmd())
}

func initDoMechanicUserAssociationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMechanicUserAssociationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMechanicUserAssociationsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/mechanic-user-associations/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"deleted": true, "id": opts.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted mechanic user association %s\n", opts.ID)
	return nil
}

func parseDoMechanicUserAssociationsDeleteOptions(cmd *cobra.Command, args []string) (doMechanicUserAssociationsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMechanicUserAssociationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
		JSON:    jsonOut,
	}, nil
}
