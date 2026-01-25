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

type doProjectSubmissionsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Project string
	Comment string
}

func newDoProjectSubmissionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Submit a project",
		Long: `Submit a project for approval.

Required flags:
  --project   Project ID (required)

Optional flags:
  --comment   Submission comment`,
		Example: `  # Submit a project
  xbe do project-submissions create --project 12345

  # Submit with a comment
  xbe do project-submissions create --project 12345 --comment "Ready for review"

  # JSON output
  xbe do project-submissions create --project 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectSubmissionsCreate,
	}
	initDoProjectSubmissionsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectSubmissionsCmd.AddCommand(newDoProjectSubmissionsCreateCmd())
}

func initDoProjectSubmissionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("comment", "", "Submission comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectSubmissionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectSubmissionsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
	}

	data := map[string]any{
		"type":          "project-submissions",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-submissions", jsonBody)
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

	row := buildProjectSubmissionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.ProjectID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created project submission %s for project %s\n", row.ID, row.ProjectID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project submission %s\n", row.ID)
	return nil
}

func parseDoProjectSubmissionsCreateOptions(cmd *cobra.Command) (doProjectSubmissionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectSubmissionsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Project: project,
		Comment: comment,
	}, nil
}

type projectSubmissionRow struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func buildProjectSubmissionRowFromSingle(resp jsonAPISingleResponse) projectSubmissionRow {
	attrs := resp.Data.Attributes
	row := projectSubmissionRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}
	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	return row
}
