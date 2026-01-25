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

type doProjectCompletionsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ProjectID string
	Comment   string
}

type projectCompletionRowCreate struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoProjectCompletionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Complete a project",
		Long: `Complete a project.

This action transitions the project status to complete. Only projects in
approved status can be completed.

Required flags:
  --project   Project ID

Optional flags:
  --comment   Comment for the completion`,
		Example: `  # Complete a project
  xbe do project-completions create --project 123 --comment "Finalized"

  # JSON output
  xbe do project-completions create --project 123 --json`,
		RunE: runDoProjectCompletionsCreate,
	}
	initDoProjectCompletionsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectCompletionsCmd.AddCommand(newDoProjectCompletionsCreateCmd())
}

func initDoProjectCompletionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("comment", "", "Comment for the completion")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("project")
}

func runDoProjectCompletionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectCompletionsCreateOptions(cmd)
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
			"type":          "project-completions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-completions", jsonBody)
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

	row := buildProjectCompletionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project completion %s\n", row.ID)
	return nil
}

func parseDoProjectCompletionsCreateOptions(cmd *cobra.Command) (doProjectCompletionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectID, _ := cmd.Flags().GetString("project")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectCompletionsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ProjectID: projectID,
		Comment:   comment,
	}, nil
}

func buildProjectCompletionRowFromSingle(resp jsonAPISingleResponse) projectCompletionRowCreate {
	resource := resp.Data
	row := projectCompletionRowCreate{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	return row
}
