package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectTransportPlanAssignmentRulesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoProjectTransportPlanAssignmentRulesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project transport plan assignment rule",
		Long: `Delete a project transport plan assignment rule.

Required flags:
  --confirm    Confirm deletion

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete an assignment rule
  xbe do project-transport-plan-assignment-rules delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanAssignmentRulesDelete,
	}
	initDoProjectTransportPlanAssignmentRulesDeleteFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanAssignmentRulesCmd.AddCommand(newDoProjectTransportPlanAssignmentRulesDeleteCmd())
}

func initDoProjectTransportPlanAssignmentRulesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanAssignmentRulesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanAssignmentRulesDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/project-transport-plan-assignment-rules/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project transport plan assignment rule %s\n", opts.ID)
	return nil
}

func parseDoProjectTransportPlanAssignmentRulesDeleteOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanAssignmentRulesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanAssignmentRulesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
