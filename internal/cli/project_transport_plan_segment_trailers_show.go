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

type projectTransportPlanSegmentTrailersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanSegmentTrailerDetails struct {
	ID                          string `json:"id"`
	ProjectTransportPlanSegment string `json:"project_transport_plan_segment_id,omitempty"`
	Trailer                     string `json:"trailer_id,omitempty"`
	CreatedAt                   string `json:"created_at,omitempty"`
	UpdatedAt                   string `json:"updated_at,omitempty"`
}

func newProjectTransportPlanSegmentTrailersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan segment trailer details",
		Long: `Show the full details of a project transport plan segment trailer.

Output Fields:
  ID
  Project Transport Plan Segment ID
  Trailer ID
  Created At
  Updated At

Arguments:
  <id>    The project transport plan segment trailer ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan segment trailer
  xbe view project-transport-plan-segment-trailers show 123

  # Output as JSON
  xbe view project-transport-plan-segment-trailers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanSegmentTrailersShow,
	}
	initProjectTransportPlanSegmentTrailersShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentTrailersCmd.AddCommand(newProjectTransportPlanSegmentTrailersShowCmd())
}

func initProjectTransportPlanSegmentTrailersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentTrailersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanSegmentTrailersShowOptions(cmd)
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
		return fmt.Errorf("project transport plan segment trailer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-segment-trailers]", "created-at,updated-at,project-transport-plan-segment,trailer")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segment-trailers/"+id, query)
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

	details := buildProjectTransportPlanSegmentTrailerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanSegmentTrailerDetails(cmd, details)
}

func parseProjectTransportPlanSegmentTrailersShowOptions(cmd *cobra.Command) (projectTransportPlanSegmentTrailersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentTrailersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanSegmentTrailerDetails(resp jsonAPISingleResponse) projectTransportPlanSegmentTrailerDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectTransportPlanSegmentTrailerDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["project-transport-plan-segment"]; ok && rel.Data != nil {
		details.ProjectTransportPlanSegment = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		details.Trailer = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanSegmentTrailerDetails(cmd *cobra.Command, details projectTransportPlanSegmentTrailerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanSegment != "" {
		fmt.Fprintf(out, "Project Transport Plan Segment ID: %s\n", details.ProjectTransportPlanSegment)
	}
	if details.Trailer != "" {
		fmt.Fprintf(out, "Trailer ID: %s\n", details.Trailer)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
