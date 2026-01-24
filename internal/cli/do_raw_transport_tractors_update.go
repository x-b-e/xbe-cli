package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawTransportTractorsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
}

func newDoRawTransportTractorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a raw transport tractor",
		Long: `Update a raw transport tractor.

Raw transport tractors are immutable after creation and do not expose
any writable fields for updates. Delete and recreate a tractor instead.

Arguments:
  <id>  Raw transport tractor ID (required).`,
		Example: `  # Attempt to update a raw transport tractor (not supported)
  xbe do raw-transport-tractors update 123`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawTransportTractorsUpdate,
	}
	initDoRawTransportTractorsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportTractorsCmd.AddCommand(newDoRawTransportTractorsUpdateCmd())
}

func initDoRawTransportTractorsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawTransportTractorsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawTransportTractorsUpdateOptions(cmd, args)
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

	err = fmt.Errorf("raw transport tractors do not support updates; delete and recreate instead")
	fmt.Fprintln(cmd.ErrOrStderr(), err)
	return err
}

func parseDoRawTransportTractorsUpdateOptions(cmd *cobra.Command, args []string) (doRawTransportTractorsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doRawTransportTractorsUpdateOptions{}, fmt.Errorf("raw transport tractor id is required")
	}

	return doRawTransportTractorsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      id,
	}, nil
}
