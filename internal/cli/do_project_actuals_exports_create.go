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

type doProjectActualsExportsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	OrganizationFormatterID string
	JobProductionPlanIDs    []string
}

func newDoProjectActualsExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project actuals export",
		Long: `Create a project actuals export.

Required flags:
  --organization-formatter   Organization formatter ID (required)
  --job-production-plan-ids  Job production plan IDs (comma-separated or repeated) (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an export with a single job production plan
  xbe do project-actuals-exports create \
    --organization-formatter 123 \
    --job-production-plan-ids 456

  # Create an export with multiple job production plans
  xbe do project-actuals-exports create \
    --organization-formatter 123 \
    --job-production-plan-ids 456,789

  # JSON output
  xbe do project-actuals-exports create \
    --organization-formatter 123 \
    --job-production-plan-ids 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectActualsExportsCreate,
	}
	initDoProjectActualsExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectActualsExportsCmd.AddCommand(newDoProjectActualsExportsCreateCmd())
}

func initDoProjectActualsExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-formatter", "", "Organization formatter ID (required)")
	cmd.Flags().StringSlice("job-production-plan-ids", nil, "Job production plan IDs (comma-separated or repeated) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization-formatter")
	cmd.MarkFlagRequired("job-production-plan-ids")
}

func runDoProjectActualsExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectActualsExportsCreateOptions(cmd)
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

	jobProductionPlanIDs := compactStringSlice(opts.JobProductionPlanIDs)
	if len(jobProductionPlanIDs) == 0 {
		err := fmt.Errorf("--job-production-plan-ids is required")
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
		"job-production-plans": map[string]any{
			"data": buildRelationshipData("job-production-plans", jobProductionPlanIDs),
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-actuals-exports",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-actuals-exports", jsonBody)
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

	row := buildProjectActualsExportRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project actuals export %s\n", row.ID)
	return nil
}

func parseDoProjectActualsExportsCreateOptions(cmd *cobra.Command) (doProjectActualsExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationFormatterID, _ := cmd.Flags().GetString("organization-formatter")
	jobProductionPlanIDs, _ := cmd.Flags().GetStringSlice("job-production-plan-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectActualsExportsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		OrganizationFormatterID: organizationFormatterID,
		JobProductionPlanIDs:    jobProductionPlanIDs,
	}, nil
}
