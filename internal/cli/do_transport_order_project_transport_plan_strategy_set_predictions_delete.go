package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTransportOrderProjectTransportPlanStrategySetPredictionsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoTransportOrderProjectTransportPlanStrategySetPredictionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a transport order strategy set prediction",
		Long: `Delete an existing transport order strategy set prediction.

Requires the --confirm flag to prevent accidental deletion.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Delete a prediction record
  xbe do transport-order-project-transport-plan-strategy-set-predictions delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTransportOrderProjectTransportPlanStrategySetPredictionsDelete,
	}
	initDoTransportOrderProjectTransportPlanStrategySetPredictionsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderProjectTransportPlanStrategySetPredictionsCmd.AddCommand(newDoTransportOrderProjectTransportPlanStrategySetPredictionsDeleteCmd())
}

func initDoTransportOrderProjectTransportPlanStrategySetPredictionsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("confirm")
}

func runDoTransportOrderProjectTransportPlanStrategySetPredictionsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportOrderProjectTransportPlanStrategySetPredictionsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := errors.New("deletion requires --confirm flag")
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

	path := fmt.Sprintf("/v1/transport-order-project-transport-plan-strategy-set-predictions/%s", opts.ID)
	body, _, err := client.Delete(cmd.Context(), path)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"id":      opts.ID,
			"deleted": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted transport order strategy set prediction %s\n", opts.ID)
	return nil
}

func parseDoTransportOrderProjectTransportPlanStrategySetPredictionsDeleteOptions(cmd *cobra.Command, args []string) (doTransportOrderProjectTransportPlanStrategySetPredictionsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderProjectTransportPlanStrategySetPredictionsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
