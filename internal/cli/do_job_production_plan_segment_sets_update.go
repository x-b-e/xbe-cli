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

type doJobProductionPlanSegmentSetsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	JobProductionPlanID string
	Name                string
	IsDefault           bool
	StartOffsetMinutes  int
}

func newDoJobProductionPlanSegmentSetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan segment set",
		Long: `Update a job production plan segment set.

Arguments:
  <id>    The segment set ID (required).

Optional flags:
  --job-production-plan  Job production plan ID
  --name                 Segment set name
  --is-default           Mark the set as default
  --start-offset-minutes Start offset in minutes`,
		Example: `  # Update the segment set name
  xbe do job-production-plan-segment-sets update 123 --name "PM shift"

  # Update offset and default flag
  xbe do job-production-plan-segment-sets update 123 --start-offset-minutes 30 --is-default`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanSegmentSetsUpdate,
	}
	initDoJobProductionPlanSegmentSetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanSegmentSetsCmd.AddCommand(newDoJobProductionPlanSegmentSetsUpdateCmd())
}

func initDoJobProductionPlanSegmentSetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("name", "", "Segment set name")
	cmd.Flags().Bool("is-default", false, "Mark the set as default")
	cmd.Flags().Int("start-offset-minutes", 0, "Start offset in minutes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanSegmentSetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanSegmentSetsUpdateOptions(cmd, args)
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

	data := map[string]any{
		"type": "job-production-plan-segment-sets",
		"id":   opts.ID,
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("is-default") {
		attributes["is-default"] = opts.IsDefault
	}
	if cmd.Flags().Changed("start-offset-minutes") {
		attributes["start-offset-minutes"] = opts.StartOffsetMinutes
	}

	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("job-production-plan") {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		}
	}

	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-segment-sets/"+opts.ID, jsonBody)
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

	row := buildJobProductionPlanSegmentSetRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan segment set %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanSegmentSetsUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanSegmentSetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	name, _ := cmd.Flags().GetString("name")
	isDefault, _ := cmd.Flags().GetBool("is-default")
	startOffsetMinutes, _ := cmd.Flags().GetInt("start-offset-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanSegmentSetsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		JobProductionPlanID: jobProductionPlanID,
		Name:                name,
		IsDefault:           isDefault,
		StartOffsetMinutes:  startOffsetMinutes,
	}, nil
}
