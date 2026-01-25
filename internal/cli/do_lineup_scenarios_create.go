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

type doLineupScenariosCreateOptions struct {
	BaseURL                                string
	Token                                  string
	JSON                                   bool
	Broker                                 string
	Customer                               string
	Name                                   string
	Date                                   string
	Window                                 string
	IncludeTruckerAssignmentsAsConstraints bool
	AddLineupsAutomatically                bool
}

func newDoLineupScenariosCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario",
		Long: `Create a lineup scenario.

Required flags:
  --broker  Broker ID (required)
  --date    Scenario date (YYYY-MM-DD) (required)
  --window  Scenario window (day or night) (required)

Optional flags:
  --customer                                   Customer ID
  --name                                       Scenario name
  --include-trucker-assignments-as-constraints Include trucker assignments as constraints
  --add-lineups-automatically                  Automatically add lineups

Examples:
  # Create a lineup scenario for a broker/date/window
  xbe do lineup-scenarios create --broker 123 --date 2026-01-23 --window day

  # Create with optional attributes
  xbe do lineup-scenarios create \
    --broker 123 \
    --date 2026-01-23 \
    --window night \
    --name "Night scenario" \
    --include-trucker-assignments-as-constraints=true \
    --add-lineups-automatically=false

  # JSON output
  xbe do lineup-scenarios create --broker 123 --date 2026-01-23 --window day --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupScenariosCreate,
	}
	initDoLineupScenariosCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenariosCmd.AddCommand(newDoLineupScenariosCreateCmd())
}

func initDoLineupScenariosCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("name", "", "Scenario name")
	cmd.Flags().String("date", "", "Scenario date (YYYY-MM-DD) (required)")
	cmd.Flags().String("window", "", "Scenario window (day or night) (required)")
	cmd.Flags().Bool("include-trucker-assignments-as-constraints", false, "Include trucker assignments as constraints")
	cmd.Flags().Bool("add-lineups-automatically", false, "Automatically add lineups")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenariosCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenariosCreateOptions(cmd)
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
	if strings.TrimSpace(opts.Name) != "" {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("include-trucker-assignments-as-constraints") {
		attributes["include-trucker-assignments-as-constraints"] = opts.IncludeTruckerAssignmentsAsConstraints
	}
	if cmd.Flags().Changed("add-lineups-automatically") {
		attributes["add-lineups-automatically"] = opts.AddLineupsAutomatically
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
			"type":          "lineup-scenarios",
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

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenarios", jsonBody)
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

	row := buildLineupScenarioRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario %s\n", row.ID)
	return nil
}

func parseDoLineupScenariosCreateOptions(cmd *cobra.Command) (doLineupScenariosCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	name, _ := cmd.Flags().GetString("name")
	date, _ := cmd.Flags().GetString("date")
	window, _ := cmd.Flags().GetString("window")
	includeAssignments, _ := cmd.Flags().GetBool("include-trucker-assignments-as-constraints")
	addLineupsAutomatically, _ := cmd.Flags().GetBool("add-lineups-automatically")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenariosCreateOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		Broker:                                 broker,
		Customer:                               customer,
		Name:                                   name,
		Date:                                   date,
		Window:                                 window,
		IncludeTruckerAssignmentsAsConstraints: includeAssignments,
		AddLineupsAutomatically:                addLineupsAutomatically,
	}, nil
}
