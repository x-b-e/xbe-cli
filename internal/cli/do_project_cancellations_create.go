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

type doProjectCancellationsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ProjectID string
	Comment   string
}

type projectCancellationRowCreate struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoProjectCancellationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Cancel a project",
		Long: `Cancel a project.

This action transitions the project status to cancelled. Only projects in
approved status can be cancelled.

Required flags:
  --project   Project ID

Optional flags:
  --comment   Comment for the cancellation`,
		Example: `  # Cancel a project
  xbe do project-cancellations create --project 123 --comment "Customer withdrew"

  # JSON output
  xbe do project-cancellations create --project 123 --json`,
		RunE: runDoProjectCancellationsCreate,
	}
	initDoProjectCancellationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectCancellationsCmd.AddCommand(newDoProjectCancellationsCreateCmd())
}

func initDoProjectCancellationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("comment", "", "Comment for the cancellation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("project")
}

func runDoProjectCancellationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectCancellationsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
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
			"type":          "project-cancellations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-cancellations", jsonBody)
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

	row := buildProjectCancellationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project cancellation %s\n", row.ID)
	return nil
}

func parseDoProjectCancellationsCreateOptions(cmd *cobra.Command) (doProjectCancellationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectID, _ := cmd.Flags().GetString("project")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectCancellationsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ProjectID: projectID,
		Comment:   comment,
	}, nil
}

func buildProjectCancellationRowFromSingle(resp jsonAPISingleResponse) projectCancellationRowCreate {
	resource := resp.Data
	row := projectCancellationRowCreate{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	return row
}
