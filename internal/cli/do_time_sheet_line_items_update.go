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

type doTimeSheetLineItemsUpdateOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	StartAt                         string
	EndAt                           string
	BreakMinutes                    int
	Description                     string
	SkipValidateOverlap             string
	IsNonJobLineItem                string
	CostCode                        string
	CraftClass                      string
	EquipmentRequirement            string
	MaintenanceRequirement          string
	TimeSheetLineItemClassification string
	ProjectCostClassification       string
	ExplicitJobProductionPlan       string
}

func newDoTimeSheetLineItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time sheet line item",
		Long: `Update an existing time sheet line item.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The time sheet line item ID (required)

Flags:
  --start-at                          Update start timestamp (ISO 8601)
  --end-at                            Update end timestamp (ISO 8601)
  --break-minutes                     Update break minutes
  --description                       Update description
  --skip-validate-overlap             Update skip overlap validation (true/false)
  --is-non-job-line-item              Update non-job line item flag (true/false)
  --cost-code                         Update cost code ID (empty to clear)
  --craft-class                       Update craft class ID (empty to clear)
  --equipment-requirement             Update equipment requirement ID (empty to clear)
  --maintenance-requirement           Update maintenance requirement ID (empty to clear)
  --time-sheet-line-item-classification Update classification ID (empty to clear)
  --project-cost-classification       Update project cost classification ID (empty to clear)
  --explicit-job-production-plan      Update explicit job production plan ID (empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update break minutes
  xbe do time-sheet-line-items update 123 --break-minutes 15

  # Update classification and cost code
  xbe do time-sheet-line-items update 123 --time-sheet-line-item-classification 456 --cost-code 789

  # Clear an equipment requirement
  xbe do time-sheet-line-items update 123 --equipment-requirement \"\"

  # Get JSON output
  xbe do time-sheet-line-items update 123 --description \"Updated\" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetLineItemsUpdate,
	}
	initDoTimeSheetLineItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetLineItemsCmd.AddCommand(newDoTimeSheetLineItemsUpdateCmd())
}

func initDoTimeSheetLineItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().Int("break-minutes", 0, "Break minutes")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("skip-validate-overlap", "", "Skip overlap validation (true/false)")
	cmd.Flags().String("is-non-job-line-item", "", "Non-job line item flag (true/false)")
	cmd.Flags().String("cost-code", "", "Cost code ID")
	cmd.Flags().String("craft-class", "", "Craft class ID")
	cmd.Flags().String("equipment-requirement", "", "Equipment requirement ID")
	cmd.Flags().String("maintenance-requirement", "", "Maintenance requirement ID")
	cmd.Flags().String("time-sheet-line-item-classification", "", "Time sheet line item classification ID")
	cmd.Flags().String("project-cost-classification", "", "Project cost classification ID")
	cmd.Flags().String("explicit-job-production-plan", "", "Explicit job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetLineItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetLineItemsUpdateOptions(cmd)
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
		return fmt.Errorf("time sheet line item id is required")
	}

	attributes := map[string]any{}
	relationships := map[string]any{}
	hasChanges := false

	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
		hasChanges = true
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
		hasChanges = true
	}
	if cmd.Flags().Changed("break-minutes") {
		attributes["break-minutes"] = opts.BreakMinutes
		hasChanges = true
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
		hasChanges = true
	}
	if opts.SkipValidateOverlap != "" {
		attributes["skip-validate-overlap"] = opts.SkipValidateOverlap == "true"
		hasChanges = true
	}
	if opts.IsNonJobLineItem != "" {
		attributes["is-non-job-line-item"] = opts.IsNonJobLineItem == "true"
		hasChanges = true
	}

	if cmd.Flags().Changed("cost-code") {
		if opts.CostCode == "" {
			relationships["cost-code"] = map[string]any{"data": nil}
		} else {
			relationships["cost-code"] = map[string]any{
				"data": map[string]any{
					"type": "cost-codes",
					"id":   opts.CostCode,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("craft-class") {
		if opts.CraftClass == "" {
			relationships["craft-class"] = map[string]any{"data": nil}
		} else {
			relationships["craft-class"] = map[string]any{
				"data": map[string]any{
					"type": "craft-classes",
					"id":   opts.CraftClass,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("equipment-requirement") {
		if opts.EquipmentRequirement == "" {
			relationships["equipment-requirement"] = map[string]any{"data": nil}
		} else {
			relationships["equipment-requirement"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-requirements",
					"id":   opts.EquipmentRequirement,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("maintenance-requirement") {
		if opts.MaintenanceRequirement == "" {
			relationships["maintenance-requirement"] = map[string]any{"data": nil}
		} else {
			relationships["maintenance-requirement"] = map[string]any{
				"data": map[string]any{
					"type": "maintenance-requirements",
					"id":   opts.MaintenanceRequirement,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("time-sheet-line-item-classification") {
		if opts.TimeSheetLineItemClassification == "" {
			relationships["time-sheet-line-item-classification"] = map[string]any{"data": nil}
		} else {
			relationships["time-sheet-line-item-classification"] = map[string]any{
				"data": map[string]any{
					"type": "time-sheet-line-item-classifications",
					"id":   opts.TimeSheetLineItemClassification,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("project-cost-classification") {
		if opts.ProjectCostClassification == "" {
			relationships["project-cost-classification"] = map[string]any{"data": nil}
		} else {
			relationships["project-cost-classification"] = map[string]any{
				"data": map[string]any{
					"type": "project-cost-classifications",
					"id":   opts.ProjectCostClassification,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-job-production-plan") {
		if opts.ExplicitJobProductionPlan == "" {
			relationships["explicit-job-production-plan"] = map[string]any{"data": nil}
		} else {
			relationships["explicit-job-production-plan"] = map[string]any{
				"data": map[string]any{
					"type": "job-production-plans",
					"id":   opts.ExplicitJobProductionPlan,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"id":   id,
		"type": "time-sheet-line-items",
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/time-sheet-line-items/"+id, jsonBody)
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

	row := buildTimeSheetLineItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time sheet line item %s\n", row.ID)
	return nil
}

func parseDoTimeSheetLineItemsUpdateOptions(cmd *cobra.Command) (doTimeSheetLineItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	breakMinutes, _ := cmd.Flags().GetInt("break-minutes")
	description, _ := cmd.Flags().GetString("description")
	skipValidateOverlap, _ := cmd.Flags().GetString("skip-validate-overlap")
	isNonJobLineItem, _ := cmd.Flags().GetString("is-non-job-line-item")
	costCode, _ := cmd.Flags().GetString("cost-code")
	craftClass, _ := cmd.Flags().GetString("craft-class")
	equipmentRequirement, _ := cmd.Flags().GetString("equipment-requirement")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	timeSheetLineItemClassification, _ := cmd.Flags().GetString("time-sheet-line-item-classification")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	explicitJobProductionPlan, _ := cmd.Flags().GetString("explicit-job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetLineItemsUpdateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		StartAt:                         startAt,
		EndAt:                           endAt,
		BreakMinutes:                    breakMinutes,
		Description:                     description,
		SkipValidateOverlap:             skipValidateOverlap,
		IsNonJobLineItem:                isNonJobLineItem,
		CostCode:                        costCode,
		CraftClass:                      craftClass,
		EquipmentRequirement:            equipmentRequirement,
		MaintenanceRequirement:          maintenanceRequirement,
		TimeSheetLineItemClassification: timeSheetLineItemClassification,
		ProjectCostClassification:       projectCostClassification,
		ExplicitJobProductionPlan:       explicitJobProductionPlan,
	}, nil
}
