package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCommitmentSimulationsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	CommitmentType string
	CommitmentID   string
	StartOn        string
	EndOn          string
	IterationCount int
}

func newDoCommitmentSimulationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a commitment simulation",
		Long: `Create a commitment simulation.

Required flags:
  --commitment-type  Commitment type (e.g., commitments, customer-commitments) (required)
  --commitment-id    Commitment ID (required)
  --start-on         Simulation start date (YYYY-MM-DD) (required)
  --end-on           Simulation end date (YYYY-MM-DD) (required)
  --iteration-count  Iteration count (1-1000) (required)`,
		Example: `  # Create a commitment simulation
  xbe do commitment-simulations create \
    --commitment-type commitments \
    --commitment-id 123 \
    --start-on 2026-01-23 \
    --end-on 2026-01-23 \
    --iteration-count 100

  # JSON output
  xbe do commitment-simulations create \
    --commitment-type commitments \
    --commitment-id 123 \
    --start-on 2026-01-23 \
    --end-on 2026-01-23 \
    --iteration-count 100 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoCommitmentSimulationsCreate,
	}
	initDoCommitmentSimulationsCreateFlags(cmd)
	return cmd
}

func init() {
	doCommitmentSimulationsCmd.AddCommand(newDoCommitmentSimulationsCreateCmd())
}

func initDoCommitmentSimulationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("commitment-type", "", "Commitment type (required)")
	cmd.Flags().String("commitment-id", "", "Commitment ID (required)")
	cmd.Flags().String("start-on", "", "Simulation start date (YYYY-MM-DD) (required)")
	cmd.Flags().String("end-on", "", "Simulation end date (YYYY-MM-DD) (required)")
	cmd.Flags().Int("iteration-count", 0, "Iteration count (1-1000) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCommitmentSimulationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCommitmentSimulationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.CommitmentType) == "" {
		err := fmt.Errorf("--commitment-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.CommitmentID) == "" {
		err := fmt.Errorf("--commitment-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.StartOn) == "" {
		err := fmt.Errorf("--start-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.EndOn) == "" {
		err := fmt.Errorf("--end-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.IterationCount <= 0 {
		err := fmt.Errorf("--iteration-count is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.IterationCount > 1000 {
		err := fmt.Errorf("--iteration-count must be 1000 or less")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if startOn, err := time.Parse("2006-01-02", opts.StartOn); err == nil {
		if endOn, err := time.Parse("2006-01-02", opts.EndOn); err == nil {
			if endOn.Before(startOn) {
				err := fmt.Errorf("--end-on must be on or after --start-on")
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
		}
	}

	attributes := map[string]any{
		"start-on":        opts.StartOn,
		"end-on":          opts.EndOn,
		"iteration-count": opts.IterationCount,
	}

	relationships := map[string]any{
		"commitment": map[string]any{
			"data": map[string]any{
				"type": opts.CommitmentType,
				"id":   opts.CommitmentID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "commitment-simulations",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/commitment-simulations", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := buildCommitmentSimulationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Status != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created commitment simulation %s (%s)\n", row.ID, row.Status)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created commitment simulation %s\n", row.ID)
	return nil
}

func parseDoCommitmentSimulationsCreateOptions(cmd *cobra.Command) (doCommitmentSimulationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	commitmentType, _ := cmd.Flags().GetString("commitment-type")
	commitmentID, _ := cmd.Flags().GetString("commitment-id")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	iterationCount, _ := cmd.Flags().GetInt("iteration-count")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCommitmentSimulationsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		CommitmentType: commitmentType,
		CommitmentID:   commitmentID,
		StartOn:        startOn,
		EndOn:          endOn,
		IterationCount: iterationCount,
	}, nil
}
