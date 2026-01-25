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

type doActionItemTrackersUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
	Status                    string
	DevEffortSize             string
	DevEffortMinutes          int
	HasDueDateAgreement       string
	IsUnplanned               string
	PriorityPosition          string
	DevAssigneeID             string
	CustomerSuccessAssigneeID string
}

func newDoActionItemTrackersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an action item tracker",
		Long: `Update an existing action item tracker.

Only the fields you specify will be updated.

Arguments:
  <id>    Action item tracker ID (required)

Flags:
  --status                  Update tracker status
  --dev-effort-size         Update dev effort size: s, m, l, xl, xxl
  --dev-effort-minutes      Update dev effort estimate in minutes (>= 0)
  --has-due-date-agreement  Update due date agreement (true/false)
  --is-unplanned            Update unplanned status (true/false)
  --priority-position       Update priority position (integer or ranking hint)
  --dev-assignee            Update development assignee user ID
  --customer-success-assignee  Update customer success assignee user ID`,
		Example: `  # Update status
  xbe do action-item-trackers update 123 --status in_development

  # Update effort sizing
  xbe do action-item-trackers update 123 --dev-effort-size l --dev-effort-minutes 240`,
		Args: cobra.ExactArgs(1),
		RunE: runDoActionItemTrackersUpdate,
	}
	initDoActionItemTrackersUpdateFlags(cmd)
	return cmd
}

func init() {
	doActionItemTrackersCmd.AddCommand(newDoActionItemTrackersUpdateCmd())
}

func initDoActionItemTrackersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Tracker status")
	cmd.Flags().String("dev-effort-size", "", "Dev effort size: s, m, l, xl, xxl")
	cmd.Flags().Int("dev-effort-minutes", 0, "Dev effort estimate in minutes (>= 0)")
	cmd.Flags().String("has-due-date-agreement", "", "Due date agreement (true/false)")
	cmd.Flags().String("is-unplanned", "", "Unplanned tracker (true/false)")
	cmd.Flags().String("priority-position", "", "Priority position (integer or ranking hint)")
	cmd.Flags().String("dev-assignee", "", "Development assignee user ID")
	cmd.Flags().String("customer-success-assignee", "", "Customer success assignee user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemTrackersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemTrackersUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.DevEffortSize != "" {
		attributes["dev-effort-size"] = opts.DevEffortSize
	}
	if cmd.Flags().Changed("dev-effort-minutes") {
		attributes["dev-effort-minutes"] = opts.DevEffortMinutes
	}
	if opts.HasDueDateAgreement != "" {
		attributes["has-due-date-agreement"] = opts.HasDueDateAgreement == "true"
	}
	if opts.IsUnplanned != "" {
		attributes["is-unplanned"] = opts.IsUnplanned == "true"
	}
	if opts.PriorityPosition != "" {
		attributes["priority-position"] = opts.PriorityPosition
	}
	if opts.DevAssigneeID != "" {
		relationships["dev-assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.DevAssigneeID,
			},
		}
	}
	if opts.CustomerSuccessAssigneeID != "" {
		relationships["customer-success-assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CustomerSuccessAssigneeID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "action-item-trackers",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/action-item-trackers/"+opts.ID, jsonBody)
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

	details := buildActionItemTrackerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemTrackerDetails(cmd, details)
}

func parseDoActionItemTrackersUpdateOptions(cmd *cobra.Command, args []string) (doActionItemTrackersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	devEffortSize, _ := cmd.Flags().GetString("dev-effort-size")
	devEffortMinutes, _ := cmd.Flags().GetInt("dev-effort-minutes")
	hasDueDateAgreement, _ := cmd.Flags().GetString("has-due-date-agreement")
	isUnplanned, _ := cmd.Flags().GetString("is-unplanned")
	priorityPosition, _ := cmd.Flags().GetString("priority-position")
	devAssigneeID, _ := cmd.Flags().GetString("dev-assignee")
	customerSuccessAssigneeID, _ := cmd.Flags().GetString("customer-success-assignee")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemTrackersUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
		Status:                    status,
		DevEffortSize:             devEffortSize,
		DevEffortMinutes:          devEffortMinutes,
		HasDueDateAgreement:       hasDueDateAgreement,
		IsUnplanned:               isUnplanned,
		PriorityPosition:          priorityPosition,
		DevAssigneeID:             devAssigneeID,
		CustomerSuccessAssigneeID: customerSuccessAssigneeID,
	}, nil
}
