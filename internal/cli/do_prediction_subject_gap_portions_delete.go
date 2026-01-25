package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPredictionSubjectGapPortionsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoPredictionSubjectGapPortionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a prediction subject gap portion",
		Long: `Delete a prediction subject gap portion.

Required flags:
  --confirm  Confirm deletion

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete a prediction subject gap portion
  xbe do prediction-subject-gap-portions delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectGapPortionsDelete,
	}
	initDoPredictionSubjectGapPortionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectGapPortionsCmd.AddCommand(newDoPredictionSubjectGapPortionsDeleteCmd())
}

func initDoPredictionSubjectGapPortionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectGapPortionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectGapPortionsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/prediction-subject-gap-portions/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted prediction subject gap portion %s\n", opts.ID)
	return nil
}

func parseDoPredictionSubjectGapPortionsDeleteOptions(cmd *cobra.Command, args []string) (doPredictionSubjectGapPortionsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectGapPortionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
