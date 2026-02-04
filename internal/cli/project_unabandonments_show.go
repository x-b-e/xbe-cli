package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectUnabandonmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectUnabandonmentDetails struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newProjectUnabandonmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project unabandonment details",
		Long: `Show full details of a project unabandonment.

Output Fields:
  ID       Unabandonment identifier
  Project  Project ID
  Comment  Comment (if provided)

Arguments:
  <id>    Project unabandonment ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project unabandonment
  xbe view project-unabandonments show 123

  # JSON output
  xbe view project-unabandonments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectUnabandonmentsShow,
	}
	initProjectUnabandonmentsShowFlags(cmd)
	return cmd
}

func init() {
	projectUnabandonmentsCmd.AddCommand(newProjectUnabandonmentsShowCmd())
}

func initProjectUnabandonmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectUnabandonmentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectUnabandonmentsShowOptions(cmd)
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
		return fmt.Errorf("project unabandonment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-unabandonments]", "project,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/project-unabandonments/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderProjectUnabandonmentsShowUnavailable(cmd, opts.JSON)
		}
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectUnabandonmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectUnabandonmentDetails(cmd, details)
}

func renderProjectUnabandonmentsShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), projectUnabandonmentDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Project unabandonments are write-only; show is not available.")
	return nil
}

func parseProjectUnabandonmentsShowOptions(cmd *cobra.Command) (projectUnabandonmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectUnabandonmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectUnabandonmentDetails(resp jsonAPISingleResponse) projectUnabandonmentDetails {
	attrs := resp.Data.Attributes
	details := projectUnabandonmentDetails{
		ID:      resp.Data.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}

	return details
}

func renderProjectUnabandonmentDetails(cmd *cobra.Command, details projectUnabandonmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project: %s\n", details.ProjectID)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
