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

type doProjectTrailerClassificationsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	Project                    string
	TrailerClassification      string
	ProjectLaborClassification string
}

func newDoProjectTrailerClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project trailer classification",
		Long: `Create a project trailer classification.

Required flags:
  --project                 Project ID (required)
  --trailer-classification  Trailer classification ID (required)

Optional flags:
  --project-labor-classification  Project labor classification ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project trailer classification
  xbe do project-trailer-classifications create --project 123 --trailer-classification 456

  # Link a project labor classification
  xbe do project-trailer-classifications create \
    --project 123 \
    --trailer-classification 456 \
    --project-labor-classification 789

  # Get JSON output
  xbe do project-trailer-classifications create --project 123 --trailer-classification 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTrailerClassificationsCreate,
	}
	initDoProjectTrailerClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTrailerClassificationsCmd.AddCommand(newDoProjectTrailerClassificationsCreateCmd())
}

func initDoProjectTrailerClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID (required)")
	cmd.Flags().String("project-labor-classification", "", "Project labor classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTrailerClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTrailerClassificationsCreateOptions(cmd)
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

	projectID := strings.TrimSpace(opts.Project)
	if projectID == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	trailerClassificationID := strings.TrimSpace(opts.TrailerClassification)
	if trailerClassificationID == "" {
		err := fmt.Errorf("--trailer-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   projectID,
			},
		},
		"trailer-classification": map[string]any{
			"data": map[string]any{
				"type": "trailer-classifications",
				"id":   trailerClassificationID,
			},
		},
	}

	if strings.TrimSpace(opts.ProjectLaborClassification) != "" {
		relationships["project-labor-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-labor-classifications",
				"id":   opts.ProjectLaborClassification,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-trailer-classifications",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-trailer-classifications", jsonBody)
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

	details := buildProjectTrailerClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project trailer classification %s\n", details.ID)
	return nil
}

func parseDoProjectTrailerClassificationsCreateOptions(cmd *cobra.Command) (doProjectTrailerClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	projectLaborClassification, _ := cmd.Flags().GetString("project-labor-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTrailerClassificationsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		Project:                    project,
		TrailerClassification:      trailerClassification,
		ProjectLaborClassification: projectLaborClassification,
	}, nil
}
