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

type projectAbandonmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectAbandonmentDetails struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newProjectAbandonmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project abandonment details",
		Long: `Show the full details of a project abandonment.

Output Fields:
  ID         Abandonment identifier
  Project ID Project identifier
  Comment    Abandonment comment

Arguments:
  <id>    The project abandonment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project abandonment
  xbe view project-abandonments show 123

  # Get JSON output
  xbe view project-abandonments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectAbandonmentsShow,
	}
	initProjectAbandonmentsShowFlags(cmd)
	return cmd
}

func init() {
	projectAbandonmentsCmd.AddCommand(newProjectAbandonmentsShowCmd())
}

func initProjectAbandonmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectAbandonmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectAbandonmentsShowOptions(cmd)
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
		return fmt.Errorf("project abandonment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-abandonments/"+id, nil)
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

	details := buildProjectAbandonmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectAbandonmentDetails(cmd, details)
}

func parseProjectAbandonmentsShowOptions(cmd *cobra.Command) (projectAbandonmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectAbandonmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectAbandonmentDetails(resp jsonAPISingleResponse) projectAbandonmentDetails {
	resource := resp.Data
	details := projectAbandonmentDetails{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}

	return details
}

func renderProjectAbandonmentDetails(cmd *cobra.Command, details projectAbandonmentDetails) error {
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
