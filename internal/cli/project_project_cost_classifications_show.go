package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectProjectCostClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectProjectCostClassificationDetails struct {
	ID                          string `json:"id"`
	ProjectID                   string `json:"project_id,omitempty"`
	ProjectCostClassificationID string `json:"project_cost_classification_id,omitempty"`
	NameOverride                string `json:"name_override,omitempty"`
	CreatedAt                   string `json:"created_at,omitempty"`
	UpdatedAt                   string `json:"updated_at,omitempty"`
}

func newProjectProjectCostClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project project cost classification details",
		Long: `Show the full details of a project project cost classification.

Output Fields:
  ID
  Project ID
  Project Cost Classification ID
  Name Override
  Created At
  Updated At

Arguments:
  <id>    The project project cost classification ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project project cost classification
  xbe view project-project-cost-classifications show 123

  # Output as JSON
  xbe view project-project-cost-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectProjectCostClassificationsShow,
	}
	initProjectProjectCostClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	projectProjectCostClassificationsCmd.AddCommand(newProjectProjectCostClassificationsShowCmd())
}

func initProjectProjectCostClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectProjectCostClassificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectProjectCostClassificationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project project cost classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-project-cost-classifications]", "name-override,created-at,updated-at,project,project-cost-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/project-project-cost-classifications/"+id, query)
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

	details := buildProjectProjectCostClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectProjectCostClassificationDetails(cmd, details)
}

func parseProjectProjectCostClassificationsShowOptions(cmd *cobra.Command) (projectProjectCostClassificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectProjectCostClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectProjectCostClassificationDetails(resp jsonAPISingleResponse) projectProjectCostClassificationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectProjectCostClassificationDetails{
		ID:           resource.ID,
		NameOverride: stringAttr(attrs, "name-override"),
		CreatedAt:    formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:    formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-cost-classification"]; ok && rel.Data != nil {
		details.ProjectCostClassificationID = rel.Data.ID
	}

	return details
}

func renderProjectProjectCostClassificationDetails(cmd *cobra.Command, details projectProjectCostClassificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.ProjectCostClassificationID != "" {
		fmt.Fprintf(out, "Project Cost Classification ID: %s\n", details.ProjectCostClassificationID)
	}
	if details.NameOverride != "" {
		fmt.Fprintf(out, "Name Override: %s\n", details.NameOverride)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
