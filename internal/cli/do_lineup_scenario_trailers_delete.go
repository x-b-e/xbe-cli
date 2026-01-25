package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doLineupScenarioTrailersDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Force   bool
}

func newDoLineupScenarioTrailersDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a lineup scenario trailer",
		Long: `Delete a lineup scenario trailer.

This action is permanent and cannot be undone.

Required:
  --confirm   Confirm deletion`,
		Example: `  # Delete a lineup scenario trailer
  xbe do lineup-scenario-trailers delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupScenarioTrailersDelete,
	}
	initDoLineupScenarioTrailersDeleteFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioTrailersCmd.AddCommand(newDoLineupScenarioTrailersDeleteCmd())
}

func initDoLineupScenarioTrailersDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioTrailersDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupScenarioTrailersDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Force {
		err := fmt.Errorf("--confirm is required to delete a lineup scenario trailer")
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

	body, _, err := client.Delete(cmd.Context(), "/v1/lineup-scenario-trailers/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted lineup scenario trailer %s\n", opts.ID)
	return nil
}

func parseDoLineupScenarioTrailersDeleteOptions(cmd *cobra.Command, args []string) (doLineupScenarioTrailersDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioTrailersDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Force:   confirm,
	}, nil
}
