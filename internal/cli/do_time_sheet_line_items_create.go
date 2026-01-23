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

type doTimeSheetLineItemsCreateOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	TimeSheet                       string
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

func newDoTimeSheetLineItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time sheet line item",
		Long: `Create a time sheet line item.

Required flags:
  --time-sheet   Time sheet ID

Optional flags:
  --start-at                          Start timestamp (ISO 8601)
  --end-at                            End timestamp (ISO 8601)
  --break-minutes                     Break minutes
  --description                       Description
  --skip-validate-overlap             Skip overlap validation (true/false)
  --is-non-job-line-item              Mark as non-job line item (true/false)
  --cost-code                         Cost code ID
  --craft-class                       Craft class ID
  --equipment-requirement             Equipment requirement ID
  --maintenance-requirement           Maintenance requirement ID
  --time-sheet-line-item-classification Time sheet line item classification ID
  --project-cost-classification       Project cost classification ID
  --explicit-job-production-plan      Explicit job production plan ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a time sheet line item
  xbe do time-sheet-line-items create \
    --time-sheet 123 \
    --start-at 2025-01-01T08:00:00Z \
    --end-at 2025-01-01T12:00:00Z \
    --break-minutes 30

  # Create a non-job line item
  xbe do time-sheet-line-items create --time-sheet 123 --is-non-job-line-item true

  # Get JSON output
  xbe do time-sheet-line-items create --time-sheet 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetLineItemsCreate,
	}
	initDoTimeSheetLineItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetLineItemsCmd.AddCommand(newDoTimeSheetLineItemsCreateCmd())
}

func initDoTimeSheetLineItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID (required)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().Int("break-minutes", 0, "Break minutes")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("skip-validate-overlap", "", "Skip overlap validation (true/false)")
	cmd.Flags().String("is-non-job-line-item", "", "Mark as non-job line item (true/false)")
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

func runDoTimeSheetLineItemsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetLineItemsCreateOptions(cmd)
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

	timeSheetID := strings.TrimSpace(opts.TimeSheet)
	if timeSheetID == "" {
		err := fmt.Errorf("--time-sheet is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("break-minutes") {
		attributes["break-minutes"] = opts.BreakMinutes
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.SkipValidateOverlap != "" {
		attributes["skip-validate-overlap"] = opts.SkipValidateOverlap == "true"
	}
	if opts.IsNonJobLineItem != "" {
		attributes["is-non-job-line-item"] = opts.IsNonJobLineItem == "true"
	}

	relationships := map[string]any{
		"time-sheet": map[string]any{
			"data": map[string]any{
				"type": "time-sheets",
				"id":   timeSheetID,
			},
		},
	}

	if opts.CostCode != "" {
		relationships["cost-code"] = map[string]any{
			"data": map[string]any{
				"type": "cost-codes",
				"id":   opts.CostCode,
			},
		}
	}
	if opts.CraftClass != "" {
		relationships["craft-class"] = map[string]any{
			"data": map[string]any{
				"type": "craft-classes",
				"id":   opts.CraftClass,
			},
		}
	}
	if opts.EquipmentRequirement != "" {
		relationships["equipment-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-requirements",
				"id":   opts.EquipmentRequirement,
			},
		}
	}
	if opts.MaintenanceRequirement != "" {
		relationships["maintenance-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirements",
				"id":   opts.MaintenanceRequirement,
			},
		}
	}
	if opts.TimeSheetLineItemClassification != "" {
		relationships["time-sheet-line-item-classification"] = map[string]any{
			"data": map[string]any{
				"type": "time-sheet-line-item-classifications",
				"id":   opts.TimeSheetLineItemClassification,
			},
		}
	}
	if opts.ProjectCostClassification != "" {
		relationships["project-cost-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-cost-classifications",
				"id":   opts.ProjectCostClassification,
			},
		}
	}
	if opts.ExplicitJobProductionPlan != "" {
		relationships["explicit-job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.ExplicitJobProductionPlan,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-sheet-line-items",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-line-items", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet line item %s\n", row.ID)
	return nil
}

func parseDoTimeSheetLineItemsCreateOptions(cmd *cobra.Command) (doTimeSheetLineItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
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

	return doTimeSheetLineItemsCreateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		TimeSheet:                       timeSheet,
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
