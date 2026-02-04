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

type maintenanceRequirementRulesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type maintenanceRequirementRuleDetails struct {
	ID                           string   `json:"id"`
	Rule                         string   `json:"rule,omitempty"`
	IsActive                     bool     `json:"is_active"`
	BrokerID                     string   `json:"broker_id,omitempty"`
	EquipmentID                  string   `json:"equipment_id,omitempty"`
	EquipmentClassificationID    string   `json:"equipment_classification_id,omitempty"`
	BusinessUnitID               string   `json:"business_unit_id,omitempty"`
	CreatedByID                  string   `json:"created_by_id,omitempty"`
	MaintenanceRequirementSetIDs []string `json:"maintenance_requirement_set_ids,omitempty"`
}

func newMaintenanceRequirementRulesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement rule details",
		Long: `Show the full details of a maintenance requirement rule.

Output Fields:
  ID
  Rule
  Is Active
  Broker ID
  Equipment ID
  Equipment Classification ID
  Business Unit ID
  Created By ID
  Maintenance Requirement Set IDs

Arguments:
  <id>    The maintenance requirement rule ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a maintenance requirement rule
  xbe view maintenance-requirement-rules show 123

  # Output as JSON
  xbe view maintenance-requirement-rules show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRequirementRulesShow,
	}
	initMaintenanceRequirementRulesShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementRulesCmd.AddCommand(newMaintenanceRequirementRulesShowCmd())
}

func initMaintenanceRequirementRulesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementRulesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaintenanceRequirementRulesShowOptions(cmd)
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
		return fmt.Errorf("maintenance requirement rule id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "broker,equipment,equipment-classification,business-unit,created-by,maintenance-requirement-sets")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-rules/"+id, query)
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

	details := buildMaintenanceRequirementRuleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaintenanceRequirementRuleDetails(cmd, details)
}

func parseMaintenanceRequirementRulesShowOptions(cmd *cobra.Command) (maintenanceRequirementRulesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementRulesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaintenanceRequirementRuleDetails(resp jsonAPISingleResponse) maintenanceRequirementRuleDetails {
	resource := resp.Data
	details := maintenanceRequirementRuleDetails{
		ID:       resource.ID,
		Rule:     stringAttr(resource.Attributes, "rule"),
		IsActive: boolAttr(resource.Attributes, "is-active"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
		details.EquipmentClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
		details.BusinessUnitID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["maintenance-requirement-sets"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			details.MaintenanceRequirementSetIDs = make([]string, 0, len(refs))
			for _, ref := range refs {
				details.MaintenanceRequirementSetIDs = append(details.MaintenanceRequirementSetIDs, ref.ID)
			}
		}
	}

	return details
}

func renderMaintenanceRequirementRuleDetails(cmd *cobra.Command, details maintenanceRequirementRuleDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Rule != "" {
		fmt.Fprintf(out, "Rule: %s\n", details.Rule)
	}
	fmt.Fprintf(out, "Is Active: %t\n", details.IsActive)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.EquipmentID != "" {
		fmt.Fprintf(out, "Equipment ID: %s\n", details.EquipmentID)
	}
	if details.EquipmentClassificationID != "" {
		fmt.Fprintf(out, "Equipment Classification ID: %s\n", details.EquipmentClassificationID)
	}
	if details.BusinessUnitID != "" {
		fmt.Fprintf(out, "Business Unit ID: %s\n", details.BusinessUnitID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if len(details.MaintenanceRequirementSetIDs) > 0 {
		fmt.Fprintf(out, "Maintenance Requirement Set IDs: %s\n", strings.Join(details.MaintenanceRequirementSetIDs, ", "))
	}

	return nil
}
