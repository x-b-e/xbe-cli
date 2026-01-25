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

type doJobProductionPlanTruckingIncidentDetectorsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	JobProductionPlanID string
	AsOf                string
	PersistChanges      bool
}

func newDoJobProductionPlanTruckingIncidentDetectorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan trucking incident detector",
		Long: `Create a job production plan trucking incident detector.

Required:
  --job-production-plan  Job production plan ID

Optional:
  --as-of            Detect incidents as of a timestamp (RFC3339)
  --persist-changes  Persist detected incident changes`,
		Example: `  # Run detector for a job production plan
  xbe do job-production-plan-trucking-incident-detectors create --job-production-plan 123

  # Run detector as of a timestamp
  xbe do job-production-plan-trucking-incident-detectors create \
    --job-production-plan 123 \
    --as-of "2026-01-23T00:00:00Z"

  # Persist detected incident changes
  xbe do job-production-plan-trucking-incident-detectors create \
    --job-production-plan 123 \
    --persist-changes`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanTruckingIncidentDetectorsCreate,
	}
	initDoJobProductionPlanTruckingIncidentDetectorsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanTruckingIncidentDetectorsCmd.AddCommand(newDoJobProductionPlanTruckingIncidentDetectorsCreateCmd())
}

func initDoJobProductionPlanTruckingIncidentDetectorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("as-of", "", "Detect incidents as of timestamp (RFC3339)")
	cmd.Flags().Bool("persist-changes", false, "Persist detected incident changes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanTruckingIncidentDetectorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanTruckingIncidentDetectorsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if strings.TrimSpace(opts.AsOf) != "" {
		attributes["as-of"] = opts.AsOf
	}
	if cmd.Flags().Changed("persist-changes") {
		attributes["persist-changes"] = opts.PersistChanges
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
	}

	data := map[string]any{
		"type":          "job-production-plan-trucking-incident-detectors",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-trucking-incident-detectors", jsonBody)
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

	row := buildJobProductionPlanTruckingIncidentDetectorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan trucking incident detector %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanTruckingIncidentDetectorsCreateOptions(cmd *cobra.Command) (doJobProductionPlanTruckingIncidentDetectorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	asOf, _ := cmd.Flags().GetString("as-of")
	persistChanges, _ := cmd.Flags().GetBool("persist-changes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanTruckingIncidentDetectorsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		JobProductionPlanID: jobProductionPlanID,
		AsOf:                asOf,
		PersistChanges:      persistChanges,
	}, nil
}

func buildJobProductionPlanTruckingIncidentDetectorRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanTruckingIncidentDetectorRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := jobProductionPlanTruckingIncidentDetectorRow{
		ID:                   resource.ID,
		AsOf:                 formatDateTime(stringAttr(attrs, "as-of")),
		PersistChanges:       boolAttr(attrs, "persist-changes"),
		IsPerformed:          boolAttr(attrs, "is-performed"),
		DetectedIncidentsCnt: sliceLenAttr(attrs, "detected-incidents"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}
