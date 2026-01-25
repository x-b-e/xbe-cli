package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doLineupScenarioLineupsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoLineupScenarioLineupsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a lineup scenario lineup",
		Long: `Delete a lineup scenario lineup.

Provide the lineup scenario lineup ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a lineup scenario lineup
  xbe do lineup-scenario-lineups delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupScenarioLineupsDelete,
	}
	initDoLineupScenarioLineupsDeleteFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioLineupsCmd.AddCommand(newDoLineupScenarioLineupsDeleteCmd())
}

func initDoLineupScenarioLineupsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioLineupsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupScenarioLineupsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a lineup scenario lineup")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/lineup-scenario-lineups/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted lineup scenario lineup %s\n", opts.ID)
	return nil
}

func parseDoLineupScenarioLineupsDeleteOptions(cmd *cobra.Command, args []string) (doLineupScenarioLineupsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioLineupsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
