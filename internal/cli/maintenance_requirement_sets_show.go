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

type maintenanceRequirementSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type maintenanceRequirementSetDetails struct {
	ID                        string   `json:"id"`
	MaintenanceType           string   `json:"maintenance_type,omitempty"`
	Status                    string   `json:"status,omitempty"`
	IsTemplate                bool     `json:"is_template"`
	TemplateName              string   `json:"template_name,omitempty"`
	IsArchived                bool     `json:"is_archived"`
	CompletedAt               string   `json:"completed_at,omitempty"`
	RuleEvaluationMeta        any      `json:"rule_evaluation_meta,omitempty"`
	BrokerID                  string   `json:"broker_id,omitempty"`
	BrokerName                string   `json:"broker_name,omitempty"`
	EquipmentClassificationID string   `json:"equipment_classification_id,omitempty"`
	EquipmentClassification   string   `json:"equipment_classification,omitempty"`
	WorkOrderID               string   `json:"work_order_id,omitempty"`
	MaintenanceRequirementIDs []string `json:"maintenance_requirement_ids,omitempty"`
	EquipmentIDs              []string `json:"equipment_ids,omitempty"`
	FileAttachmentIDs         []string `json:"file_attachment_ids,omitempty"`
}

func newMaintenanceRequirementSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement set details",
		Long: `Show the full details of a maintenance requirement set.

Output Fields:
  ID                         Maintenance requirement set identifier
  Maintenance Type           Maintenance type (inspection/maintenance)
  Status                     Current status
  Is Template                Whether this is a template set
  Template Name              Template name (if template)
  Is Archived                Archived status
  Completed At               Completed timestamp
  Broker                     Broker name
  Broker ID                  Broker identifier
  Equipment Classification   Equipment classification name
  Equipment Classification ID Equipment classification identifier
  Work Order ID              Work order identifier
  Maintenance Requirement IDs Linked maintenance requirement IDs
  Equipment IDs              Linked equipment IDs
  File Attachment IDs        Linked file attachment IDs
  Rule Evaluation Meta       Rule evaluation metadata

Arguments:
  <id>    Maintenance requirement set ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a maintenance requirement set
  xbe view maintenance-requirement-sets show 123

  # JSON output
  xbe view maintenance-requirement-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRequirementSetsShow,
	}
	initMaintenanceRequirementSetsShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementSetsCmd.AddCommand(newMaintenanceRequirementSetsShowCmd())
}

func initMaintenanceRequirementSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceRequirementSetsShowOptions(cmd)
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
		return fmt.Errorf("maintenance requirement set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[maintenance-requirement-sets]", "maintenance-type,status,is-template,template-name,is-archived,completed-at,rule-evaluation-meta,broker,equipment-classification,work-order,maintenance-requirements,equipments,file-attachments")
	query.Set("include", "broker,equipment-classification")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[equipment-classifications]", "name,abbreviation")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-sets/"+id, query)
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

	details := buildMaintenanceRequirementSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaintenanceRequirementSetDetails(cmd, details)
}

func parseMaintenanceRequirementSetsShowOptions(cmd *cobra.Command) (maintenanceRequirementSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaintenanceRequirementSetDetails(resp jsonAPISingleResponse) maintenanceRequirementSetDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := maintenanceRequirementSetDetails{
		ID:                 resp.Data.ID,
		MaintenanceType:    stringAttr(attrs, "maintenance-type"),
		Status:             stringAttr(attrs, "status"),
		IsTemplate:         boolAttr(attrs, "is-template"),
		TemplateName:       strings.TrimSpace(stringAttr(attrs, "template-name")),
		IsArchived:         boolAttr(attrs, "is-archived"),
		CompletedAt:        formatDateTime(stringAttr(attrs, "completed-at")),
		RuleEvaluationMeta: anyAttr(attrs, "rule-evaluation-meta"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}
	if rel, ok := resp.Data.Relationships["equipment-classification"]; ok && rel.Data != nil {
		details.EquipmentClassificationID = rel.Data.ID
		if ec, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			name := stringAttr(ec.Attributes, "name")
			abbrev := stringAttr(ec.Attributes, "abbreviation")
			details.EquipmentClassification = firstNonEmpty(formatEquipmentClassificationLabel(name, abbrev), name)
		}
	}
	if rel, ok := resp.Data.Relationships["work-order"]; ok && rel.Data != nil {
		details.WorkOrderID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["maintenance-requirements"]; ok {
		details.MaintenanceRequirementIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["equipments"]; ok {
		details.EquipmentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderMaintenanceRequirementSetDetails(cmd *cobra.Command, details maintenanceRequirementSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaintenanceType != "" {
		fmt.Fprintf(out, "Maintenance Type: %s\n", details.MaintenanceType)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	fmt.Fprintf(out, "Is Template: %s\n", formatBool(details.IsTemplate))
	if details.TemplateName != "" {
		fmt.Fprintf(out, "Template Name: %s\n", details.TemplateName)
	}
	fmt.Fprintf(out, "Is Archived: %s\n", formatBool(details.IsArchived))
	if details.CompletedAt != "" {
		fmt.Fprintf(out, "Completed At: %s\n", details.CompletedAt)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.EquipmentClassification != "" {
		fmt.Fprintf(out, "Equipment Classification: %s\n", details.EquipmentClassification)
	}
	if details.EquipmentClassificationID != "" {
		fmt.Fprintf(out, "Equipment Classification ID: %s\n", details.EquipmentClassificationID)
	}
	if details.WorkOrderID != "" {
		fmt.Fprintf(out, "Work Order ID: %s\n", details.WorkOrderID)
	}
	if len(details.MaintenanceRequirementIDs) > 0 {
		fmt.Fprintf(out, "Maintenance Requirement IDs: %s\n", strings.Join(details.MaintenanceRequirementIDs, ", "))
	}
	if len(details.EquipmentIDs) > 0 {
		fmt.Fprintf(out, "Equipment IDs: %s\n", strings.Join(details.EquipmentIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachment IDs: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}
	if details.RuleEvaluationMeta != nil {
		fmt.Fprintf(out, "Rule Evaluation Meta: %s\n", formatAny(details.RuleEvaluationMeta))
	}

	return nil
}
