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

type doCustomerSettingsUpdateOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	ID                              string
	IsAuditingTimeCardApprovals     bool
	EnableRecapNotifications        bool
	PlanRequiresProject             bool
	PlanRequiresBusinessUnit        bool
	AutoCancelShiftsWithoutActivity bool
}

func newDoCustomerSettingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update customer settings",
		Long: `Update customer settings.

Common flags:
  --is-auditing-time-card-approvals       Enable time card approval auditing
  --enable-recap-notifications            Enable recap notifications
  --plan-requires-project                 Require project for plans
  --plan-requires-business-unit           Require business unit for plans
  --auto-cancel-shifts-without-activity   Auto-cancel shifts without activity`,
		Example: `  # Enable recap notifications
  xbe do customer-settings update 123 --enable-recap-notifications

  # Require project for plans
  xbe do customer-settings update 123 --plan-requires-project`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerSettingsUpdate,
	}
	initDoCustomerSettingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerSettingsCmd.AddCommand(newDoCustomerSettingsUpdateCmd())
}

func initDoCustomerSettingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-auditing-time-card-approvals", false, "Enable time card approval auditing")
	cmd.Flags().Bool("enable-recap-notifications", false, "Enable recap notifications")
	cmd.Flags().Bool("plan-requires-project", false, "Require project for plans")
	cmd.Flags().Bool("plan-requires-business-unit", false, "Require business unit for plans")
	cmd.Flags().Bool("auto-cancel-shifts-without-activity", false, "Auto-cancel shifts without activity")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerSettingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerSettingsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("is-auditing-time-card-approvals") {
		attributes["is-auditing-time-card-approvals"] = opts.IsAuditingTimeCardApprovals
	}
	if cmd.Flags().Changed("enable-recap-notifications") {
		attributes["enable-recap-notifications"] = opts.EnableRecapNotifications
	}
	if cmd.Flags().Changed("plan-requires-project") {
		attributes["plan-requires-project"] = opts.PlanRequiresProject
	}
	if cmd.Flags().Changed("plan-requires-business-unit") {
		attributes["plan-requires-business-unit"] = opts.PlanRequiresBusinessUnit
	}
	if cmd.Flags().Changed("auto-cancel-shifts-without-activity") {
		attributes["auto-cancel-shifts-without-activity"] = opts.AutoCancelShiftsWithoutActivity
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "customer-settings",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-settings/"+opts.ID, jsonBody)
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
		row := customerSettingRow{
			ID:                              resp.Data.ID,
			IsAuditingTimeCardApprovals:     boolAttr(resp.Data.Attributes, "is-auditing-time-card-approvals"),
			EnableRecapNotifications:        boolAttr(resp.Data.Attributes, "enable-recap-notifications"),
			PlanRequiresProject:             boolAttr(resp.Data.Attributes, "plan-requires-project"),
			PlanRequiresBusinessUnit:        boolAttr(resp.Data.Attributes, "plan-requires-business-unit"),
			AutoCancelShiftsWithoutActivity: boolAttr(resp.Data.Attributes, "auto-cancel-shifts-without-activity"),
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer settings %s\n", resp.Data.ID)
	return nil
}

func parseDoCustomerSettingsUpdateOptions(cmd *cobra.Command, args []string) (doCustomerSettingsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isAuditingTimeCardApprovals, _ := cmd.Flags().GetBool("is-auditing-time-card-approvals")
	enableRecapNotifications, _ := cmd.Flags().GetBool("enable-recap-notifications")
	planRequiresProject, _ := cmd.Flags().GetBool("plan-requires-project")
	planRequiresBusinessUnit, _ := cmd.Flags().GetBool("plan-requires-business-unit")
	autoCancelShiftsWithoutActivity, _ := cmd.Flags().GetBool("auto-cancel-shifts-without-activity")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerSettingsUpdateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		ID:                              args[0],
		IsAuditingTimeCardApprovals:     isAuditingTimeCardApprovals,
		EnableRecapNotifications:        enableRecapNotifications,
		PlanRequiresProject:             planRequiresProject,
		PlanRequiresBusinessUnit:        planRequiresBusinessUnit,
		AutoCancelShiftsWithoutActivity: autoCancelShiftsWithoutActivity,
	}, nil
}
