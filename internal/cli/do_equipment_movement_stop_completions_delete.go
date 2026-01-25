package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doEquipmentMovementStopCompletionsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoEquipmentMovementStopCompletionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an equipment movement stop completion",
		Long: `Delete an equipment movement stop completion.

This permanently deletes the stop completion.

Arguments:
  <id>    The stop completion ID (required).`,
		Example: `  # Delete a stop completion
  xbe do equipment-movement-stop-completions delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementStopCompletionsDelete,
	}
	initDoEquipmentMovementStopCompletionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementStopCompletionsCmd.AddCommand(newDoEquipmentMovementStopCompletionsDeleteCmd())
}

func initDoEquipmentMovementStopCompletionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementStopCompletionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementStopCompletionsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a stop completion")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/equipment-movement-stop-completions/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted equipment movement stop completion %s\n", opts.ID)
	return nil
}

func parseDoEquipmentMovementStopCompletionsDeleteOptions(cmd *cobra.Command, args []string) (doEquipmentMovementStopCompletionsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	id := strings.TrimSpace(args[0])
	if id == "" {
		return doEquipmentMovementStopCompletionsDeleteOptions{}, fmt.Errorf("equipment movement stop completion id is required")
	}

	return doEquipmentMovementStopCompletionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      id,
		Confirm: confirm,
	}, nil
}
