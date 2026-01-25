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

type projectCancellationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectCancellationDetails struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newProjectCancellationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project cancellation details",
		Long: `Show the full details of a project cancellation.

Output Fields:
  ID         Cancellation identifier
  Project ID Project identifier
  Comment    Cancellation comment

Arguments:
  <id>    The project cancellation ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project cancellation
  xbe view project-cancellations show 123

  # Get JSON output
  xbe view project-cancellations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectCancellationsShow,
	}
	initProjectCancellationsShowFlags(cmd)
	return cmd
}

func init() {
	projectCancellationsCmd.AddCommand(newProjectCancellationsShowCmd())
}

func initProjectCancellationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectCancellationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectCancellationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project cancellation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-cancellations/"+id, nil)
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

	details := buildProjectCancellationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectCancellationDetails(cmd, details)
}

func parseProjectCancellationsShowOptions(cmd *cobra.Command) (projectCancellationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectCancellationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectCancellationDetails(resp jsonAPISingleResponse) projectCancellationDetails {
	resource := resp.Data
	details := projectCancellationDetails{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}

	return details
}

func renderProjectCancellationDetails(cmd *cobra.Command, details projectCancellationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
