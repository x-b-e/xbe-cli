package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawTransportOrdersUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
}

func newDoRawTransportOrdersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a raw transport order",
		Long: `Update a raw transport order.

Raw transport orders are create-only in the API, so no fields are writable
after creation. Use create to ingest a new payload and delete to remove it.

Arguments:
  <id>    Raw transport order ID (required)`,
		Example: `  # Update is not supported
  xbe do raw-transport-orders update 123`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawTransportOrdersUpdate,
	}
	initDoRawTransportOrdersUpdateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportOrdersCmd.AddCommand(newDoRawTransportOrdersUpdateCmd())
}

func initDoRawTransportOrdersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawTransportOrdersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawTransportOrdersUpdateOptions(cmd, args)
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

	err = fmt.Errorf("raw transport orders are create-only; no fields can be updated")
	fmt.Fprintln(cmd.ErrOrStderr(), err)
	return err
}

func parseDoRawTransportOrdersUpdateOptions(cmd *cobra.Command, args []string) (doRawTransportOrdersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportOrdersUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
	}, nil
}
