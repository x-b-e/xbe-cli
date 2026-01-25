package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTractorOdometerReadingsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoTractorOdometerReadingsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a tractor odometer reading",
		Long: `Delete a tractor odometer reading.

Required:
  --confirm   Confirm deletion

Arguments:
  <id>    The odometer reading ID (required).`,
		Example: `  # Delete a tractor odometer reading
  xbe do tractor-odometer-readings delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTractorOdometerReadingsDelete,
	}
	initDoTractorOdometerReadingsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTractorOdometerReadingsCmd.AddCommand(newDoTractorOdometerReadingsDeleteCmd())
}

func initDoTractorOdometerReadingsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorOdometerReadingsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorOdometerReadingsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm is required to delete a tractor odometer reading")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/tractor-odometer-readings/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted tractor odometer reading %s\n", opts.ID)
	return nil
}

func parseDoTractorOdometerReadingsDeleteOptions(cmd *cobra.Command, args []string) (doTractorOdometerReadingsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doTractorOdometerReadingsDeleteOptions{}, fmt.Errorf("tractor odometer reading id is required")
	}

	return doTractorOdometerReadingsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      id,
		Confirm: confirm,
	}, nil
}
