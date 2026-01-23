package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doServiceTypeUnitOfMeasureQuantitiesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoServiceTypeUnitOfMeasureQuantitiesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a service type unit of measure quantity",
		Long: `Delete a service type unit of measure quantity.

Deleting quantities is only allowed when the quantified resource permits it.

Arguments:
  <id>  The service type unit of measure quantity ID (required).

Flags:
  --confirm  Confirm deletion`,
		Example: `  # Delete a service type unit of measure quantity
  xbe do service-type-unit-of-measure-quantities delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoServiceTypeUnitOfMeasureQuantitiesDelete,
	}
	initDoServiceTypeUnitOfMeasureQuantitiesDeleteFlags(cmd)
	return cmd
}

func init() {
	doServiceTypeUnitOfMeasureQuantitiesCmd.AddCommand(newDoServiceTypeUnitOfMeasureQuantitiesDeleteCmd())
}

func initDoServiceTypeUnitOfMeasureQuantitiesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoServiceTypeUnitOfMeasureQuantitiesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoServiceTypeUnitOfMeasureQuantitiesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm is required to delete a service type unit of measure quantity")
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
	body, _, err := client.Delete(cmd.Context(), "/v1/service-type-unit-of-measure-quantities/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		fmt.Fprintf(cmd.OutOrStdout(), "{\"id\":\"%s\",\"deleted\":true}\n", opts.ID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted service type unit of measure quantity %s\n", opts.ID)
	return nil
}

func parseDoServiceTypeUnitOfMeasureQuantitiesDeleteOptions(cmd *cobra.Command, args []string) (doServiceTypeUnitOfMeasureQuantitiesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doServiceTypeUnitOfMeasureQuantitiesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
