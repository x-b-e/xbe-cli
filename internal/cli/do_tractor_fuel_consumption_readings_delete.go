package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTractorFuelConsumptionReadingsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoTractorFuelConsumptionReadingsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a tractor fuel consumption reading",
		Long: `Delete a tractor fuel consumption reading.

Requires the --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a tractor fuel consumption reading
  xbe do tractor-fuel-consumption-readings delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTractorFuelConsumptionReadingsDelete,
	}
	initDoTractorFuelConsumptionReadingsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTractorFuelConsumptionReadingsCmd.AddCommand(newDoTractorFuelConsumptionReadingsDeleteCmd())
}

func initDoTractorFuelConsumptionReadingsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("confirm")
}

func runDoTractorFuelConsumptionReadingsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorFuelConsumptionReadingsDeleteOptions(cmd, args)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	path := fmt.Sprintf("/v1/tractor-fuel-consumption-readings/%s", opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted tractor fuel consumption reading %s\n", opts.ID)
	return nil
}

func parseDoTractorFuelConsumptionReadingsDeleteOptions(cmd *cobra.Command, args []string) (doTractorFuelConsumptionReadingsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorFuelConsumptionReadingsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
