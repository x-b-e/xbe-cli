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

type doActionItemTrackersCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ActionItemID              string
	Status                    string
	DevEffortSize             string
	DevEffortMinutes          int
	HasDueDateAgreement       string
	IsUnplanned               string
	PriorityPosition          string
	DevAssigneeID             string
	CustomerSuccessAssigneeID string
}

func newDoActionItemTrackersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an action item tracker",
		Long: `Create a new action item tracker.

Required flags:
  --action-item             Action item ID

Optional flags:
  --status                  Tracker status
  --dev-effort-size         Dev effort size: s, m, l, xl, xxl
  --dev-effort-minutes      Dev effort estimate in minutes (>= 0)
  --has-due-date-agreement  Due date agreement (true/false)
  --is-unplanned            Unplanned tracker (true/false)
  --priority-position       Priority position (integer or ranking hint)
  --dev-assignee            Development assignee user ID
  --customer-success-assignee  Customer success assignee user ID`,
		Example: `  # Create a tracker with status and size
  xbe do action-item-trackers create --action-item 123 --status ready_for_work --dev-effort-size m

  # Create a tracker with assignments
  xbe do action-item-trackers create --action-item 123 --dev-assignee 45 --customer-success-assignee 67`,
		RunE: runDoActionItemTrackersCreate,
	}
	initDoActionItemTrackersCreateFlags(cmd)
	return cmd
}

func init() {
	doActionItemTrackersCmd.AddCommand(newDoActionItemTrackersCreateCmd())
}

func initDoActionItemTrackersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("action-item", "", "Action item ID (required)")
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

	cmd.MarkFlagRequired("action-item")
}

func runDoActionItemTrackersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoActionItemTrackersCreateOptions(cmd)
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

	relationships := map[string]any{
		"action-item": map[string]any{
			"data": map[string]any{
				"type": "action-items",
				"id":   opts.ActionItemID,
			},
		},
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

	data := map[string]any{
		"type": "action-item-trackers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/action-item-trackers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created action item tracker %s\n", details.ID)
	return nil
}

func parseDoActionItemTrackersCreateOptions(cmd *cobra.Command) (doActionItemTrackersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	actionItemID, _ := cmd.Flags().GetString("action-item")
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

	return doActionItemTrackersCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ActionItemID:              actionItemID,
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
