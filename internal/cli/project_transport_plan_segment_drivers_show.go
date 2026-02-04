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

type projectTransportPlanSegmentDriversShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanSegmentDriverDetails struct {
	ID                          string `json:"id"`
	ProjectTransportPlanSegment string `json:"project_transport_plan_segment_id,omitempty"`
	Driver                      string `json:"driver_id,omitempty"`
	CreatedAt                   string `json:"created_at,omitempty"`
	UpdatedAt                   string `json:"updated_at,omitempty"`
}

func newProjectTransportPlanSegmentDriversShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan segment driver details",
		Long: `Show the full details of a project transport plan segment driver.

Output Fields:
  ID
  Project Transport Plan Segment ID
  Driver (User) ID
  Created At
  Updated At

Arguments:
  <id>    The project transport plan segment driver ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan segment driver
  xbe view project-transport-plan-segment-drivers show 123

  # Output as JSON
  xbe view project-transport-plan-segment-drivers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanSegmentDriversShow,
	}
	initProjectTransportPlanSegmentDriversShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentDriversCmd.AddCommand(newProjectTransportPlanSegmentDriversShowCmd())
}

func initProjectTransportPlanSegmentDriversShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentDriversShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanSegmentDriversShowOptions(cmd)
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
		return fmt.Errorf("project transport plan segment driver id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-segment-drivers]", "created-at,updated-at,project-transport-plan-segment,driver")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segment-drivers/"+id, query)
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

	details := buildProjectTransportPlanSegmentDriverDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanSegmentDriverDetails(cmd, details)
}

func parseProjectTransportPlanSegmentDriversShowOptions(cmd *cobra.Command) (projectTransportPlanSegmentDriversShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentDriversShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanSegmentDriverDetails(resp jsonAPISingleResponse) projectTransportPlanSegmentDriverDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectTransportPlanSegmentDriverDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["project-transport-plan-segment"]; ok && rel.Data != nil {
		details.ProjectTransportPlanSegment = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		details.Driver = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanSegmentDriverDetails(cmd *cobra.Command, details projectTransportPlanSegmentDriverDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanSegment != "" {
		fmt.Fprintf(out, "Project Transport Plan Segment ID: %s\n", details.ProjectTransportPlanSegment)
	}
	if details.Driver != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.Driver)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
