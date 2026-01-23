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

type doJobProductionPlanSegmentSetsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	JobProductionPlanID string
	Name                string
	IsDefault           bool
	StartOffsetMinutes  int
}

func newDoJobProductionPlanSegmentSetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan segment set",
		Long: `Create a job production plan segment set.

Required:
  --job-production-plan  Job production plan ID

Optional:
  --name                 Segment set name
  --is-default           Mark the set as default
  --start-offset-minutes Start offset in minutes`,
		Example: `  # Create a segment set
  xbe do job-production-plan-segment-sets create --job-production-plan 123 --name "AM shift"

  # Create a default segment set with offset
  xbe do job-production-plan-segment-sets create \
    --job-production-plan 123 \
    --is-default \
    --start-offset-minutes 15`,
		RunE: runDoJobProductionPlanSegmentSetsCreate,
	}
	initDoJobProductionPlanSegmentSetsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanSegmentSetsCmd.AddCommand(newDoJobProductionPlanSegmentSetsCreateCmd())
}

func initDoJobProductionPlanSegmentSetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("name", "", "Segment set name")
	cmd.Flags().Bool("is-default", false, "Mark the set as default")
	cmd.Flags().Int("start-offset-minutes", 0, "Start offset in minutes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan")
}

func runDoJobProductionPlanSegmentSetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanSegmentSetsCreateOptions(cmd)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("is-default") {
		attributes["is-default"] = opts.IsDefault
	}
	if cmd.Flags().Changed("start-offset-minutes") {
		attributes["start-offset-minutes"] = opts.StartOffsetMinutes
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-production-plan-segment-sets",
			"attributes": attributes,
			"relationships": map[string]any{
				"job-production-plan": map[string]any{
					"data": map[string]any{
						"type": "job-production-plans",
						"id":   opts.JobProductionPlanID,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-segment-sets", jsonBody)
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

	if opts.JSON {
		row := buildJobProductionPlanSegmentSetRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan segment set %s\n", resp.Data.ID)
	return nil
}

func parseDoJobProductionPlanSegmentSetsCreateOptions(cmd *cobra.Command) (doJobProductionPlanSegmentSetsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	name, _ := cmd.Flags().GetString("name")
	isDefault, _ := cmd.Flags().GetBool("is-default")
	startOffsetMinutes, _ := cmd.Flags().GetInt("start-offset-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanSegmentSetsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		JobProductionPlanID: jobProductionPlanID,
		Name:                name,
		IsDefault:           isDefault,
		StartOffsetMinutes:  startOffsetMinutes,
	}, nil
}
