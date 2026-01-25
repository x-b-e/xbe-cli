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

type maintenanceRequirementRuleMaintenanceRequirementSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type maintenanceRequirementRuleMaintenanceRequirementSetDetails struct {
	ID                           string `json:"id"`
	MaintenanceRequirementRuleID string `json:"maintenance_requirement_rule_id,omitempty"`
	MaintenanceRequirementSetID  string `json:"maintenance_requirement_set_id,omitempty"`
}

func newMaintenanceRequirementRuleMaintenanceRequirementSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement rule maintenance requirement set details",
		Long: `Show the full details of a maintenance requirement rule maintenance requirement set.

Output Fields:
  ID                           Maintenance requirement rule maintenance requirement set identifier
  Maintenance Requirement Rule Maintenance requirement rule ID
  Maintenance Requirement Set  Maintenance requirement set ID

Arguments:
  <id>    The maintenance requirement rule maintenance requirement set ID (required). You can find IDs using the list command.`,
		Example: `  # Show a maintenance requirement rule maintenance requirement set
  xbe view maintenance-requirement-rule-maintenance-requirement-sets show 123

  # Get JSON output
  xbe view maintenance-requirement-rule-maintenance-requirement-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRequirementRuleMaintenanceRequirementSetsShow,
	}
	initMaintenanceRequirementRuleMaintenanceRequirementSetsShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementRuleMaintenanceRequirementSetsCmd.AddCommand(newMaintenanceRequirementRuleMaintenanceRequirementSetsShowCmd())
}

func initMaintenanceRequirementRuleMaintenanceRequirementSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementRuleMaintenanceRequirementSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceRequirementRuleMaintenanceRequirementSetsShowOptions(cmd)
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
		return fmt.Errorf("maintenance requirement rule maintenance requirement set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-rule-maintenance-requirement-sets/"+id, query)
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

	details := buildMaintenanceRequirementRuleMaintenanceRequirementSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaintenanceRequirementRuleMaintenanceRequirementSetDetails(cmd, details)
}

func parseMaintenanceRequirementRuleMaintenanceRequirementSetsShowOptions(cmd *cobra.Command) (maintenanceRequirementRuleMaintenanceRequirementSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementRuleMaintenanceRequirementSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaintenanceRequirementRuleMaintenanceRequirementSetDetails(resp jsonAPISingleResponse) maintenanceRequirementRuleMaintenanceRequirementSetDetails {
	resource := resp.Data
	details := maintenanceRequirementRuleMaintenanceRequirementSetDetails{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["maintenance-requirement-rule"]; ok && rel.Data != nil {
		details.MaintenanceRequirementRuleID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["maintenance-requirement-set"]; ok && rel.Data != nil {
		details.MaintenanceRequirementSetID = rel.Data.ID
	}

	return details
}

func renderMaintenanceRequirementRuleMaintenanceRequirementSetDetails(cmd *cobra.Command, details maintenanceRequirementRuleMaintenanceRequirementSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaintenanceRequirementRuleID != "" {
		fmt.Fprintf(out, "Maintenance Requirement Rule: %s\n", details.MaintenanceRequirementRuleID)
	}
	if details.MaintenanceRequirementSetID != "" {
		fmt.Fprintf(out, "Maintenance Requirement Set: %s\n", details.MaintenanceRequirementSetID)
	}

	return nil
}
