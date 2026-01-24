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

type doProjectProjectCostClassificationsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	Project                   string
	ProjectCostClassification string
	NameOverride              string
}

func newDoProjectProjectCostClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project project cost classification",
		Long: `Create a project project cost classification.

Required flags:
  --project                      Project ID
  --project-cost-classification  Project cost classification ID

Optional flags:
  --name-override                Override the classification name for this project

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project project cost classification
  xbe do project-project-cost-classifications create --project 123 --project-cost-classification 456

  # Create with name override
  xbe do project-project-cost-classifications create --project 123 --project-cost-classification 456 --name-override "Custom Name"`,
		Args: cobra.NoArgs,
		RunE: runDoProjectProjectCostClassificationsCreate,
	}
	initDoProjectProjectCostClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectProjectCostClassificationsCmd.AddCommand(newDoProjectProjectCostClassificationsCreateCmd())
}

func initDoProjectProjectCostClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("project-cost-classification", "", "Project cost classification ID")
	cmd.Flags().String("name-override", "", "Override the classification name for this project")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectProjectCostClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectProjectCostClassificationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.ProjectCostClassification) == "" {
		err := fmt.Errorf("--project-cost-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("name-override") {
		attributes["name-override"] = opts.NameOverride
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
		"project-cost-classification": map[string]any{
			"data": map[string]any{
				"type": "project-cost-classifications",
				"id":   opts.ProjectCostClassification,
			},
		},
	}

	data := map[string]any{
		"type":          "project-project-cost-classifications",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-project-cost-classifications", jsonBody)
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

	row := buildProjectProjectCostClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project project cost classification %s\n", row.ID)
	return nil
}

func parseDoProjectProjectCostClassificationsCreateOptions(cmd *cobra.Command) (doProjectProjectCostClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	nameOverride, _ := cmd.Flags().GetString("name-override")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectProjectCostClassificationsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		Project:                   project,
		ProjectCostClassification: projectCostClassification,
		NameOverride:              nameOverride,
	}, nil
}

func buildProjectProjectCostClassificationRowFromSingle(resp jsonAPISingleResponse) projectProjectCostClassificationRow {
	attrs := resp.Data.Attributes

	row := projectProjectCostClassificationRow{
		ID:           resp.Data.ID,
		NameOverride: stringAttr(attrs, "name-override"),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-cost-classification"]; ok && rel.Data != nil {
		row.ProjectCostClassificationID = rel.Data.ID
	}

	return row
}
