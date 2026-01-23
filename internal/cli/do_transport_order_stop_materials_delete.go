package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTransportOrderStopMaterialsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoTransportOrderStopMaterialsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a transport order stop material",
		Long: `Delete a transport order stop material.

Provide the stop material ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a transport order stop material
  xbe do transport-order-stop-materials delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTransportOrderStopMaterialsDelete,
	}
	initDoTransportOrderStopMaterialsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderStopMaterialsCmd.AddCommand(newDoTransportOrderStopMaterialsDeleteCmd())
}

func initDoTransportOrderStopMaterialsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportOrderStopMaterialsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportOrderStopMaterialsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a transport order stop material")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/transport-order-stop-materials/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted transport order stop material %s\n", opts.ID)
	return nil
}

func parseDoTransportOrderStopMaterialsDeleteOptions(cmd *cobra.Command, args []string) (doTransportOrderStopMaterialsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderStopMaterialsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
