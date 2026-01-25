package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCustomerIncidentDefaultAssigneesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoCustomerIncidentDefaultAssigneesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a customer incident default assignee",
		Long: `Delete a customer incident default assignee.

Required flags:
  --confirm    Confirm deletion

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Delete a default assignee
  xbe do customer-incident-default-assignees delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerIncidentDefaultAssigneesDelete,
	}
	initDoCustomerIncidentDefaultAssigneesDeleteFlags(cmd)
	return cmd
}

func init() {
	doCustomerIncidentDefaultAssigneesCmd.AddCommand(newDoCustomerIncidentDefaultAssigneesDeleteCmd())
}

func initDoCustomerIncidentDefaultAssigneesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerIncidentDefaultAssigneesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerIncidentDefaultAssigneesDeleteOptions(cmd, args)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/customer-incident-default-assignees/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted customer incident default assignee %s\n", opts.ID)
	return nil
}

func parseDoCustomerIncidentDefaultAssigneesDeleteOptions(cmd *cobra.Command, args []string) (doCustomerIncidentDefaultAssigneesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerIncidentDefaultAssigneesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
