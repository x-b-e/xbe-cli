package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanAlarmsUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	Tons                               string
	BaseMaterialTypeFullyQualifiedName string
	MaxLatencyMinutes                  string
	Note                               string
}

func newDoJobProductionPlanAlarmsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan alarm",
		Long: `Update a job production plan alarm.

Arguments:
  <id>    The alarm ID (required).

Optional flags:
  --tons                                  Tonnage trigger (must be > 0)
  --base-material-type-fully-qualified-name  Base material type filter
  --max-latency-minutes                   Max latency in minutes (must be > 0)
  --note                                  Alarm note`,
		Example: `  # Update tonnage trigger
  xbe do job-production-plan-alarms update 123 --tons 200

  # Update base material type and note
  xbe do job-production-plan-alarms update 123 \
    --base-material-type-fully-qualified-name "Asphalt Mixture" \
    --note "Updated alarm note"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanAlarmsUpdate,
	}
	initDoJobProductionPlanAlarmsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanAlarmsCmd.AddCommand(newDoJobProductionPlanAlarmsUpdateCmd())
}

func initDoJobProductionPlanAlarmsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tons", "", "Tonnage trigger")
	cmd.Flags().String("base-material-type-fully-qualified-name", "", "Base material type fully qualified name")
	cmd.Flags().String("max-latency-minutes", "", "Max latency in minutes")
	cmd.Flags().String("note", "", "Alarm note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanAlarmsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanAlarmsUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan alarm id is required")
	}

	attributes := map[string]any{}

	if cmd.Flags().Changed("tons") {
		if strings.TrimSpace(opts.Tons) == "" {
			err := fmt.Errorf("--tons cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		tonsValue, err := strconv.ParseFloat(opts.Tons, 64)
		if err != nil || tonsValue <= 0 {
			err := fmt.Errorf("--tons must be a number greater than 0")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["tons"] = tonsValue
	}
	if cmd.Flags().Changed("base-material-type-fully-qualified-name") {
		attributes["base-material-type-fully-qualified-name"] = opts.BaseMaterialTypeFullyQualifiedName
	}
	if cmd.Flags().Changed("max-latency-minutes") {
		if strings.TrimSpace(opts.MaxLatencyMinutes) == "" {
			err := fmt.Errorf("--max-latency-minutes cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		maxLatencyValue, err := strconv.Atoi(opts.MaxLatencyMinutes)
		if err != nil || maxLatencyValue <= 0 {
			err := fmt.Errorf("--max-latency-minutes must be an integer greater than 0")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["max-latency-minutes"] = maxLatencyValue
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields provided to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-production-plan-alarms",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-alarms/"+id, jsonBody)
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

	row := buildJobProductionPlanAlarmRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan alarm %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanAlarmsUpdateOptions(cmd *cobra.Command) (doJobProductionPlanAlarmsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tns, _ := cmd.Flags().GetString("tons")
	baseMaterialType, _ := cmd.Flags().GetString("base-material-type-fully-qualified-name")
	maxLatencyMinutes, _ := cmd.Flags().GetString("max-latency-minutes")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanAlarmsUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Tons:                               tns,
		BaseMaterialTypeFullyQualifiedName: baseMaterialType,
		MaxLatencyMinutes:                  maxLatencyMinutes,
		Note:                               note,
	}, nil
}
