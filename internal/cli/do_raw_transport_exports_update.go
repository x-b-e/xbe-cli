package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawTransportExportsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
}

func newDoRawTransportExportsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a raw transport export",
		Long: `Update a raw transport export.

Raw transport exports are immutable after creation and do not expose
any writable fields for updates. Delete and recreate an export instead.

Arguments:
  <id>  Raw transport export ID (required).`,
		Example: `  # Attempt to update a raw transport export (not supported)
  xbe do raw-transport-exports update 123`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawTransportExportsUpdate,
	}
	initDoRawTransportExportsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportExportsCmd.AddCommand(newDoRawTransportExportsUpdateCmd())
}

func initDoRawTransportExportsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawTransportExportsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawTransportExportsUpdateOptions(cmd, args)
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

	err = fmt.Errorf("raw transport exports do not support updates; delete and recreate instead")
	fmt.Fprintln(cmd.ErrOrStderr(), err)
	return err
}

func parseDoRawTransportExportsUpdateOptions(cmd *cobra.Command, args []string) (doRawTransportExportsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doRawTransportExportsUpdateOptions{}, fmt.Errorf("raw transport export id is required")
	}

	return doRawTransportExportsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      id,
	}, nil
}
