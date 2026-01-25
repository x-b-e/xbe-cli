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

type doProjectUnabandonmentsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Project string
	Comment string
}

func newDoProjectUnabandonmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unabandon a project",
		Long: `Unabandon an abandoned project.

Required flags:
  --project   Project ID (required)

Optional flags:
  --comment   Unabandonment comment`,
		Example: `  # Unabandon a project
  xbe do project-unabandonments create --project 12345

  # Unabandon with a comment
  xbe do project-unabandonments create --project 12345 --comment "Restored"

  # JSON output
  xbe do project-unabandonments create --project 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectUnabandonmentsCreate,
	}
	initDoProjectUnabandonmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectUnabandonmentsCmd.AddCommand(newDoProjectUnabandonmentsCreateCmd())
}

func initDoProjectUnabandonmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("comment", "", "Unabandonment comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectUnabandonmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectUnabandonmentsCreateOptions(cmd)
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
		"type":          "project-unabandonments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-unabandonments", jsonBody)
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

	row := buildProjectUnabandonmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.ProjectID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created project unabandonment %s for project %s\n", row.ID, row.ProjectID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project unabandonment %s\n", row.ID)
	return nil
}

func parseDoProjectUnabandonmentsCreateOptions(cmd *cobra.Command) (doProjectUnabandonmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectUnabandonmentsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Project: project,
		Comment: comment,
	}, nil
}

type projectUnabandonmentRow struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func buildProjectUnabandonmentRowFromSingle(resp jsonAPISingleResponse) projectUnabandonmentRow {
	attrs := resp.Data.Attributes
	row := projectUnabandonmentRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}
	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	return row
}
