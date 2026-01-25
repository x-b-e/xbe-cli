package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doIncidentUnitOfMeasureQuantitiesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoIncidentUnitOfMeasureQuantitiesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an incident unit of measure quantity",
		Long: `Delete an incident unit of measure quantity.

Requires --confirm flag to prevent accidental deletion.`,
		Example: `  # Delete an incident unit of measure quantity
  xbe do incident-unit-of-measure-quantities delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoIncidentUnitOfMeasureQuantitiesDelete,
	}
	initDoIncidentUnitOfMeasureQuantitiesDeleteFlags(cmd)
	return cmd
}

func init() {
	doIncidentUnitOfMeasureQuantitiesCmd.AddCommand(newDoIncidentUnitOfMeasureQuantitiesDeleteCmd())
}

func initDoIncidentUnitOfMeasureQuantitiesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("confirm")
}

func runDoIncidentUnitOfMeasureQuantitiesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoIncidentUnitOfMeasureQuantitiesDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/incident-unit-of-measure-quantities/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted incident unit of measure quantity %s\n", opts.ID)
	return nil
}

func parseDoIncidentUnitOfMeasureQuantitiesDeleteOptions(cmd *cobra.Command, args []string) (doIncidentUnitOfMeasureQuantitiesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentUnitOfMeasureQuantitiesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
