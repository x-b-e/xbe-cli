package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialSiteInventoryLocationsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoMaterialSiteInventoryLocationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a material site inventory location",
		Long: `Delete a material site inventory location.

Required flags:
  --confirm    Confirm deletion

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Delete a material site inventory location
  xbe do material-site-inventory-locations delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialSiteInventoryLocationsDelete,
	}
	initDoMaterialSiteInventoryLocationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteInventoryLocationsCmd.AddCommand(newDoMaterialSiteInventoryLocationsDeleteCmd())
}

func initDoMaterialSiteInventoryLocationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("confirm")
}

func runDoMaterialSiteInventoryLocationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSiteInventoryLocationsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/material-site-inventory-locations/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted material site inventory location %s\n", opts.ID)
	return nil
}

func parseDoMaterialSiteInventoryLocationsDeleteOptions(cmd *cobra.Command, args []string) (doMaterialSiteInventoryLocationsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteInventoryLocationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
