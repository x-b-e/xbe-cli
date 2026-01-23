package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type commitmentSimulationPeriodsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type commitmentSimulationPeriodDetails struct {
	ID                     string   `json:"id"`
	Date                   string   `json:"date,omitempty"`
	Window                 string   `json:"window,omitempty"`
	Iterations             string   `json:"iterations,omitempty"`
	Tons                   string   `json:"tons,omitempty"`
	MatchingPeriodIDs      []string `json:"matching_period_ids,omitempty"`
	CommitmentSimulationID string   `json:"commitment_simulation_id,omitempty"`
	CommitmentType         string   `json:"commitment_type,omitempty"`
	CommitmentID           string   `json:"commitment_id,omitempty"`
}

func newCommitmentSimulationPeriodsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show commitment simulation period details",
		Long: `Show the full details of a commitment simulation period.

Output Fields:
  ID                Commitment simulation period identifier
  DATE              Period date
  WINDOW            Period window
  ITERATIONS        Iteration count
  TONS              Tons (customer commitments only)
  MATCHING PERIODS  Matching period IDs
  COMMITMENT SIM    Commitment simulation ID
  COMMITMENT        Commitment type and ID

Arguments:
  <id>  Commitment simulation period ID (required). Find IDs using the list command.`,
		Example: `  # View a commitment simulation period by ID
  xbe view commitment-simulation-periods show 123

  # Get JSON output
  xbe view commitment-simulation-periods show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCommitmentSimulationPeriodsShow,
	}
	initCommitmentSimulationPeriodsShowFlags(cmd)
	return cmd
}

func init() {
	commitmentSimulationPeriodsCmd.AddCommand(newCommitmentSimulationPeriodsShowCmd())
}

func initCommitmentSimulationPeriodsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentSimulationPeriodsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCommitmentSimulationPeriodsShowOptions(cmd)
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
		return fmt.Errorf("commitment simulation period id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[commitment-simulation-periods]", "date,window,iterations,matching-period-ids,tons,commitment-simulation,commitment")

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-simulation-periods/"+id, query)
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

	details := buildCommitmentSimulationPeriodDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCommitmentSimulationPeriodDetails(cmd, details)
}

func parseCommitmentSimulationPeriodsShowOptions(cmd *cobra.Command) (commitmentSimulationPeriodsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentSimulationPeriodsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCommitmentSimulationPeriodDetails(resp jsonAPISingleResponse) commitmentSimulationPeriodDetails {
	attrs := resp.Data.Attributes

	details := commitmentSimulationPeriodDetails{
		ID:                resp.Data.ID,
		Date:              formatDate(stringAttr(attrs, "date")),
		Window:            stringAttr(attrs, "window"),
		Iterations:        stringAttr(attrs, "iterations"),
		Tons:              stringAttr(attrs, "tons"),
		MatchingPeriodIDs: stringSliceAttr(attrs, "matching-period-ids"),
	}

	if rel, ok := resp.Data.Relationships["commitment-simulation"]; ok && rel.Data != nil {
		details.CommitmentSimulationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["commitment"]; ok && rel.Data != nil {
		details.CommitmentType = rel.Data.Type
		details.CommitmentID = rel.Data.ID
	}

	return details
}

func renderCommitmentSimulationPeriodDetails(cmd *cobra.Command, details commitmentSimulationPeriodDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Date != "" {
		fmt.Fprintf(out, "Date: %s\n", details.Date)
	}
	if details.Window != "" {
		fmt.Fprintf(out, "Window: %s\n", details.Window)
	}
	if details.Iterations != "" {
		fmt.Fprintf(out, "Iterations: %s\n", details.Iterations)
	}
	if details.Tons != "" {
		fmt.Fprintf(out, "Tons: %s\n", details.Tons)
	}
	if details.CommitmentSimulationID != "" {
		fmt.Fprintf(out, "Commitment Simulation: %s\n", details.CommitmentSimulationID)
	}
	if details.CommitmentID != "" || details.CommitmentType != "" {
		fmt.Fprintf(out, "Commitment: %s\n", formatPolymorphic(details.CommitmentType, details.CommitmentID))
	}
	if len(details.MatchingPeriodIDs) > 0 {
		fmt.Fprintf(out, "Matching Period IDs: %s\n", strings.Join(details.MatchingPeriodIDs, ", "))
	}

	return nil
}
