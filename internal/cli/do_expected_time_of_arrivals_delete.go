package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doExpectedTimeOfArrivalsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoExpectedTimeOfArrivalsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an expected time of arrival",
		Long: `Delete an expected time of arrival.

Provide the expected time of arrival ID as an argument. The --confirm flag is
required to prevent accidental deletions.`,
		Example: `  # Delete an expected time of arrival
  xbe do expected-time-of-arrivals delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoExpectedTimeOfArrivalsDelete,
	}
	initDoExpectedTimeOfArrivalsDeleteFlags(cmd)
	return cmd
}

func init() {
	doExpectedTimeOfArrivalsCmd.AddCommand(newDoExpectedTimeOfArrivalsDeleteCmd())
}

func initDoExpectedTimeOfArrivalsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExpectedTimeOfArrivalsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExpectedTimeOfArrivalsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete an expected time of arrival")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/expected-time-of-arrivals/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted expected time of arrival %s\n", opts.ID)
	return nil
}

func parseDoExpectedTimeOfArrivalsDeleteOptions(cmd *cobra.Command, args []string) (doExpectedTimeOfArrivalsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExpectedTimeOfArrivalsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
