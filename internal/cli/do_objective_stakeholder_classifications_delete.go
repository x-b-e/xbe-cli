package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doObjectiveStakeholderClassificationsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoObjectiveStakeholderClassificationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an objective stakeholder classification",
		Long: `Delete an objective stakeholder classification.

Provide the classification ID as an argument. The --confirm flag is required
to prevent accidental deletions.

Note: Only admin users can delete objective stakeholder classifications.

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete an objective stakeholder classification
  xbe do objective-stakeholder-classifications delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoObjectiveStakeholderClassificationsDelete,
	}
	initDoObjectiveStakeholderClassificationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doObjectiveStakeholderClassificationsCmd.AddCommand(newDoObjectiveStakeholderClassificationsDeleteCmd())
}

func initDoObjectiveStakeholderClassificationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoObjectiveStakeholderClassificationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoObjectiveStakeholderClassificationsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete an objective stakeholder classification")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/objective-stakeholder-classifications/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted objective stakeholder classification %s\n", opts.ID)
	return nil
}

func parseDoObjectiveStakeholderClassificationsDeleteOptions(cmd *cobra.Command, args []string) (doObjectiveStakeholderClassificationsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doObjectiveStakeholderClassificationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
