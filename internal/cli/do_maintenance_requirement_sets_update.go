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

type doMaintenanceRequirementSetsUpdateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ID                      string
	MaintenanceType         string
	Status                  string
	IsTemplate              bool
	TemplateName            string
	IsArchived              bool
	EquipmentClassification string
	Broker                  string
	WorkOrder               string
}

func newDoMaintenanceRequirementSetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a maintenance requirement set",
		Long: `Update a maintenance requirement set.

Optional flags:
  --maintenance-type          Maintenance type (inspection or maintenance)
  --status                    Status (editing, ready_for_work, on_hold, in_progress, completed)
  --is-template               Whether the set is a template
  --template-name             Template name
  --is-archived               Archive status
  --equipment-classification  Equipment classification ID (set empty to clear)
  --broker                    Broker ID
  --work-order                Work order ID (set empty to clear)`,
		Example: `  # Update status
  xbe do maintenance-requirement-sets update 123 --status ready_for_work

  # Update template name
  xbe do maintenance-requirement-sets update 123 --template-name "Updated Template"

  # Update equipment classification
  xbe do maintenance-requirement-sets update 123 --equipment-classification 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaintenanceRequirementSetsUpdate,
	}
	initDoMaintenanceRequirementSetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementSetsCmd.AddCommand(newDoMaintenanceRequirementSetsUpdateCmd())
}

func initDoMaintenanceRequirementSetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maintenance-type", "", "Maintenance type (inspection or maintenance)")
	cmd.Flags().String("status", "", "Status (editing, ready_for_work, on_hold, in_progress, completed)")
	cmd.Flags().Bool("is-template", false, "Whether the set is a template")
	cmd.Flags().String("template-name", "", "Template name")
	cmd.Flags().Bool("is-archived", false, "Archive status")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("work-order", "", "Work order ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementSetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaintenanceRequirementSetsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("maintenance-type") {
		attributes["maintenance-type"] = opts.MaintenanceType
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("is-template") {
		attributes["is-template"] = opts.IsTemplate
	}
	if cmd.Flags().Changed("template-name") {
		attributes["template-name"] = opts.TemplateName
	}
	if cmd.Flags().Changed("is-archived") {
		attributes["is-archived"] = opts.IsArchived
	}

	if cmd.Flags().Changed("equipment-classification") {
		if opts.EquipmentClassification == "" {
			relationships["equipment-classification"] = map[string]any{"data": nil}
		} else {
			relationships["equipment-classification"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-classifications",
					"id":   opts.EquipmentClassification,
				},
			}
		}
	}
	if cmd.Flags().Changed("broker") {
		if opts.Broker == "" {
			relationships["broker"] = map[string]any{"data": nil}
		} else {
			relationships["broker"] = map[string]any{
				"data": map[string]any{
					"type": "brokers",
					"id":   opts.Broker,
				},
			}
		}
	}
	if cmd.Flags().Changed("work-order") {
		if opts.WorkOrder == "" {
			relationships["work-order"] = map[string]any{"data": nil}
		} else {
			relationships["work-order"] = map[string]any{
				"data": map[string]any{
					"type": "work-orders",
					"id":   opts.WorkOrder,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "maintenance-requirement-sets",
		"id":   opts.ID,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/maintenance-requirement-sets/"+opts.ID, jsonBody)
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

	row := buildMaintenanceRequirementSetRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated maintenance requirement set %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementSetsUpdateOptions(cmd *cobra.Command, args []string) (doMaintenanceRequirementSetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maintenanceType, _ := cmd.Flags().GetString("maintenance-type")
	status, _ := cmd.Flags().GetString("status")
	isTemplate, _ := cmd.Flags().GetBool("is-template")
	templateName, _ := cmd.Flags().GetString("template-name")
	isArchived, _ := cmd.Flags().GetBool("is-archived")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	broker, _ := cmd.Flags().GetString("broker")
	workOrder, _ := cmd.Flags().GetString("work-order")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementSetsUpdateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ID:                      args[0],
		MaintenanceType:         maintenanceType,
		Status:                  status,
		IsTemplate:              isTemplate,
		TemplateName:            templateName,
		IsArchived:              isArchived,
		EquipmentClassification: equipmentClassification,
		Broker:                  broker,
		WorkOrder:               workOrder,
	}, nil
}
