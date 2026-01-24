package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectPhaseCostItemPriceEstimatesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoProjectPhaseCostItemPriceEstimatesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project phase cost item price estimate",
		Long: `Delete a project phase cost item price estimate.

Requires --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete a price estimate
  xbe do project-phase-cost-item-price-estimates delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhaseCostItemPriceEstimatesDelete,
	}
	initDoProjectPhaseCostItemPriceEstimatesDeleteFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemPriceEstimatesCmd.AddCommand(newDoProjectPhaseCostItemPriceEstimatesDeleteCmd())
}

func initDoProjectPhaseCostItemPriceEstimatesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("confirm")
}

func runDoProjectPhaseCostItemPriceEstimatesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhaseCostItemPriceEstimatesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("deletion requires --confirm flag")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/project-phase-cost-item-price-estimates/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project phase cost item price estimate %s\n", opts.ID)
	return nil
}

func parseDoProjectPhaseCostItemPriceEstimatesDeleteOptions(cmd *cobra.Command, args []string) (doProjectPhaseCostItemPriceEstimatesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemPriceEstimatesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
