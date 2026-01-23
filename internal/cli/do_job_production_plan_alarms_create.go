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

type doJobProductionPlanAlarmsCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	JobProductionPlanID                string
	Tons                               string
	BaseMaterialTypeFullyQualifiedName string
	MaxLatencyMinutes                  string
	Note                               string
}

func newDoJobProductionPlanAlarmsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan alarm",
		Long: `Create a job production plan alarm.

Required flags:
  --job-production-plan  Job production plan ID (required)
  --tons                 Tonnage trigger (required, must be > 0)

Optional flags:
  --base-material-type-fully-qualified-name  Base material type filter
  --max-latency-minutes                      Max latency in minutes (must be > 0)
  --note                                     Alarm note`,
		Example: `  # Create an alarm with a tonnage trigger
  xbe do job-production-plan-alarms create \
    --job-production-plan 123 \
    --tons 150

  # Create with base material type and latency
  xbe do job-production-plan-alarms create \
    --job-production-plan 123 \
    --tons 200 \
    --base-material-type-fully-qualified-name "Asphalt Mixture" \
    --max-latency-minutes 45 \
    --note "Alert at 200 tons"

  # Get JSON output
  xbe do job-production-plan-alarms create \
    --job-production-plan 123 \
    --tons 150 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanAlarmsCreate,
	}
	initDoJobProductionPlanAlarmsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanAlarmsCmd.AddCommand(newDoJobProductionPlanAlarmsCreateCmd())
}

func initDoJobProductionPlanAlarmsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("tons", "", "Tonnage trigger (required)")
	cmd.Flags().String("base-material-type-fully-qualified-name", "", "Base material type fully qualified name")
	cmd.Flags().String("max-latency-minutes", "", "Max latency in minutes")
	cmd.Flags().String("note", "", "Alarm note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanAlarmsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanAlarmsCreateOptions(cmd)
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

	if opts.JobProductionPlanID == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Tons == "" {
		err := fmt.Errorf("--tons is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	tonsValue, err := strconv.ParseFloat(opts.Tons, 64)
	if err != nil || tonsValue <= 0 {
		err := fmt.Errorf("--tons must be a number greater than 0")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"tons": tonsValue,
	}

	if opts.BaseMaterialTypeFullyQualifiedName != "" {
		attributes["base-material-type-fully-qualified-name"] = opts.BaseMaterialTypeFullyQualifiedName
	}
	if opts.MaxLatencyMinutes != "" {
		maxLatencyValue, err := strconv.Atoi(opts.MaxLatencyMinutes)
		if err != nil || maxLatencyValue <= 0 {
			err := fmt.Errorf("--max-latency-minutes must be an integer greater than 0")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["max-latency-minutes"] = maxLatencyValue
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-alarms",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-alarms", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan alarm %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanAlarmsCreateOptions(cmd *cobra.Command) (doJobProductionPlanAlarmsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	tns, _ := cmd.Flags().GetString("tons")
	baseMaterialType, _ := cmd.Flags().GetString("base-material-type-fully-qualified-name")
	maxLatencyMinutes, _ := cmd.Flags().GetString("max-latency-minutes")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanAlarmsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		JobProductionPlanID:                jobProductionPlanID,
		Tons:                               tns,
		BaseMaterialTypeFullyQualifiedName: baseMaterialType,
		MaxLatencyMinutes:                  maxLatencyMinutes,
		Note:                               note,
	}, nil
}

func buildJobProductionPlanAlarmRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanAlarmRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := jobProductionPlanAlarmRow{
		ID:                                 resource.ID,
		Tons:                               floatAttr(attrs, "tons"),
		BaseMaterialTypeFullyQualifiedName: stringAttr(attrs, "base-material-type-fully-qualified-name"),
		MaxLatencyMinutes:                  intAttr(attrs, "max-latency-minutes"),
		PlannedAt:                          formatDateTime(stringAttr(attrs, "planned-at")),
		FulfilledAt:                        formatDateTime(stringAttr(attrs, "fulfilled-at")),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}
