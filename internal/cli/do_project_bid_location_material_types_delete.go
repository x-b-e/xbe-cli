package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectBidLocationMaterialTypesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoProjectBidLocationMaterialTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project bid location material type",
		Long: `Delete a project bid location material type.

Required flags:
  --confirm    Confirm deletion`,
		Example: `  # Delete a project bid location material type
  xbe do project-bid-location-material-types delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectBidLocationMaterialTypesDelete,
	}
	initDoProjectBidLocationMaterialTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doProjectBidLocationMaterialTypesCmd.AddCommand(newDoProjectBidLocationMaterialTypesDeleteCmd())
}

func initDoProjectBidLocationMaterialTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectBidLocationMaterialTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectBidLocationMaterialTypesDeleteOptions(cmd, args)
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

	body, _, err := client.Delete(cmd.Context(), "/v1/project-bid-location-material-types/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project bid location material type %s\n", opts.ID)
	return nil
}

func parseDoProjectBidLocationMaterialTypesDeleteOptions(cmd *cobra.Command, args []string) (doProjectBidLocationMaterialTypesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectBidLocationMaterialTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
