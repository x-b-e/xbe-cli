package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doDeveloperReferenceTypesDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoDeveloperReferenceTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a developer reference type",
		Long: `Delete a developer reference type.

Provide the reference type ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a developer reference type
  xbe do developer-reference-types delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperReferenceTypesDelete,
	}
	initDoDeveloperReferenceTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doDeveloperReferenceTypesCmd.AddCommand(newDoDeveloperReferenceTypesDeleteCmd())
}

func initDoDeveloperReferenceTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperReferenceTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperReferenceTypesDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a developer reference type")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/developer-reference-types/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted developer reference type %s\n", opts.ID)
	return nil
}

func parseDoDeveloperReferenceTypesDeleteOptions(cmd *cobra.Command, args []string) (doDeveloperReferenceTypesDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperReferenceTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
