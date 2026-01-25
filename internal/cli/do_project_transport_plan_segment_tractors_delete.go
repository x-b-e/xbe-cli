package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectTransportPlanSegmentTractorsDeleteOptions struct {
	BaseURL string
	Token   string
	ID      string
	Confirm bool
}

func newDoProjectTransportPlanSegmentTractorsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project transport plan segment tractor",
		Long: `Delete a project transport plan segment tractor.

Provide the segment tractor ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a project transport plan segment tractor
  xbe do project-transport-plan-segment-tractors delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanSegmentTractorsDelete,
	}
	initDoProjectTransportPlanSegmentTractorsDeleteFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanSegmentTractorsCmd.AddCommand(newDoProjectTransportPlanSegmentTractorsDeleteCmd())
}

func initDoProjectTransportPlanSegmentTractorsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanSegmentTractorsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanSegmentTractorsDeleteOptions(cmd, args)
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
		err := fmt.Errorf("--confirm flag is required to delete a project transport plan segment tractor")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Delete(cmd.Context(), "/v1/project-transport-plan-segment-tractors/"+opts.ID)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project transport plan segment tractor %s\n", opts.ID)
	return nil
}

func parseDoProjectTransportPlanSegmentTractorsDeleteOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanSegmentTractorsDeleteOptions, error) {
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanSegmentTractorsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
