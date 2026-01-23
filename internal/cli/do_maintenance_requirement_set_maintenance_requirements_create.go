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

type doMaintenanceRequirementSetMaintenanceRequirementsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	MaintenanceRequirementSet string
	MaintenanceRequirement    string
}

func newDoMaintenanceRequirementSetMaintenanceRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a maintenance requirement set maintenance requirement",
		Long: `Create a maintenance requirement set maintenance requirement.

Required flags:
  --maintenance-requirement-set  Maintenance requirement set ID (required)
  --maintenance-requirement      Maintenance requirement ID (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Link a maintenance requirement to a set
  xbe do maintenance-requirement-set-maintenance-requirements create \
    --maintenance-requirement-set 123 \
    --maintenance-requirement 456`,
		Args: cobra.NoArgs,
		RunE: runDoMaintenanceRequirementSetMaintenanceRequirementsCreate,
	}
	initDoMaintenanceRequirementSetMaintenanceRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementSetMaintenanceRequirementsCmd.AddCommand(newDoMaintenanceRequirementSetMaintenanceRequirementsCreateCmd())
}

func initDoMaintenanceRequirementSetMaintenanceRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maintenance-requirement-set", "", "Maintenance requirement set ID (required)")
	cmd.Flags().String("maintenance-requirement", "", "Maintenance requirement ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("maintenance-requirement-set")
	_ = cmd.MarkFlagRequired("maintenance-requirement")
}

func runDoMaintenanceRequirementSetMaintenanceRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceRequirementSetMaintenanceRequirementsCreateOptions(cmd)
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

	maintenanceRequirementSetID := strings.TrimSpace(opts.MaintenanceRequirementSet)
	if maintenanceRequirementSetID == "" {
		err := fmt.Errorf("--maintenance-requirement-set is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	maintenanceRequirementID := strings.TrimSpace(opts.MaintenanceRequirement)
	if maintenanceRequirementID == "" {
		err := fmt.Errorf("--maintenance-requirement is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"maintenance-requirement-set": map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirement-sets",
				"id":   maintenanceRequirementSetID,
			},
		},
		"maintenance-requirement": map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirements",
				"id":   maintenanceRequirementID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "maintenance-requirement-set-maintenance-requirements",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/maintenance-requirement-set-maintenance-requirements", jsonBody)
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

	row := buildMaintenanceRequirementSetMaintenanceRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created maintenance requirement set maintenance requirement %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementSetMaintenanceRequirementsCreateOptions(cmd *cobra.Command) (doMaintenanceRequirementSetMaintenanceRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maintenanceRequirementSet, _ := cmd.Flags().GetString("maintenance-requirement-set")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementSetMaintenanceRequirementsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		MaintenanceRequirementSet: maintenanceRequirementSet,
		MaintenanceRequirement:    maintenanceRequirement,
	}, nil
}

func buildMaintenanceRequirementSetMaintenanceRequirementRowFromSingle(resp jsonAPISingleResponse) maintenanceRequirementSetMaintenanceRequirementRow {
	resource := resp.Data
	row := maintenanceRequirementSetMaintenanceRequirementRow{ID: resource.ID}

	row.MaintenanceRequirementSetID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement-set")
	row.MaintenanceRequirementID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement")

	return row
}
