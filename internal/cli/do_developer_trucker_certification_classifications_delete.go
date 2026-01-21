package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doDeveloperTruckerCertificationClassificationsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoDeveloperTruckerCertificationClassificationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a developer trucker certification classification",
		Long: `Delete a developer trucker certification classification.

Provide the classification ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Note: Classifications that have certifications assigned cannot be deleted.`,
		Example: `  # Delete a developer trucker certification classification
  xbe do developer-trucker-certification-classifications delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperTruckerCertificationClassificationsDelete,
	}
	initDoDeveloperTruckerCertificationClassificationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doDeveloperTruckerCertificationClassificationsCmd.AddCommand(newDoDeveloperTruckerCertificationClassificationsDeleteCmd())
}

func initDoDeveloperTruckerCertificationClassificationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperTruckerCertificationClassificationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperTruckerCertificationClassificationsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a developer trucker certification classification")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/developer-trucker-certification-classifications/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted developer trucker certification classification %s\n", opts.ID)
	return nil
}

func parseDoDeveloperTruckerCertificationClassificationsDeleteOptions(cmd *cobra.Command, args []string) (doDeveloperTruckerCertificationClassificationsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperTruckerCertificationClassificationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
