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

type doMaintenanceRequirementRuleEvaluationClerksCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	Equipment string
}

type maintenanceRequirementRuleEvaluationClerkRow struct {
	ID          string `json:"id"`
	EquipmentID string `json:"equipment_id,omitempty"`
}

func newDoMaintenanceRequirementRuleEvaluationClerksCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Request maintenance requirement rule evaluation for equipment",
		Long: `Request evaluation of maintenance requirement rules for equipment.

Required flags:
  --equipment  Equipment ID (required)`,
		Example: `  # Evaluate maintenance requirement rules for equipment
  xbe do maintenance-requirement-rule-evaluation-clerks create --equipment 123

  # Output as JSON
  xbe do maintenance-requirement-rule-evaluation-clerks create --equipment 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaintenanceRequirementRuleEvaluationClerksCreate,
	}
	initDoMaintenanceRequirementRuleEvaluationClerksCreateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementRuleEvaluationClerksCmd.AddCommand(newDoMaintenanceRequirementRuleEvaluationClerksCreateCmd())
}

func initDoMaintenanceRequirementRuleEvaluationClerksCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment", "", "Equipment ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementRuleEvaluationClerksCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceRequirementRuleEvaluationClerksCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Equipment) == "" {
		err := fmt.Errorf("--equipment is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"equipment-id": opts.Equipment,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "maintenance-requirement-rule-evaluation-clerks",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/maintenance-requirement-rule-evaluation-clerks", jsonBody)
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

	row := maintenanceRequirementRuleEvaluationClerkRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created maintenance requirement rule evaluation clerk %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementRuleEvaluationClerksCreateOptions(cmd *cobra.Command) (doMaintenanceRequirementRuleEvaluationClerksCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipment, _ := cmd.Flags().GetString("equipment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementRuleEvaluationClerksCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		Equipment: equipment,
	}, nil
}

func maintenanceRequirementRuleEvaluationClerkRowFromSingle(resp jsonAPISingleResponse) maintenanceRequirementRuleEvaluationClerkRow {
	attrs := resp.Data.Attributes
	return maintenanceRequirementRuleEvaluationClerkRow{
		ID:          resp.Data.ID,
		EquipmentID: stringAttr(attrs, "equipment-id"),
	}
}
