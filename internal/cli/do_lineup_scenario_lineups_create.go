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

type doLineupScenarioLineupsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	LineupScenario string
	Lineup         string
}

func newDoLineupScenarioLineupsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario lineup",
		Long: `Create a lineup scenario lineup.

Required flags:
  --lineup-scenario  Lineup scenario ID (required)
  --lineup           Lineup ID (required)

Note: The lineup must match the lineup scenario broker and date window.`,
		Example: `  # Link a lineup to a scenario
  xbe do lineup-scenario-lineups create --lineup-scenario 123 --lineup 456

  # Output as JSON
  xbe do lineup-scenario-lineups create --lineup-scenario 123 --lineup 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupScenarioLineupsCreate,
	}
	initDoLineupScenarioLineupsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioLineupsCmd.AddCommand(newDoLineupScenarioLineupsCreateCmd())
}

func initDoLineupScenarioLineupsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-scenario", "", "Lineup scenario ID (required)")
	cmd.Flags().String("lineup", "", "Lineup ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioLineupsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenarioLineupsCreateOptions(cmd)
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

	if opts.LineupScenario == "" {
		err := fmt.Errorf("--lineup-scenario is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Lineup == "" {
		err := fmt.Errorf("--lineup is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"lineup-scenario": map[string]any{
			"data": map[string]any{
				"type": "lineup-scenarios",
				"id":   opts.LineupScenario,
			},
		},
		"lineup": map[string]any{
			"data": map[string]any{
				"type": "lineups",
				"id":   opts.Lineup,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-scenario-lineups",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenario-lineups", jsonBody)
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

	row := lineupScenarioLineupRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario lineup %s\n", row.ID)
	return nil
}

func parseDoLineupScenarioLineupsCreateOptions(cmd *cobra.Command) (doLineupScenarioLineupsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupScenario, _ := cmd.Flags().GetString("lineup-scenario")
	lineup, _ := cmd.Flags().GetString("lineup")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioLineupsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		LineupScenario: lineupScenario,
		Lineup:         lineup,
	}, nil
}
