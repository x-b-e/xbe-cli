package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doIncidentHeadlineSuggestionsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoIncidentHeadlineSuggestionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an incident headline suggestion",
		Long: `Delete an incident headline suggestion.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete an incident headline suggestion
  xbe do incident-headline-suggestions delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoIncidentHeadlineSuggestionsDelete,
	}
	initDoIncidentHeadlineSuggestionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doIncidentHeadlineSuggestionsCmd.AddCommand(newDoIncidentHeadlineSuggestionsDeleteCmd())
}

func initDoIncidentHeadlineSuggestionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentHeadlineSuggestionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoIncidentHeadlineSuggestionsDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/incident-headline-suggestions/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted incident headline suggestion %s\n", opts.ID)
	return nil
}

func parseDoIncidentHeadlineSuggestionsDeleteOptions(cmd *cobra.Command, args []string) (doIncidentHeadlineSuggestionsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentHeadlineSuggestionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
