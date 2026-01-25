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

type projectTrailerClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTrailerClassificationDetails struct {
	ID                           string `json:"id"`
	ProjectID                    string `json:"project_id,omitempty"`
	TrailerClassificationID      string `json:"trailer_classification_id,omitempty"`
	ProjectLaborClassificationID string `json:"project_labor_classification_id,omitempty"`
	CreatedAt                    string `json:"created_at,omitempty"`
	UpdatedAt                    string `json:"updated_at,omitempty"`
}

func newProjectTrailerClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project trailer classification details",
		Long: `Show the full details of a project trailer classification.

Output Fields:
  ID
  Project ID
  Trailer Classification ID
  Project Labor Classification ID
  Created At
  Updated At

Arguments:
  <id>    The project trailer classification ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project trailer classification
  xbe view project-trailer-classifications show 123

  # Get JSON output
  xbe view project-trailer-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTrailerClassificationsShow,
	}
	initProjectTrailerClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	projectTrailerClassificationsCmd.AddCommand(newProjectTrailerClassificationsShowCmd())
}

func initProjectTrailerClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTrailerClassificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTrailerClassificationsShowOptions(cmd)
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
		return fmt.Errorf("project trailer classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-trailer-classifications/"+id, nil)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectTrailerClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTrailerClassificationDetails(cmd, details)
}

func parseProjectTrailerClassificationsShowOptions(cmd *cobra.Command) (projectTrailerClassificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTrailerClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTrailerClassificationDetails(resp jsonAPISingleResponse) projectTrailerClassificationDetails {
	attrs := resp.Data.Attributes
	details := projectTrailerClassificationDetails{
		ID:        resp.Data.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trailer-classification"]; ok && rel.Data != nil {
		details.TrailerClassificationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-labor-classification"]; ok && rel.Data != nil {
		details.ProjectLaborClassificationID = rel.Data.ID
	}

	return details
}

func renderProjectTrailerClassificationDetails(cmd *cobra.Command, details projectTrailerClassificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.TrailerClassificationID != "" {
		fmt.Fprintf(out, "Trailer Classification ID: %s\n", details.TrailerClassificationID)
	}
	if details.ProjectLaborClassificationID != "" {
		fmt.Fprintf(out, "Project Labor Classification ID: %s\n", details.ProjectLaborClassificationID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
