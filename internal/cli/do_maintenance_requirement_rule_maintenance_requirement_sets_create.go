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

type doMaintenanceRequirementRuleMaintenanceRequirementSetsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	MaintenanceRequirementRuleID string
	MaintenanceRequirementSetID  string
}

func newDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a maintenance requirement rule maintenance requirement set",
		Long: `Create a maintenance requirement rule maintenance requirement set.

Required:
  --maintenance-requirement-rule  Maintenance requirement rule ID
  --maintenance-requirement-set   Maintenance requirement set ID (must be a template)`,
		Example: `  # Create a maintenance requirement rule maintenance requirement set
  xbe do maintenance-requirement-rule-maintenance-requirement-sets create --maintenance-requirement-rule 123 --maintenance-requirement-set 456`,
		RunE: runDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreate,
	}
	initDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementRuleMaintenanceRequirementSetsCmd.AddCommand(newDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreateCmd())
}

func initDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maintenance-requirement-rule", "", "Maintenance requirement rule ID")
	cmd.Flags().String("maintenance-requirement-set", "", "Maintenance requirement set ID (must be a template)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("maintenance-requirement-rule")
	_ = cmd.MarkFlagRequired("maintenance-requirement-set")
}

func runDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreateOptions(cmd)
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

	relationships := map[string]any{
		"maintenance-requirement-rule": map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirement-rules",
				"id":   opts.MaintenanceRequirementRuleID,
			},
		},
		"maintenance-requirement-set": map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirement-sets",
				"id":   opts.MaintenanceRequirementSetID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "maintenance-requirement-rule-maintenance-requirement-sets",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/maintenance-requirement-rule-maintenance-requirement-sets", jsonBody)
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

	if opts.JSON {
		row := maintenanceRequirementRuleMaintenanceRequirementSetRow{
			ID: resp.Data.ID,
		}
		if rel, ok := resp.Data.Relationships["maintenance-requirement-rule"]; ok && rel.Data != nil {
			row.MaintenanceRequirementRuleID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["maintenance-requirement-set"]; ok && rel.Data != nil {
			row.MaintenanceRequirementSetID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created maintenance requirement rule maintenance requirement set %s\n", resp.Data.ID)
	return nil
}

func parseDoMaintenanceRequirementRuleMaintenanceRequirementSetsCreateOptions(cmd *cobra.Command) (doMaintenanceRequirementRuleMaintenanceRequirementSetsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maintenanceRequirementRuleID, _ := cmd.Flags().GetString("maintenance-requirement-rule")
	maintenanceRequirementSetID, _ := cmd.Flags().GetString("maintenance-requirement-set")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementRuleMaintenanceRequirementSetsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		MaintenanceRequirementRuleID: maintenanceRequirementRuleID,
		MaintenanceRequirementSetID:  maintenanceRequirementSetID,
	}, nil
}
