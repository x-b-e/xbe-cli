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

type doMaintenanceEvaluationsTriggerOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	EquipmentID string
}

type triggerResult struct {
	EquipmentID string `json:"equipment_id"`
	Message     string `json:"message"`
}

func newDoMaintenanceEvaluationsTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger maintenance evaluation for equipment",
		Long: `Trigger a maintenance requirement rule evaluation for equipment.

This command initiates an evaluation of maintenance rules for the specified
equipment, which will determine if any maintenance is due.

Note: This is an admin operation that requires appropriate permissions.

Required:
  --equipment-id    The equipment ID to evaluate`,
		Example: `  # Trigger evaluation for equipment
  xbe do maintenance evaluations trigger --equipment-id 123

  # Get result as JSON
  xbe do maintenance evaluations trigger --equipment-id 123 --json`,
		RunE: runDoMaintenanceEvaluationsTrigger,
	}
	initDoMaintenanceEvaluationsTriggerFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceEvaluationsCmd.AddCommand(newDoMaintenanceEvaluationsTriggerCmd())
}

func initDoMaintenanceEvaluationsTriggerFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment-id", "", "Equipment ID to evaluate (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("equipment-id")
}

func runDoMaintenanceEvaluationsTrigger(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceEvaluationsTriggerOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication
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

	// Validate required fields
	if opts.EquipmentID == "" {
		err := fmt.Errorf("--equipment-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build request body
	requestBody := map[string]any{
		"data": map[string]any{
			"type": "maintenance-requirement-rule-evaluation-clerks",
			"attributes": map[string]any{
				"equipment-id": opts.EquipmentID,
			},
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

	result := triggerResult{
		EquipmentID: opts.EquipmentID,
		Message:     fmt.Sprintf("Triggered evaluation for equipment ID: %s", opts.EquipmentID),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintln(cmd.OutOrStdout(), result.Message)
	return nil
}

func parseDoMaintenanceEvaluationsTriggerOptions(cmd *cobra.Command) (doMaintenanceEvaluationsTriggerOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentID, _ := cmd.Flags().GetString("equipment-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceEvaluationsTriggerOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		EquipmentID: equipmentID,
	}, nil
}
