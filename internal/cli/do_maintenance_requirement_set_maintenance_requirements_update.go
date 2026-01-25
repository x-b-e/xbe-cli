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

type doMaintenanceRequirementSetMaintenanceRequirementsUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
	MaintenanceRequirementSet string
	MaintenanceRequirement    string
}

func newDoMaintenanceRequirementSetMaintenanceRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a maintenance requirement set maintenance requirement",
		Long: `Update a maintenance requirement set maintenance requirement.

Optional flags:
  --maintenance-requirement-set  Maintenance requirement set ID
  --maintenance-requirement      Maintenance requirement ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Move a requirement to another set
  xbe do maintenance-requirement-set-maintenance-requirements update 123 \
    --maintenance-requirement-set 456

  # Change the requirement on a set
  xbe do maintenance-requirement-set-maintenance-requirements update 123 \
    --maintenance-requirement 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaintenanceRequirementSetMaintenanceRequirementsUpdate,
	}
	initDoMaintenanceRequirementSetMaintenanceRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementSetMaintenanceRequirementsCmd.AddCommand(newDoMaintenanceRequirementSetMaintenanceRequirementsUpdateCmd())
}

func initDoMaintenanceRequirementSetMaintenanceRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maintenance-requirement-set", "", "Maintenance requirement set ID")
	cmd.Flags().String("maintenance-requirement", "", "Maintenance requirement ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementSetMaintenanceRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaintenanceRequirementSetMaintenanceRequirementsUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("maintenance requirement set maintenance requirement id is required")
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("maintenance-requirement-set") {
		if strings.TrimSpace(opts.MaintenanceRequirementSet) == "" {
			err := fmt.Errorf("--maintenance-requirement-set cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["maintenance-requirement-set"] = map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirement-sets",
				"id":   opts.MaintenanceRequirementSet,
			},
		}
	}
	if cmd.Flags().Changed("maintenance-requirement") {
		if strings.TrimSpace(opts.MaintenanceRequirement) == "" {
			err := fmt.Errorf("--maintenance-requirement cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["maintenance-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirements",
				"id":   opts.MaintenanceRequirement,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "maintenance-requirement-set-maintenance-requirements",
			"id":            id,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/maintenance-requirement-set-maintenance-requirements/"+id, jsonBody)
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
		row := buildMaintenanceRequirementSetMaintenanceRequirementRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated maintenance requirement set maintenance requirement %s\n", resp.Data.ID)
	return nil
}

func parseDoMaintenanceRequirementSetMaintenanceRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doMaintenanceRequirementSetMaintenanceRequirementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maintenanceRequirementSet, _ := cmd.Flags().GetString("maintenance-requirement-set")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementSetMaintenanceRequirementsUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
		MaintenanceRequirementSet: maintenanceRequirementSet,
		MaintenanceRequirement:    maintenanceRequirement,
	}, nil
}
