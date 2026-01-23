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

type maintenanceRequirementSetMaintenanceRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type maintenanceRequirementSetMaintenanceRequirementDetails struct {
	ID                          string `json:"id"`
	MaintenanceRequirementSetID string `json:"maintenance_requirement_set_id,omitempty"`
	MaintenanceRequirementID    string `json:"maintenance_requirement_id,omitempty"`
}

func newMaintenanceRequirementSetMaintenanceRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement set maintenance requirement details",
		Long: `Show the full details of a maintenance requirement set maintenance requirement.

Output Fields:
  ID
  Maintenance Requirement Set ID
  Maintenance Requirement ID

Arguments:
  <id>    The record ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a record
  xbe view maintenance-requirement-set-maintenance-requirements show 123

  # JSON output
  xbe view maintenance-requirement-set-maintenance-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRequirementSetMaintenanceRequirementsShow,
	}
	initMaintenanceRequirementSetMaintenanceRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementSetMaintenanceRequirementsCmd.AddCommand(newMaintenanceRequirementSetMaintenanceRequirementsShowCmd())
}

func initMaintenanceRequirementSetMaintenanceRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementSetMaintenanceRequirementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceRequirementSetMaintenanceRequirementsShowOptions(cmd)
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
		return fmt.Errorf("maintenance requirement set maintenance requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[maintenance-requirement-set-maintenance-requirements]", "maintenance-requirement-set,maintenance-requirement")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-set-maintenance-requirements/"+id, query)
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

	details := buildMaintenanceRequirementSetMaintenanceRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaintenanceRequirementSetMaintenanceRequirementDetails(cmd, details)
}

func parseMaintenanceRequirementSetMaintenanceRequirementsShowOptions(cmd *cobra.Command) (maintenanceRequirementSetMaintenanceRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementSetMaintenanceRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaintenanceRequirementSetMaintenanceRequirementDetails(resp jsonAPISingleResponse) maintenanceRequirementSetMaintenanceRequirementDetails {
	resource := resp.Data
	details := maintenanceRequirementSetMaintenanceRequirementDetails{ID: resource.ID}

	details.MaintenanceRequirementSetID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement-set")
	details.MaintenanceRequirementID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement")

	return details
}

func renderMaintenanceRequirementSetMaintenanceRequirementDetails(cmd *cobra.Command, details maintenanceRequirementSetMaintenanceRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaintenanceRequirementSetID != "" {
		fmt.Fprintf(out, "Maintenance Requirement Set: %s\n", details.MaintenanceRequirementSetID)
	}
	if details.MaintenanceRequirementID != "" {
		fmt.Fprintf(out, "Maintenance Requirement: %s\n", details.MaintenanceRequirementID)
	}
	return nil
}
