package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type timeSheetLineItemsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetLineItemDetails struct {
	ID                                       string   `json:"id"`
	TimeSheetID                              string   `json:"time_sheet_id,omitempty"`
	StartAt                                  string   `json:"start_at,omitempty"`
	EndAt                                    string   `json:"end_at,omitempty"`
	BreakMinutes                             int      `json:"break_minutes,omitempty"`
	Description                              string   `json:"description,omitempty"`
	SkipValidateOverlap                      bool     `json:"skip_validate_overlap"`
	IsNonJobLineItem                         bool     `json:"is_non_job_line_item"`
	DurationSeconds                          int      `json:"duration_seconds,omitempty"`
	CostCodeID                               string   `json:"cost_code_id,omitempty"`
	CraftClassID                             string   `json:"craft_class_id,omitempty"`
	CraftClassEffectiveID                    string   `json:"craft_class_effective_id,omitempty"`
	TimeSheetLineItemClassificationID        string   `json:"time_sheet_line_item_classification_id,omitempty"`
	ProjectCostClassificationID              string   `json:"project_cost_classification_id,omitempty"`
	EquipmentRequirementID                   string   `json:"equipment_requirement_id,omitempty"`
	EquipmentRequirementIDs                  []string `json:"equipment_requirement_ids,omitempty"`
	TimeSheetLineItemEquipmentRequirementIDs []string `json:"time_sheet_line_item_equipment_requirement_ids,omitempty"`
	MaintenanceRequirementID                 string   `json:"maintenance_requirement_id,omitempty"`
	TimeCardID                               string   `json:"time_card_id,omitempty"`
	JobProductionPlanID                      string   `json:"job_production_plan_id,omitempty"`
	ImplicitJobProductionPlanID              string   `json:"implicit_job_production_plan_id,omitempty"`
	ExplicitJobProductionPlanID              string   `json:"explicit_job_production_plan_id,omitempty"`
}

func newTimeSheetLineItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet line item details",
		Long: `Show the full details of a time sheet line item.

Output Fields:
  ID
  Time Sheet ID
  Start At
  End At
  Break Minutes
  Description
  Skip Validate Overlap
  Is Non-Job Line Item
  Duration Seconds
  Cost Code ID
  Craft Class ID
  Craft Class Effective ID
  Time Sheet Line Item Classification ID
  Project Cost Classification ID
  Equipment Requirement ID
  Equipment Requirement IDs
  Time Sheet Line Item Equipment Requirement IDs
  Maintenance Requirement ID
  Time Card ID
  Job Production Plan ID
  Implicit Job Production Plan ID
  Explicit Job Production Plan ID

Arguments:
  <id>    The time sheet line item ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet line item
  xbe view time-sheet-line-items show 123

  # Get JSON output
  xbe view time-sheet-line-items show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetLineItemsShow,
	}
	initTimeSheetLineItemsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetLineItemsCmd.AddCommand(newTimeSheetLineItemsShowCmd())
}

func initTimeSheetLineItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetLineItemsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeSheetLineItemsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("time sheet line item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-line-items]", "start-at,end-at,break-minutes,description,skip-validate-overlap,is-non-job-line-item,duration-seconds,time-sheet,cost-code,craft-class,craft-class-effective,time-sheet-line-item-classification,project-cost-classification,equipment-requirement,maintenance-requirement,time-card,job-production-plan,implicit-job-production-plan,explicit-job-production-plan,time-sheet-line-item-equipment-requirements,equipment-requirements")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-line-items/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTimeSheetLineItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetLineItemDetails(cmd, details)
}

func parseTimeSheetLineItemsShowOptions(cmd *cobra.Command) (timeSheetLineItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetLineItemsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetLineItemDetails(resp jsonAPISingleResponse) timeSheetLineItemDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return timeSheetLineItemDetails{
		ID:                                       resource.ID,
		TimeSheetID:                              relationshipIDFromMap(resource.Relationships, "time-sheet"),
		StartAt:                                  formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                                    formatDateTime(stringAttr(attrs, "end-at")),
		BreakMinutes:                             intAttr(attrs, "break-minutes"),
		Description:                              stringAttr(attrs, "description"),
		SkipValidateOverlap:                      boolAttr(attrs, "skip-validate-overlap"),
		IsNonJobLineItem:                         boolAttr(attrs, "is-non-job-line-item"),
		DurationSeconds:                          intAttr(attrs, "duration-seconds"),
		CostCodeID:                               relationshipIDFromMap(resource.Relationships, "cost-code"),
		CraftClassID:                             relationshipIDFromMap(resource.Relationships, "craft-class"),
		CraftClassEffectiveID:                    relationshipIDFromMap(resource.Relationships, "craft-class-effective"),
		TimeSheetLineItemClassificationID:        relationshipIDFromMap(resource.Relationships, "time-sheet-line-item-classification"),
		ProjectCostClassificationID:              relationshipIDFromMap(resource.Relationships, "project-cost-classification"),
		EquipmentRequirementID:                   relationshipIDFromMap(resource.Relationships, "equipment-requirement"),
		EquipmentRequirementIDs:                  relationshipIDsFromMap(resource.Relationships, "equipment-requirements"),
		TimeSheetLineItemEquipmentRequirementIDs: relationshipIDsFromMap(resource.Relationships, "time-sheet-line-item-equipment-requirements"),
		MaintenanceRequirementID:                 relationshipIDFromMap(resource.Relationships, "maintenance-requirement"),
		TimeCardID:                               relationshipIDFromMap(resource.Relationships, "time-card"),
		JobProductionPlanID:                      relationshipIDFromMap(resource.Relationships, "job-production-plan"),
		ImplicitJobProductionPlanID:              relationshipIDFromMap(resource.Relationships, "implicit-job-production-plan"),
		ExplicitJobProductionPlanID:              relationshipIDFromMap(resource.Relationships, "explicit-job-production-plan"),
	}
}

func renderTimeSheetLineItemDetails(cmd *cobra.Command, details timeSheetLineItemDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet ID: %s\n", details.TimeSheetID)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.BreakMinutes != 0 {
		fmt.Fprintf(out, "Break Minutes: %d\n", details.BreakMinutes)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	fmt.Fprintf(out, "Skip Validate Overlap: %t\n", details.SkipValidateOverlap)
	fmt.Fprintf(out, "Is Non-Job Line Item: %t\n", details.IsNonJobLineItem)
	if details.DurationSeconds != 0 {
		fmt.Fprintf(out, "Duration Seconds: %d\n", details.DurationSeconds)
	}
	if details.CostCodeID != "" {
		fmt.Fprintf(out, "Cost Code ID: %s\n", details.CostCodeID)
	}
	if details.CraftClassID != "" {
		fmt.Fprintf(out, "Craft Class ID: %s\n", details.CraftClassID)
	}
	if details.CraftClassEffectiveID != "" {
		fmt.Fprintf(out, "Craft Class Effective ID: %s\n", details.CraftClassEffectiveID)
	}
	if details.TimeSheetLineItemClassificationID != "" {
		fmt.Fprintf(out, "Time Sheet Line Item Classification ID: %s\n", details.TimeSheetLineItemClassificationID)
	}
	if details.ProjectCostClassificationID != "" {
		fmt.Fprintf(out, "Project Cost Classification ID: %s\n", details.ProjectCostClassificationID)
	}
	if details.EquipmentRequirementID != "" {
		fmt.Fprintf(out, "Equipment Requirement ID: %s\n", details.EquipmentRequirementID)
	}
	if len(details.EquipmentRequirementIDs) > 0 {
		fmt.Fprintf(out, "Equipment Requirement IDs: %s\n", strings.Join(details.EquipmentRequirementIDs, ", "))
	}
	if len(details.TimeSheetLineItemEquipmentRequirementIDs) > 0 {
		fmt.Fprintf(out, "Time Sheet Line Item Equipment Requirement IDs: %s\n", strings.Join(details.TimeSheetLineItemEquipmentRequirementIDs, ", "))
	}
	if details.MaintenanceRequirementID != "" {
		fmt.Fprintf(out, "Maintenance Requirement ID: %s\n", details.MaintenanceRequirementID)
	}
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card ID: %s\n", details.TimeCardID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.ImplicitJobProductionPlanID != "" {
		fmt.Fprintf(out, "Implicit Job Production Plan ID: %s\n", details.ImplicitJobProductionPlanID)
	}
	if details.ExplicitJobProductionPlanID != "" {
		fmt.Fprintf(out, "Explicit Job Production Plan ID: %s\n", details.ExplicitJobProductionPlanID)
	}

	return nil
}
