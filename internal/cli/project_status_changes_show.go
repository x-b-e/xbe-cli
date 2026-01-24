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

type projectStatusChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectStatusChangeDetails struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	Comment     string `json:"comment,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
}

func newProjectStatusChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project status change details",
		Long: `Show the full details of a project status change.

Output Fields:
  ID
  Project ID
  Status
  Changed At
  Changed By (user ID)
  Comment

Arguments:
  <id>    The project status change ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project status change
  xbe view project-status-changes show 123

  # JSON output
  xbe view project-status-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectStatusChangesShow,
	}
	initProjectStatusChangesShowFlags(cmd)
	return cmd
}

func init() {
	projectStatusChangesCmd.AddCommand(newProjectStatusChangesShowCmd())
}

func initProjectStatusChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectStatusChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectStatusChangesShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project status change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-status-changes]", "status,changed-at,comment,project,changed-by")
	query.Set("include", "project,changed-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-status-changes/"+id, query)
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

	details := buildProjectStatusChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectStatusChangeDetails(cmd, details)
}

func parseProjectStatusChangesShowOptions(cmd *cobra.Command) (projectStatusChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectStatusChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectStatusChangeDetails(resp jsonAPISingleResponse) projectStatusChangeDetails {
	attrs := resp.Data.Attributes
	details := projectStatusChangeDetails{
		ID:        resp.Data.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
	}

	details.ProjectID = relationshipIDFromMap(resp.Data.Relationships, "project")
	details.ChangedByID = relationshipIDFromMap(resp.Data.Relationships, "changed-by")

	return details
}

func renderProjectStatusChangeDetails(cmd *cobra.Command, details projectStatusChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ChangedAt != "" {
		fmt.Fprintf(out, "Changed At: %s\n", details.ChangedAt)
	}
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
