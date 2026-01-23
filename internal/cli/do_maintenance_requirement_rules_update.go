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

type doMaintenanceRequirementRulesUpdateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ID                      string
	Rule                    string
	IsActive                bool
	Broker                  string
	Equipment               string
	EquipmentClassification string
	BusinessUnit            string
}

func newDoMaintenanceRequirementRulesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a maintenance requirement rule",
		Long: `Update a maintenance requirement rule.

Optional flags:
  --rule                     Rule text
  --is-active                Whether the rule is active
  --broker                   Broker ID
  --equipment                Equipment ID
  --equipment-classification Equipment classification ID
  --business-unit            Business unit ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update rule text
  xbe do maintenance-requirement-rules update 123 --rule "Updated rule"

  # Disable a rule
  xbe do maintenance-requirement-rules update 123 --is-active=false

  # Update scope relationships
  xbe do maintenance-requirement-rules update 123 --equipment 456 --business-unit 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaintenanceRequirementRulesUpdate,
	}
	initDoMaintenanceRequirementRulesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementRulesCmd.AddCommand(newDoMaintenanceRequirementRulesUpdateCmd())
}

func initDoMaintenanceRequirementRulesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rule", "", "Rule text")
	cmd.Flags().Bool("is-active", false, "Whether the rule is active")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("business-unit", "", "Business unit ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementRulesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaintenanceRequirementRulesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("rule") {
		attributes["rule"] = opts.Rule
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}

	if cmd.Flags().Changed("broker") {
		if opts.Broker == "" {
			err := fmt.Errorf("--broker cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if cmd.Flags().Changed("equipment") {
		if opts.Equipment == "" {
			relationships["equipment"] = map[string]any{"data": nil}
		} else {
			relationships["equipment"] = map[string]any{
				"data": map[string]any{
					"type": "equipment",
					"id":   opts.Equipment,
				},
			}
		}
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
	if cmd.Flags().Changed("business-unit") {
		if opts.BusinessUnit == "" {
			relationships["business-unit"] = map[string]any{"data": nil}
		} else {
			relationships["business-unit"] = map[string]any{
				"data": map[string]any{
					"type": "business-units",
					"id":   opts.BusinessUnit,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "maintenance-requirement-rules",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/maintenance-requirement-rules/"+opts.ID, jsonBody)
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

	row := buildMaintenanceRequirementRuleRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated maintenance requirement rule %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementRulesUpdateOptions(cmd *cobra.Command, args []string) (doMaintenanceRequirementRulesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rule, _ := cmd.Flags().GetString("rule")
	isActive, _ := cmd.Flags().GetBool("is-active")
	broker, _ := cmd.Flags().GetString("broker")
	equipment, _ := cmd.Flags().GetString("equipment")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementRulesUpdateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ID:                      args[0],
		Rule:                    rule,
		IsActive:                isActive,
		Broker:                  broker,
		Equipment:               equipment,
		EquipmentClassification: equipmentClassification,
		BusinessUnit:            businessUnit,
	}, nil
}
