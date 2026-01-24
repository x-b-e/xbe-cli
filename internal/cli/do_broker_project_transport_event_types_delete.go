package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doBrokerProjectTransportEventTypesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoBrokerProjectTransportEventTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a broker project transport event type",
		Long: `Delete a broker project transport event type.

Provide the broker project transport event type ID as an argument. The --confirm
flag is required to prevent accidental deletions.`,
		Example: `  # Delete a broker project transport event type
  xbe do broker-project-transport-event-types delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerProjectTransportEventTypesDelete,
	}
	initDoBrokerProjectTransportEventTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doBrokerProjectTransportEventTypesCmd.AddCommand(newDoBrokerProjectTransportEventTypesDeleteCmd())
}

func initDoBrokerProjectTransportEventTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerProjectTransportEventTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerProjectTransportEventTypesDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a broker project transport event type")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/broker-project-transport-event-types/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted broker project transport event type %s\n", opts.ID)
	return nil
}

func parseDoBrokerProjectTransportEventTypesDeleteOptions(cmd *cobra.Command, args []string) (doBrokerProjectTransportEventTypesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerProjectTransportEventTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
