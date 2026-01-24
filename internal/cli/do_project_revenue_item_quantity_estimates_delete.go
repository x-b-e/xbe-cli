package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectRevenueItemQuantityEstimatesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoProjectRevenueItemQuantityEstimatesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project revenue item quantity estimate",
		Long: `Delete a project revenue item quantity estimate.

Requires the --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a quantity estimate
  xbe do project-revenue-item-quantity-estimates delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectRevenueItemQuantityEstimatesDelete,
	}
	initDoProjectRevenueItemQuantityEstimatesDeleteFlags(cmd)
	return cmd
}

func init() {
	doProjectRevenueItemQuantityEstimatesCmd.AddCommand(newDoProjectRevenueItemQuantityEstimatesDeleteCmd())
}

func initDoProjectRevenueItemQuantityEstimatesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectRevenueItemQuantityEstimatesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectRevenueItemQuantityEstimatesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a project revenue item quantity estimate")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/project-revenue-item-quantity-estimates/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project revenue item quantity estimate %s\n", opts.ID)
	return nil
}

func parseDoProjectRevenueItemQuantityEstimatesDeleteOptions(cmd *cobra.Command, args []string) (doProjectRevenueItemQuantityEstimatesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRevenueItemQuantityEstimatesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
