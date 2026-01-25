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

type doLineupScenarioGeneratorsCreateOptions struct {
	BaseURL                                   string
	Token                                     string
	JSON                                      bool
	Broker                                    string
	Customer                                  string
	Date                                      string
	Window                                    string
	IncludeTruckerAssignmentsAsConstraints    bool
	TruckerAssignmentLimitsLookbackWindowDays int
	SkipMinimumAssignmentCount                bool
	SkipCreateLineupScenarioSolution          bool
	UseMostRecentLineupScenarioConstraints    bool
}

func newDoLineupScenarioGeneratorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario generator",
		Long: `Create a lineup scenario generator.

Required flags:
  --broker  Broker ID (required)
  --date    Scenario date (YYYY-MM-DD) (required)
  --window  Scenario window (day or night) (required)

Optional flags:
  --customer                                      Customer ID
  --include-trucker-assignments-as-constraints    Include trucker assignments as constraints
  --trucker-assignment-limits-lookback-window-days Lookback window in days for assignment limits
  --skip-minimum-assignment-count                 Skip minimum assignment count constraint
  --skip-create-lineup-scenario-solution          Skip creating lineup scenario solutions
  --use-most-recent-lineup-scenario-constraints   Use most recent scenario constraints`,
		Example: `  # Create a generator for a broker/date/window
  xbe do lineup-scenario-generators create --broker 123 --date 2026-01-23 --window day

  # Create with assignment constraints
  xbe do lineup-scenario-generators create \\
    --broker 123 \\
    --date 2026-01-23 \\
    --window night \\
    --include-trucker-assignments-as-constraints=true \\
    --trucker-assignment-limits-lookback-window-days 7 \\
    --skip-minimum-assignment-count=true \\
    --skip-create-lineup-scenario-solution=true \\
    --use-most-recent-lineup-scenario-constraints=false

  # JSON output
  xbe do lineup-scenario-generators create --broker 123 --date 2026-01-23 --window day --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupScenarioGeneratorsCreate,
	}
	initDoLineupScenarioGeneratorsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioGeneratorsCmd.AddCommand(newDoLineupScenarioGeneratorsCreateCmd())
}

func initDoLineupScenarioGeneratorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("date", "", "Scenario date (YYYY-MM-DD) (required)")
	cmd.Flags().String("window", "", "Scenario window (day or night) (required)")
	cmd.Flags().Bool("include-trucker-assignments-as-constraints", false, "Include trucker assignments as constraints")
	cmd.Flags().Int("trucker-assignment-limits-lookback-window-days", 0, "Lookback window in days for assignment limits")
	cmd.Flags().Bool("skip-minimum-assignment-count", false, "Skip minimum assignment count constraint")
	cmd.Flags().Bool("skip-create-lineup-scenario-solution", false, "Skip creating lineup scenario solutions")
	cmd.Flags().Bool("use-most-recent-lineup-scenario-constraints", false, "Use most recent scenario constraints")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioGeneratorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenarioGeneratorsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Broker) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Date) == "" {
		err := fmt.Errorf("--date is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Window) == "" {
		err := fmt.Errorf("--window is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"date":   opts.Date,
		"window": opts.Window,
	}

	if cmd.Flags().Changed("include-trucker-assignments-as-constraints") {
		attributes["include-trucker-assignments-as-constraints"] = opts.IncludeTruckerAssignmentsAsConstraints
	}
	if cmd.Flags().Changed("trucker-assignment-limits-lookback-window-days") {
		attributes["trucker-assignment-limits-lookback-window-days"] = opts.TruckerAssignmentLimitsLookbackWindowDays
	}
	if cmd.Flags().Changed("skip-minimum-assignment-count") {
		attributes["skip-minimum-assignment-count"] = opts.SkipMinimumAssignmentCount
	}
	if cmd.Flags().Changed("skip-create-lineup-scenario-solution") {
		attributes["skip-create-lineup-scenario-solution"] = opts.SkipCreateLineupScenarioSolution
	}
	if cmd.Flags().Changed("use-most-recent-lineup-scenario-constraints") {
		attributes["use-most-recent-lineup-scenario-constraints"] = opts.UseMostRecentLineupScenarioConstraints
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}
	if strings.TrimSpace(opts.Customer) != "" {
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-scenario-generators",
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

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenario-generators", jsonBody)
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

	row := buildLineupScenarioGeneratorRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario generator %s\n", row.ID)
	return nil
}

func parseDoLineupScenarioGeneratorsCreateOptions(cmd *cobra.Command) (doLineupScenarioGeneratorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	date, _ := cmd.Flags().GetString("date")
	window, _ := cmd.Flags().GetString("window")
	includeAssignments, _ := cmd.Flags().GetBool("include-trucker-assignments-as-constraints")
	lookbackWindowDays, _ := cmd.Flags().GetInt("trucker-assignment-limits-lookback-window-days")
	skipMinimumAssignmentCount, _ := cmd.Flags().GetBool("skip-minimum-assignment-count")
	skipCreateSolution, _ := cmd.Flags().GetBool("skip-create-lineup-scenario-solution")
	useMostRecentConstraints, _ := cmd.Flags().GetBool("use-most-recent-lineup-scenario-constraints")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioGeneratorsCreateOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		Broker:                                 broker,
		Customer:                               customer,
		Date:                                   date,
		Window:                                 window,
		IncludeTruckerAssignmentsAsConstraints: includeAssignments,
		TruckerAssignmentLimitsLookbackWindowDays: lookbackWindowDays,
		SkipMinimumAssignmentCount:                skipMinimumAssignmentCount,
		SkipCreateLineupScenarioSolution:          skipCreateSolution,
		UseMostRecentLineupScenarioConstraints:    useMostRecentConstraints,
	}, nil
}
