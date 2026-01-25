package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectPhaseDatesEstimatesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoProjectPhaseDatesEstimatesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project phase dates estimate",
		Long: `Delete a project phase dates estimate.

This action cannot be undone.`,
		Example: `  # Delete a dates estimate
  xbe do project-phase-dates-estimates delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhaseDatesEstimatesDelete,
	}
	initDoProjectPhaseDatesEstimatesDeleteFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseDatesEstimatesCmd.AddCommand(newDoProjectPhaseDatesEstimatesDeleteCmd())
}

func initDoProjectPhaseDatesEstimatesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseDatesEstimatesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhaseDatesEstimatesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("deletion requires --confirm")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/project-phase-dates-estimates/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project phase dates estimate %s\n", opts.ID)
	return nil
}

func parseDoProjectPhaseDatesEstimatesDeleteOptions(cmd *cobra.Command, args []string) (doProjectPhaseDatesEstimatesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseDatesEstimatesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
