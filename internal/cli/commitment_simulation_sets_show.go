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

type commitmentSimulationSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type commitmentSimulationSetDetails struct {
	ID                      string   `json:"id"`
	Status                  string   `json:"status,omitempty"`
	StartOn                 string   `json:"start_on,omitempty"`
	EndOn                   string   `json:"end_on,omitempty"`
	IterationCount          string   `json:"iteration_count,omitempty"`
	OrganizationType        string   `json:"organization_type,omitempty"`
	OrganizationID          string   `json:"organization_id,omitempty"`
	CommitmentSimulationIDs []string `json:"commitment_simulation_ids,omitempty"`
}

func newCommitmentSimulationSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show commitment simulation set details",
		Long: `Show the full details of a commitment simulation set.

Output Fields:
  ID                     Commitment simulation set identifier
  STATUS                 Processing status (enqueued or processed)
  START                  Start date
  END                    End date
  ITERATIONS             Iteration count
  ORGANIZATION           Organization type and ID
  COMMITMENT SIMULATIONS Commitment simulation IDs

Arguments:
  <id>  Commitment simulation set ID (required). Find IDs using the list command.`,
		Example: `  # View a commitment simulation set by ID
  xbe view commitment-simulation-sets show 123

  # Get JSON output
  xbe view commitment-simulation-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCommitmentSimulationSetsShow,
	}
	initCommitmentSimulationSetsShowFlags(cmd)
	return cmd
}

func init() {
	commitmentSimulationSetsCmd.AddCommand(newCommitmentSimulationSetsShowCmd())
}

func initCommitmentSimulationSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentSimulationSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCommitmentSimulationSetsShowOptions(cmd)
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
		return fmt.Errorf("commitment simulation set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[commitment-simulation-sets]", "start-on,end-on,iteration-count,status,organization,commitment-simulations")

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-simulation-sets/"+id, query)
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

	details := buildCommitmentSimulationSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCommitmentSimulationSetDetails(cmd, details)
}

func parseCommitmentSimulationSetsShowOptions(cmd *cobra.Command) (commitmentSimulationSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentSimulationSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCommitmentSimulationSetDetails(resp jsonAPISingleResponse) commitmentSimulationSetDetails {
	attrs := resp.Data.Attributes

	details := commitmentSimulationSetDetails{
		ID:             resp.Data.ID,
		Status:         stringAttr(attrs, "status"),
		StartOn:        formatDate(stringAttr(attrs, "start-on")),
		EndOn:          formatDate(stringAttr(attrs, "end-on")),
		IterationCount: stringAttr(attrs, "iteration-count"),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["commitment-simulations"]; ok {
		details.CommitmentSimulationIDs = relationshipIDList(rel)
	}

	return details
}

func renderCommitmentSimulationSetDetails(cmd *cobra.Command, details commitmentSimulationSetDetails) error {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", details.EndOn)
	}
	if details.IterationCount != "" {
		fmt.Fprintf(out, "Iteration Count: %s\n", details.IterationCount)
	}
	if details.OrganizationType != "" || details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization: %s\n", formatPolymorphic(details.OrganizationType, details.OrganizationID))
	}
	if len(details.CommitmentSimulationIDs) > 0 {
		fmt.Fprintf(out, "Commitment Simulations: %s\n", strings.Join(details.CommitmentSimulationIDs, ", "))
	}
	return nil
}
