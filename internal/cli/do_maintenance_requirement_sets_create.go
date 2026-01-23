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

type doMaintenanceRequirementSetsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	MaintenanceType         string
	Status                  string
	IsTemplate              bool
	TemplateName            string
	IsArchived              bool
	EquipmentClassification string
	Broker                  string
	WorkOrder               string
}

func newDoMaintenanceRequirementSetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a maintenance requirement set",
		Long: `Create a maintenance requirement set.

Required flags:
  --maintenance-type  Maintenance type (inspection or maintenance)
  --broker            Broker ID

	Optional flags:
  --status                   Status (editing, ready_for_work, on_hold, in_progress, completed)
  --is-template              Whether the set is a template (defaults to true)
  --template-name            Template name (required for templates)
  --is-archived              Archive the set
  --equipment-classification Equipment classification ID
  --work-order               Work order ID (cannot be used with templates)`,
		Example: `  # Create a non-template maintenance requirement set
  xbe do maintenance-requirement-sets create --maintenance-type maintenance --broker 123 --is-template=false

  # Create a template set
  xbe do maintenance-requirement-sets create --maintenance-type inspection --broker 123 --is-template --template-name "Quarterly Inspection"

  # Create linked to a work order
  xbe do maintenance-requirement-sets create --maintenance-type maintenance --broker 123 --is-template=false --work-order 456`,
		Args: cobra.NoArgs,
		RunE: runDoMaintenanceRequirementSetsCreate,
	}
	initDoMaintenanceRequirementSetsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementSetsCmd.AddCommand(newDoMaintenanceRequirementSetsCreateCmd())
}

func initDoMaintenanceRequirementSetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maintenance-type", "", "Maintenance type (inspection or maintenance) (required)")
	cmd.Flags().String("status", "", "Status (editing, ready_for_work, on_hold, in_progress, completed)")
	cmd.Flags().Bool("is-template", false, "Whether the set is a template (defaults to true)")
	cmd.Flags().String("template-name", "", "Template name (required for templates)")
	cmd.Flags().Bool("is-archived", false, "Archive the set")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("work-order", "", "Work order ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("maintenance-type")
	cmd.MarkFlagRequired("broker")
}

func runDoMaintenanceRequirementSetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceRequirementSetsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !cmd.Flags().Changed("is-template") && strings.TrimSpace(opts.TemplateName) == "" {
		err := fmt.Errorf("--template-name is required unless --is-template=false is set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.IsTemplate && strings.TrimSpace(opts.TemplateName) == "" {
		err := fmt.Errorf("--template-name is required when --is-template is true")
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

	attributes := map[string]any{
		"maintenance-type": opts.MaintenanceType,
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("is-template") {
		attributes["is-template"] = opts.IsTemplate
	}
	if opts.TemplateName != "" {
		attributes["template-name"] = opts.TemplateName
	}
	if cmd.Flags().Changed("is-archived") {
		attributes["is-archived"] = opts.IsArchived
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	if opts.EquipmentClassification != "" {
		relationships["equipment-classification"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassification,
			},
		}
	}
	if opts.WorkOrder != "" {
		relationships["work-order"] = map[string]any{
			"data": map[string]any{
				"type": "work-orders",
				"id":   opts.WorkOrder,
			},
		}
	}

	data := map[string]any{
		"type":          "maintenance-requirement-sets",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/maintenance-requirement-sets", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created maintenance requirement set %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementSetsCreateOptions(cmd *cobra.Command) (doMaintenanceRequirementSetsCreateOptions, error) {
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

	return doMaintenanceRequirementSetsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
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
