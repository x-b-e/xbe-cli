package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCommitmentSimulationSetsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
}

func newDoCommitmentSimulationSetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a commitment simulation set",
		Long: `Update a commitment simulation set.

Commitment simulation sets are immutable after creation and do not expose
any writable fields for updates. Delete and recreate a set instead.

Arguments:
  <id>  Commitment simulation set ID (required).`,
		Example: `  # Attempt to update a commitment simulation set (not supported)
  xbe do commitment-simulation-sets update 123`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCommitmentSimulationSetsUpdate,
	}
	initDoCommitmentSimulationSetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCommitmentSimulationSetsCmd.AddCommand(newDoCommitmentSimulationSetsUpdateCmd())
}

func initDoCommitmentSimulationSetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCommitmentSimulationSetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCommitmentSimulationSetsUpdateOptions(cmd, args)
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

	err = fmt.Errorf("commitment simulation sets do not support updates; delete and recreate instead")
	fmt.Fprintln(cmd.ErrOrStderr(), err)
	return err
}

func parseDoCommitmentSimulationSetsUpdateOptions(cmd *cobra.Command, args []string) (doCommitmentSimulationSetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doCommitmentSimulationSetsUpdateOptions{}, fmt.Errorf("commitment simulation set id is required")
	}

	return doCommitmentSimulationSetsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      id,
	}, nil
}
