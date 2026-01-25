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

type projectTransportPlanTrailersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanTrailerDetails struct {
	ID                                    string   `json:"id"`
	Status                                string   `json:"status,omitempty"`
	WindowStartAtCached                   string   `json:"window_start_at_cached,omitempty"`
	WindowEndAtCached                     string   `json:"window_end_at_cached,omitempty"`
	ProjectTransportPlanID                string   `json:"project_transport_plan_id,omitempty"`
	SegmentStartID                        string   `json:"segment_start_id,omitempty"`
	SegmentEndID                          string   `json:"segment_end_id,omitempty"`
	TrailerID                             string   `json:"trailer_id,omitempty"`
	ProjectTransportPlanSegmentIDs        []string `json:"project_transport_plan_segment_ids,omitempty"`
	ProjectTransportPlanSegmentTrailerIDs []string `json:"project_transport_plan_segment_trailer_ids,omitempty"`
}

func newProjectTransportPlanTrailersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan trailer details",
		Long: `Show the full details of a project transport plan trailer assignment.

Output Fields:
  ID
  Status
  Window Start Cached
  Window End Cached
  Project Transport Plan ID
  Segment Start ID
  Segment End ID
  Trailer ID
  Segment IDs
  Segment Trailer IDs

Arguments:
  <id>    Trailer assignment ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show trailer assignment details
  xbe view project-transport-plan-trailers show 123

  # JSON output
  xbe view project-transport-plan-trailers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanTrailersShow,
	}
	initProjectTransportPlanTrailersShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanTrailersCmd.AddCommand(newProjectTransportPlanTrailersShowCmd())
}

func initProjectTransportPlanTrailersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanTrailersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanTrailersShowOptions(cmd)
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
		return fmt.Errorf("project transport plan trailer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-trailers]", strings.Join([]string{
		"status",
		"window-start-at-cached",
		"window-end-at-cached",
		"project-transport-plan",
		"segment-start",
		"segment-end",
		"trailer",
		"project-transport-plan-segment-trailers",
		"segments",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-trailers/"+id, query)
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

	details := buildProjectTransportPlanTrailerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanTrailerDetails(cmd, details)
}

func parseProjectTransportPlanTrailersShowOptions(cmd *cobra.Command) (projectTransportPlanTrailersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanTrailersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanTrailerDetails(resp jsonAPISingleResponse) projectTransportPlanTrailerDetails {
	attrs := resp.Data.Attributes
	details := projectTransportPlanTrailerDetails{
		ID:                  resp.Data.ID,
		Status:              stringAttr(attrs, "status"),
		WindowStartAtCached: formatDateTime(stringAttr(attrs, "window-start-at-cached")),
		WindowEndAtCached:   formatDateTime(stringAttr(attrs, "window-end-at-cached")),
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["segment-start"]; ok && rel.Data != nil {
		details.SegmentStartID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["segment-end"]; ok && rel.Data != nil {
		details.SegmentEndID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trailer"]; ok && rel.Data != nil {
		details.TrailerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["segments"]; ok {
		details.ProjectTransportPlanSegmentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segment-trailers"]; ok {
		details.ProjectTransportPlanSegmentTrailerIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectTransportPlanTrailerDetails(cmd *cobra.Command, details projectTransportPlanTrailerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.WindowStartAtCached != "" {
		fmt.Fprintf(out, "Window Start Cached: %s\n", details.WindowStartAtCached)
	}
	if details.WindowEndAtCached != "" {
		fmt.Fprintf(out, "Window End Cached: %s\n", details.WindowEndAtCached)
	}
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.SegmentStartID != "" {
		fmt.Fprintf(out, "Segment Start ID: %s\n", details.SegmentStartID)
	}
	if details.SegmentEndID != "" {
		fmt.Fprintf(out, "Segment End ID: %s\n", details.SegmentEndID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer ID: %s\n", details.TrailerID)
	}
	if len(details.ProjectTransportPlanSegmentIDs) > 0 {
		fmt.Fprintf(out, "Segment IDs: %s\n", strings.Join(details.ProjectTransportPlanSegmentIDs, ", "))
	}
	if len(details.ProjectTransportPlanSegmentTrailerIDs) > 0 {
		fmt.Fprintf(out, "Segment Trailer IDs: %s\n", strings.Join(details.ProjectTransportPlanSegmentTrailerIDs, ", "))
	}

	return nil
}
