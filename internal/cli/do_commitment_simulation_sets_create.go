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

type doCommitmentSimulationSetsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	StartOn          string
	EndOn            string
	IterationCount   int
	OrganizationType string
	OrganizationID   string
}

func newDoCommitmentSimulationSetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a commitment simulation set",
		Long: `Create a commitment simulation set.

Required flags:
  --organization-type  Organization type (brokers, customers, truckers) (required)
  --organization-id    Organization ID (required)
  --start-on           Start date (YYYY-MM-DD) (required)
  --end-on             End date (YYYY-MM-DD) (required)
  --iteration-count    Iteration count (1-1000) (required)`,
		Example: `  # Create a commitment simulation set
  xbe do commitment-simulation-sets create --organization-type brokers --organization-id 123 --start-on 2025-01-01 --end-on 2025-01-07 --iteration-count 10

  # Output as JSON
  xbe do commitment-simulation-sets create --organization-type brokers --organization-id 123 --start-on 2025-01-01 --end-on 2025-01-07 --iteration-count 10 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCommitmentSimulationSetsCreate,
	}
	initDoCommitmentSimulationSetsCreateFlags(cmd)
	return cmd
}

func init() {
	doCommitmentSimulationSetsCmd.AddCommand(newDoCommitmentSimulationSetsCreateCmd())
}

func initDoCommitmentSimulationSetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD) (required)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD) (required)")
	cmd.Flags().Int("iteration-count", 0, "Iteration count (1-1000) (required)")
	cmd.Flags().String("organization-type", "", "Organization type (brokers, customers, truckers) (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCommitmentSimulationSetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCommitmentSimulationSetsCreateOptions(cmd)
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

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.StartOn == "" {
		err := fmt.Errorf("--start-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.EndOn == "" {
		err := fmt.Errorf("--end-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !cmd.Flags().Changed("iteration-count") || opts.IterationCount <= 0 {
		err := fmt.Errorf("--iteration-count is required and must be greater than 0")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"start-on":        opts.StartOn,
		"end-on":          opts.EndOn,
		"iteration-count": opts.IterationCount,
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "commitment-simulation-sets",
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

	body, _, err := client.Post(cmd.Context(), "/v1/commitment-simulation-sets", jsonBody)
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

	row := buildCommitmentSimulationSetRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created commitment simulation set %s\n", row.ID)
	return nil
}

func parseDoCommitmentSimulationSetsCreateOptions(cmd *cobra.Command) (doCommitmentSimulationSetsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	iterationCount, _ := cmd.Flags().GetInt("iteration-count")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCommitmentSimulationSetsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		StartOn:          startOn,
		EndOn:            endOn,
		IterationCount:   iterationCount,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
	}, nil
}

func buildCommitmentSimulationSetRowFromSingle(resp jsonAPISingleResponse) commitmentSimulationSetRow {
	attrs := resp.Data.Attributes

	row := commitmentSimulationSetRow{
		ID:             resp.Data.ID,
		Status:         stringAttr(attrs, "status"),
		StartOn:        formatDate(stringAttr(attrs, "start-on")),
		EndOn:          formatDate(stringAttr(attrs, "end-on")),
		IterationCount: stringAttr(attrs, "iteration-count"),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}
