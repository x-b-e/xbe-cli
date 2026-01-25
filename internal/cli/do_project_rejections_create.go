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

type doProjectRejectionsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Project string
	Comment string
}

type projectRejectionRow struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoProjectRejectionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Reject a project",
		Long: `Reject a project.

Required flags:
  --project  Project ID (required)

Optional flags:
  --comment  Rejection comment`,
		Example: `  # Reject a project
  xbe do project-rejections create --project 123 --comment "Not ready"`,
		Args: cobra.NoArgs,
		RunE: runDoProjectRejectionsCreate,
	}
	initDoProjectRejectionsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectRejectionsCmd.AddCommand(newDoProjectRejectionsCreateCmd())
}

func initDoProjectRejectionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("comment", "", "Rejection comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectRejectionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectRejectionsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-rejections",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-rejections", jsonBody)
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

	row := projectRejectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project rejection %s\n", row.ID)
	return nil
}

func projectRejectionRowFromSingle(resp jsonAPISingleResponse) projectRejectionRow {
	attrs := resp.Data.Attributes
	row := projectRejectionRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}

	return row
}

func parseDoProjectRejectionsCreateOptions(cmd *cobra.Command) (doProjectRejectionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRejectionsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Project: project,
		Comment: comment,
	}, nil
}
