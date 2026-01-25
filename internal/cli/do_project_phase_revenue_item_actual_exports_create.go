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

type doProjectPhaseRevenueItemActualExportsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	OrganizationFormatterID    string
	ProjectPhaseRevenueItemIDs []string
	RevenueDate                string
}

func newDoProjectPhaseRevenueItemActualExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase revenue item actual export",
		Long: `Create a project phase revenue item actual export.

Required flags:
  --organization-formatter           Organization formatter ID (required)
  --project-phase-revenue-item-ids   Project phase revenue item IDs (comma-separated or repeated) (required)

Optional flags:
  --revenue-date                     Revenue date for the export (YYYY-MM-DD)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an export with a single item
  xbe do project-phase-revenue-item-actual-exports create \
    --organization-formatter 123 \
    --project-phase-revenue-item-ids 456

  # Create an export with multiple items and a revenue date
  xbe do project-phase-revenue-item-actual-exports create \
    --organization-formatter 123 \
    --project-phase-revenue-item-ids 456,789 \
    --revenue-date 2025-01-15

  # JSON output
  xbe do project-phase-revenue-item-actual-exports create \
    --organization-formatter 123 \
    --project-phase-revenue-item-ids 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseRevenueItemActualExportsCreate,
	}
	initDoProjectPhaseRevenueItemActualExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseRevenueItemActualExportsCmd.AddCommand(newDoProjectPhaseRevenueItemActualExportsCreateCmd())
}

func initDoProjectPhaseRevenueItemActualExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-formatter", "", "Organization formatter ID (required)")
	cmd.Flags().StringSlice("project-phase-revenue-item-ids", nil, "Project phase revenue item IDs (comma-separated or repeated) (required)")
	cmd.Flags().String("revenue-date", "", "Revenue date for the export (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization-formatter")
	cmd.MarkFlagRequired("project-phase-revenue-item-ids")
}

func runDoProjectPhaseRevenueItemActualExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseRevenueItemActualExportsCreateOptions(cmd)
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

	projectPhaseRevenueItemIDs := compactStringSlice(opts.ProjectPhaseRevenueItemIDs)
	if len(projectPhaseRevenueItemIDs) == 0 {
		err := fmt.Errorf("--project-phase-revenue-item-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.OrganizationFormatterID) == "" {
		err := fmt.Errorf("--organization-formatter is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"organization-formatter": map[string]any{
			"data": map[string]any{
				"type": "organization-formatters",
				"id":   opts.OrganizationFormatterID,
			},
		},
		"project-phase-revenue-items": map[string]any{
			"data": buildRelationshipData("project-phase-revenue-items", projectPhaseRevenueItemIDs),
		},
	}

	requestData := map[string]any{
		"type":          "project-phase-revenue-item-actual-exports",
		"relationships": relationships,
	}

	revenueDate := strings.TrimSpace(opts.RevenueDate)
	if revenueDate != "" {
		requestData["attributes"] = map[string]any{
			"revenue-date": revenueDate,
		}
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-revenue-item-actual-exports", jsonBody)
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

	row := buildProjectPhaseRevenueItemActualExportRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase revenue item actual export %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseRevenueItemActualExportsCreateOptions(cmd *cobra.Command) (doProjectPhaseRevenueItemActualExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationFormatterID, _ := cmd.Flags().GetString("organization-formatter")
	projectPhaseRevenueItemIDs, _ := cmd.Flags().GetStringSlice("project-phase-revenue-item-ids")
	revenueDate, _ := cmd.Flags().GetString("revenue-date")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseRevenueItemActualExportsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		OrganizationFormatterID:    organizationFormatterID,
		ProjectPhaseRevenueItemIDs: projectPhaseRevenueItemIDs,
		RevenueDate:                revenueDate,
	}, nil
}
