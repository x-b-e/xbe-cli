package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPredictionSubjectGapsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoPredictionSubjectGapsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a prediction subject gap",
		Long: `Delete a prediction subject gap.

Provide the gap ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete a prediction subject gap
  xbe do prediction-subject-gaps delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectGapsDelete,
	}
	initDoPredictionSubjectGapsDeleteFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectGapsCmd.AddCommand(newDoPredictionSubjectGapsDeleteCmd())
}

func initDoPredictionSubjectGapsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectGapsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectGapsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a prediction subject gap")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/prediction-subject-gaps/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted prediction subject gap %s\n", opts.ID)
	return nil
}

func parseDoPredictionSubjectGapsDeleteOptions(cmd *cobra.Command, args []string) (doPredictionSubjectGapsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectGapsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
