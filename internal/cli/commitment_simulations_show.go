package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type commitmentSimulationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type commitmentSimulationDetails struct {
	ID                            string   `json:"id"`
	CommitmentType                string   `json:"commitment_type,omitempty"`
	CommitmentID                  string   `json:"commitment_id,omitempty"`
	StartOn                       string   `json:"start_on,omitempty"`
	EndOn                         string   `json:"end_on,omitempty"`
	IterationCount                int      `json:"iteration_count,omitempty"`
	Status                        string   `json:"status,omitempty"`
	CreatedAt                     string   `json:"created_at,omitempty"`
	UpdatedAt                     string   `json:"updated_at,omitempty"`
	CommitmentSimulationPeriodIDs []string `json:"commitment_simulation_period_ids,omitempty"`
}

func newCommitmentSimulationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show commitment simulation details",
		Long: `Show the full details of a commitment simulation.

Output Fields:
  ID
  Commitment (type and ID)
  Start On
  End On
  Iteration Count
  Status
  Created At
  Updated At
  Commitment Simulation Period IDs

Arguments:
  <id>  The commitment simulation ID (required).`,
		Example: `  # Show a commitment simulation
  xbe view commitment-simulations show 123

  # JSON output
  xbe view commitment-simulations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCommitmentSimulationsShow,
	}
	initCommitmentSimulationsShowFlags(cmd)
	return cmd
}

func init() {
	commitmentSimulationsCmd.AddCommand(newCommitmentSimulationsShowCmd())
}

func initCommitmentSimulationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentSimulationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseCommitmentSimulationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("commitment simulation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-simulations/"+id, nil)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildCommitmentSimulationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCommitmentSimulationDetails(cmd, details)
}

func parseCommitmentSimulationsShowOptions(cmd *cobra.Command) (commitmentSimulationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentSimulationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCommitmentSimulationDetails(resp jsonAPISingleResponse) commitmentSimulationDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := commitmentSimulationDetails{
		ID:             resource.ID,
		StartOn:        formatDate(stringAttr(attrs, "start-on")),
		EndOn:          formatDate(stringAttr(attrs, "end-on")),
		IterationCount: intAttr(attrs, "iteration-count"),
		Status:         stringAttr(attrs, "status"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["commitment"]; ok && rel.Data != nil {
		details.CommitmentType = rel.Data.Type
		details.CommitmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["commitment-simulation-periods"]; ok {
		details.CommitmentSimulationPeriodIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderCommitmentSimulationDetails(cmd *cobra.Command, details commitmentSimulationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CommitmentType != "" && details.CommitmentID != "" {
		fmt.Fprintf(out, "Commitment: %s/%s\n", details.CommitmentType, details.CommitmentID)
	}
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", details.EndOn)
	}
	if details.IterationCount > 0 {
		fmt.Fprintf(out, "Iteration Count: %d\n", details.IterationCount)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if len(details.CommitmentSimulationPeriodIDs) > 0 {
		fmt.Fprintf(out, "Commitment Simulation Period IDs: %s\n", strings.Join(details.CommitmentSimulationPeriodIDs, ", "))
	}

	return nil
}
