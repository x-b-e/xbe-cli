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

type doMaintenanceRequirementRulesCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	Rule                    string
	IsActive                bool
	Broker                  string
	Equipment               string
	EquipmentClassification string
	BusinessUnit            string
}

func newDoMaintenanceRequirementRulesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a maintenance requirement rule",
		Long: `Create a maintenance requirement rule.

Required flags:
  --rule                     Rule text (required)
  --broker                   Broker ID (required)

Scope flags (at least one required):
  --equipment                Equipment ID
  --equipment-classification Equipment classification ID
  --business-unit            Business unit ID

Optional flags:
  --is-active                Whether the rule is active

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an equipment classification rule
  xbe do maintenance-requirement-rules create \
    --rule "Service every 100 hours" \
    --broker 123 \
    --equipment-classification 456 \
    --is-active

  # Create an equipment-specific rule
  xbe do maintenance-requirement-rules create \
    --rule "Inspect before each shift" \
    --broker 123 \
    --equipment 789`,
		Args: cobra.NoArgs,
		RunE: runDoMaintenanceRequirementRulesCreate,
	}
	initDoMaintenanceRequirementRulesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementRulesCmd.AddCommand(newDoMaintenanceRequirementRulesCreateCmd())
}

func initDoMaintenanceRequirementRulesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rule", "", "Rule text (required)")
	cmd.Flags().Bool("is-active", false, "Whether the rule is active")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("business-unit", "", "Business unit ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementRulesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceRequirementRulesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Rule) == "" {
		err := fmt.Errorf("--rule is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Equipment == "" && opts.EquipmentClassification == "" && opts.BusinessUnit == "" {
		err := fmt.Errorf("at least one of --equipment, --equipment-classification, or --business-unit is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"rule": opts.Rule,
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}
	if opts.Equipment != "" {
		relationships["equipment"] = map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if opts.EquipmentClassification != "" {
		relationships["equipment-classification"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassification,
			},
		}
	}
	if opts.BusinessUnit != "" {
		relationships["business-unit"] = map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   opts.BusinessUnit,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "maintenance-requirement-rules",
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

	body, _, err := client.Post(cmd.Context(), "/v1/maintenance-requirement-rules", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created maintenance requirement rule %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementRulesCreateOptions(cmd *cobra.Command) (doMaintenanceRequirementRulesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rule, _ := cmd.Flags().GetString("rule")
	isActive, _ := cmd.Flags().GetBool("is-active")
	broker, _ := cmd.Flags().GetString("broker")
	equipment, _ := cmd.Flags().GetString("equipment")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementRulesCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		Rule:                    rule,
		IsActive:                isActive,
		Broker:                  broker,
		Equipment:               equipment,
		EquipmentClassification: equipmentClassification,
		BusinessUnit:            businessUnit,
	}, nil
}
