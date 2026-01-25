package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawTransportExportsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
}

func newDoRawTransportExportsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a raw transport export",
		Long: `Delete a raw transport export.

Raw transport exports are read-only after creation and do not support deletion.

Arguments:
  <id>  Raw transport export ID (required).`,
		Example: `  # Attempt to delete a raw transport export (not supported)
  xbe do raw-transport-exports delete 123`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawTransportExportsDelete,
	}
	initDoRawTransportExportsDeleteFlags(cmd)
	return cmd
}

func init() {
	doRawTransportExportsCmd.AddCommand(newDoRawTransportExportsDeleteCmd())
}

func initDoRawTransportExportsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawTransportExportsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawTransportExportsDeleteOptions(cmd, args)
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

	err = fmt.Errorf("raw transport exports do not support deletion")
	fmt.Fprintln(cmd.ErrOrStderr(), err)
	return err
}

func parseDoRawTransportExportsDeleteOptions(cmd *cobra.Command, args []string) (doRawTransportExportsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doRawTransportExportsDeleteOptions{}, fmt.Errorf("raw transport export id is required")
	}

	return doRawTransportExportsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      id,
	}, nil
}
