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

type doProjectMarginMatricesCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ProjectID string
}

type projectMarginMatrixRowCreate struct {
	ID            string `json:"id"`
	ProjectID     string `json:"project_id,omitempty"`
	ScenarioCount int    `json:"scenario_count,omitempty"`
}

func newDoProjectMarginMatricesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project margin matrix",
		Long: `Create a project margin matrix.

Required flags:
  --project   Project ID`,
		Example: `  # Create a project margin matrix
  xbe do project-margin-matrices create --project 123

  # JSON output
  xbe do project-margin-matrices create --project 123 --json`,
		RunE: runDoProjectMarginMatricesCreate,
	}
	initDoProjectMarginMatricesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectMarginMatricesCmd.AddCommand(newDoProjectMarginMatricesCreateCmd())
}

func initDoProjectMarginMatricesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("project")
}

func runDoProjectMarginMatricesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectMarginMatricesCreateOptions(cmd)
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
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.ProjectID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-margin-matrices",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-margin-matrices", jsonBody)
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

	row := buildProjectMarginMatrixRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.ProjectID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created project margin matrix %s for project %s\n", row.ID, row.ProjectID)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created project margin matrix %s\n", row.ID)
	return nil
}

func parseDoProjectMarginMatricesCreateOptions(cmd *cobra.Command) (doProjectMarginMatricesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectID, _ := cmd.Flags().GetString("project")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectMarginMatricesCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ProjectID: projectID,
	}, nil
}

func buildProjectMarginMatrixRowFromSingle(resp jsonAPISingleResponse) projectMarginMatrixRowCreate {
	resource := resp.Data
	row := projectMarginMatrixRowCreate{
		ID:            resource.ID,
		ScenarioCount: projectMarginMatrixScenarioCount(resource.Attributes["scenarios"]),
	}
	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	return row
}
