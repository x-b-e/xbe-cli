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

type doOrganizationProjectActualsExportsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ProjectActualsExport string
	DryRun               bool
}

type organizationProjectActualsExportRow struct {
	ID                     string `json:"id"`
	ProjectActualsExportID string `json:"project_actuals_export_id,omitempty"`
	DryRun                 bool   `json:"dry_run"`
	ExportResults          any    `json:"export_results,omitempty"`
	ExportErrors           any    `json:"export_errors,omitempty"`
}

func newDoOrganizationProjectActualsExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Export organization project actuals",
		Long: `Export organization project actuals.

Required flags:
  --project-actuals-export  Project actuals export ID (required)

Optional flags:
  --dry-run                 Validate export without sending

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Export organization project actuals
  xbe do organization-project-actuals-exports create --project-actuals-export 123

  # Run export as a dry run
  xbe do organization-project-actuals-exports create --project-actuals-export 123 --dry-run

  # Output as JSON
  xbe do organization-project-actuals-exports create --project-actuals-export 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationProjectActualsExportsCreate,
	}
	initDoOrganizationProjectActualsExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationProjectActualsExportsCmd.AddCommand(newDoOrganizationProjectActualsExportsCreateCmd())
}

func initDoOrganizationProjectActualsExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-actuals-export", "", "Project actuals export ID (required)")
	cmd.Flags().Bool("dry-run", false, "Validate export without sending")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationProjectActualsExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationProjectActualsExportsCreateOptions(cmd)
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

	opts.ProjectActualsExport = strings.TrimSpace(opts.ProjectActualsExport)
	if opts.ProjectActualsExport == "" {
		err := fmt.Errorf("--project-actuals-export is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("dry-run") {
		attributes["dry-run"] = opts.DryRun
	}

	relationships := map[string]any{
		"project-actuals-export": map[string]any{
			"data": map[string]any{
				"type": "project-actuals-exports",
				"id":   opts.ProjectActualsExport,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "organization-project-actuals-exports",
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

	body, _, err := client.Post(cmd.Context(), "/v1/organization-project-actuals-exports", jsonBody)
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

	row := organizationProjectActualsExportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization project actuals export %s\n", row.ID)
	return nil
}

func organizationProjectActualsExportRowFromSingle(resp jsonAPISingleResponse) organizationProjectActualsExportRow {
	attrs := resp.Data.Attributes
	row := organizationProjectActualsExportRow{
		ID:            resp.Data.ID,
		DryRun:        boolAttr(attrs, "dry-run"),
		ExportResults: anyAttr(attrs, "export-results"),
		ExportErrors:  anyAttr(attrs, "export-errors"),
	}

	if rel, ok := resp.Data.Relationships["project-actuals-export"]; ok && rel.Data != nil {
		row.ProjectActualsExportID = rel.Data.ID
	}

	return row
}

func parseDoOrganizationProjectActualsExportsCreateOptions(cmd *cobra.Command) (doOrganizationProjectActualsExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectActualsExport, _ := cmd.Flags().GetString("project-actuals-export")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationProjectActualsExportsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ProjectActualsExport: projectActualsExport,
		DryRun:               dryRun,
	}, nil
}
