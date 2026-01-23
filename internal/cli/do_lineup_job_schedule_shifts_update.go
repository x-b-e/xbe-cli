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

type doLineupJobScheduleShiftsUpdateOptions struct {
	BaseURL                             string
	Token                               string
	JSON                                bool
	ID                                  string
	Trucker                             string
	Driver                              string
	TrailerClassification               string
	TrailerClassificationEquivalentType string
	IsBrokered                          string
	IsReadyToDispatch                   string
	ExcludeFromLineupScenarios          string
	TravelMinutes                       string
	LoadedTonsMax                       string
	ExplicitMaterialTransactionTonsMax  string
	NotifyDriverOnLateShiftAssignment   string
	IsExpectingTimeCard                 string
}

func newDoLineupJobScheduleShiftsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a lineup job schedule shift",
		Long: `Update a lineup job schedule shift.

Optional attributes:
  --trailer-classification-equivalent-type  Trailer classification equivalent type (tendered, assigned)
  --is-brokered                             Update brokered flag (true/false)
  --is-ready-to-dispatch                    Update ready to dispatch (true/false)
  --exclude-from-lineup-scenarios           Update exclude from lineup scenarios (true/false)
  --travel-minutes                          Update travel minutes (deprecated)
  --loaded-tons-max                         Update loaded tons max (deprecated alias)
  --explicit-material-transaction-tons-max  Update explicit material transaction tons max
  --notify-driver-on-late-shift-assignment  Update notify driver on late shift assignment (true/false)
  --is-expecting-time-card                  Update expecting time card (true/false)

Optional relationships:
  --trucker                 Trucker ID (empty clears)
  --driver                  Driver (user) ID (empty clears)
  --trailer-classification  Trailer classification ID (empty clears)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update readiness flags
  xbe do lineup-job-schedule-shifts update 123 --is-ready-to-dispatch true

  # Update trailer classification equivalent type
  xbe do lineup-job-schedule-shifts update 123 --trailer-classification-equivalent-type assigned`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupJobScheduleShiftsUpdate,
	}
	initDoLineupJobScheduleShiftsUpdateFlags(cmd)
	return cmd
}

func init() {
	doLineupJobScheduleShiftsCmd.AddCommand(newDoLineupJobScheduleShiftsUpdateCmd())
}

func initDoLineupJobScheduleShiftsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID (empty clears)")
	cmd.Flags().String("driver", "", "Driver (user) ID (empty clears)")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID (empty clears)")
	cmd.Flags().String("trailer-classification-equivalent-type", "", "Trailer classification equivalent type (tendered, assigned)")
	cmd.Flags().String("is-brokered", "", "Update brokered flag (true/false)")
	cmd.Flags().String("is-ready-to-dispatch", "", "Update ready to dispatch (true/false)")
	cmd.Flags().String("exclude-from-lineup-scenarios", "", "Update exclude from lineup scenarios (true/false)")
	cmd.Flags().String("travel-minutes", "", "Update travel minutes (deprecated)")
	cmd.Flags().String("loaded-tons-max", "", "Update loaded tons max (deprecated alias)")
	cmd.Flags().String("explicit-material-transaction-tons-max", "", "Update explicit material transaction tons max")
	cmd.Flags().String("notify-driver-on-late-shift-assignment", "", "Update notify driver on late shift assignment (true/false)")
	cmd.Flags().String("is-expecting-time-card", "", "Update expecting time card (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupJobScheduleShiftsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupJobScheduleShiftsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("trailer-classification-equivalent-type") {
		attributes["trailer-classification-equivalent-type"] = opts.TrailerClassificationEquivalentType
	}
	if cmd.Flags().Changed("is-brokered") {
		attributes["is-brokered"] = opts.IsBrokered == "true"
	}
	if cmd.Flags().Changed("is-ready-to-dispatch") {
		attributes["is-ready-to-dispatch"] = opts.IsReadyToDispatch == "true"
	}
	if cmd.Flags().Changed("exclude-from-lineup-scenarios") {
		attributes["exclude-from-lineup-scenarios"] = opts.ExcludeFromLineupScenarios == "true"
	}
	if cmd.Flags().Changed("travel-minutes") {
		attributes["travel-minutes"] = opts.TravelMinutes
	}
	if cmd.Flags().Changed("loaded-tons-max") {
		attributes["loaded-tons-max"] = opts.LoadedTonsMax
	}
	if cmd.Flags().Changed("explicit-material-transaction-tons-max") {
		attributes["explicit-material-transaction-tons-max"] = opts.ExplicitMaterialTransactionTonsMax
	}
	if cmd.Flags().Changed("notify-driver-on-late-shift-assignment") {
		attributes["notify-driver-on-late-shift-assignment"] = opts.NotifyDriverOnLateShiftAssignment == "true"
	}
	if cmd.Flags().Changed("is-expecting-time-card") {
		attributes["is-expecting-time-card"] = opts.IsExpectingTimeCard == "true"
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("trucker") {
		if opts.Trucker == "" {
			relationships["trucker"] = map[string]any{"data": nil}
		} else {
			relationships["trucker"] = map[string]any{
				"data": map[string]string{
					"type": "truckers",
					"id":   opts.Trucker,
				},
			}
		}
	}
	if cmd.Flags().Changed("driver") {
		if opts.Driver == "" {
			relationships["driver"] = map[string]any{"data": nil}
		} else {
			relationships["driver"] = map[string]any{
				"data": map[string]string{
					"type": "users",
					"id":   opts.Driver,
				},
			}
		}
	}
	if cmd.Flags().Changed("trailer-classification") {
		if opts.TrailerClassification == "" {
			relationships["trailer-classification"] = map[string]any{"data": nil}
		} else {
			relationships["trailer-classification"] = map[string]any{
				"data": map[string]string{
					"type": "trailer-classifications",
					"id":   opts.TrailerClassification,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestData := map[string]any{
		"type": "lineup-job-schedule-shifts",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		requestData["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestData["relationships"] = relationships
	}

	requestBody := map[string]any{"data": requestData}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)
	body, _, err := client.Patch(cmd.Context(), "/v1/lineup-job-schedule-shifts/"+opts.ID, jsonBody)
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

	result := buildLineupJobScheduleShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated lineup job schedule shift %s\n", result.ID)
	return nil
}

func parseDoLineupJobScheduleShiftsUpdateOptions(cmd *cobra.Command, args []string) (doLineupJobScheduleShiftsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	driver, _ := cmd.Flags().GetString("driver")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	trailerClassificationEquivalentType, _ := cmd.Flags().GetString("trailer-classification-equivalent-type")
	isBrokered, _ := cmd.Flags().GetString("is-brokered")
	isReadyToDispatch, _ := cmd.Flags().GetString("is-ready-to-dispatch")
	excludeFromLineupScenarios, _ := cmd.Flags().GetString("exclude-from-lineup-scenarios")
	travelMinutes, _ := cmd.Flags().GetString("travel-minutes")
	loadedTonsMax, _ := cmd.Flags().GetString("loaded-tons-max")
	explicitMaterialTransactionTonsMax, _ := cmd.Flags().GetString("explicit-material-transaction-tons-max")
	notifyDriverOnLateShiftAssignment, _ := cmd.Flags().GetString("notify-driver-on-late-shift-assignment")
	isExpectingTimeCard, _ := cmd.Flags().GetString("is-expecting-time-card")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupJobScheduleShiftsUpdateOptions{
		BaseURL:                             baseURL,
		Token:                               token,
		JSON:                                jsonOut,
		ID:                                  args[0],
		Trucker:                             trucker,
		Driver:                              driver,
		TrailerClassification:               trailerClassification,
		TrailerClassificationEquivalentType: trailerClassificationEquivalentType,
		IsBrokered:                          isBrokered,
		IsReadyToDispatch:                   isReadyToDispatch,
		ExcludeFromLineupScenarios:          excludeFromLineupScenarios,
		TravelMinutes:                       travelMinutes,
		LoadedTonsMax:                       loadedTonsMax,
		ExplicitMaterialTransactionTonsMax:  explicitMaterialTransactionTonsMax,
		NotifyDriverOnLateShiftAssignment:   notifyDriverOnLateShiftAssignment,
		IsExpectingTimeCard:                 isExpectingTimeCard,
	}, nil
}
