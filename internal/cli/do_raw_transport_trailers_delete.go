package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawTransportTrailersDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
	ID      string
}

func newDoRawTransportTrailersDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a raw transport trailer",
		Long: `Delete a raw transport trailer.

Requires the --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a raw transport trailer
  xbe do raw-transport-trailers delete 123 --confirm

  # Output JSON
  xbe do raw-transport-trailers delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawTransportTrailersDelete,
	}
	initDoRawTransportTrailersDeleteFlags(cmd)
	return cmd
}

func init() {
	doRawTransportTrailersCmd.AddCommand(newDoRawTransportTrailersDeleteCmd())
}

func initDoRawTransportTrailersDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawTransportTrailersDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawTransportTrailersDeleteOptions(cmd, args[0])
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/raw-transport-trailers/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	result := map[string]any{
		"id":      opts.ID,
		"deleted": true,
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted raw transport trailer %s\n", opts.ID)
	return nil
}

func parseDoRawTransportTrailersDeleteOptions(cmd *cobra.Command, id string) (doRawTransportTrailersDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportTrailersDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
		ID:      id,
	}, nil
}
