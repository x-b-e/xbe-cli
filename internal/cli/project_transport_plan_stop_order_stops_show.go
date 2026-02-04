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

type projectTransportPlanStopOrderStopsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanStopOrderStopDetails struct {
	ID                         string `json:"id"`
	ProjectTransportPlanStopID string `json:"project_transport_plan_stop_id,omitempty"`
	TransportOrderStopID       string `json:"transport_order_stop_id,omitempty"`
	ProjectTransportPlanID     string `json:"project_transport_plan_id,omitempty"`
	TransportOrderID           string `json:"transport_order_id,omitempty"`
	CreatedByID                string `json:"created_by_id,omitempty"`
	CreatedByName              string `json:"created_by_name,omitempty"`
	CreatedByEmail             string `json:"created_by_email,omitempty"`
}

func newProjectTransportPlanStopOrderStopsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan stop order stop details",
		Long: `Show the full details of a project transport plan stop order stop.

Output Fields:
  ID         Link identifier
  PLAN STOP  Project transport plan stop ID
  ORDER STOP Transport order stop ID
  PLAN       Project transport plan ID
  ORDER      Transport order ID
  CREATED BY User who created the link (if available)

Arguments:
  <id>  Project transport plan stop order stop ID (required).`,
		Example: `  # Show a project transport plan stop order stop
  xbe view project-transport-plan-stop-order-stops show 123

  # Output as JSON
  xbe view project-transport-plan-stop-order-stops show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanStopOrderStopsShow,
	}
	initProjectTransportPlanStopOrderStopsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStopOrderStopsCmd.AddCommand(newProjectTransportPlanStopOrderStopsShowCmd())
}

func initProjectTransportPlanStopOrderStopsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStopOrderStopsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanStopOrderStopsShowOptions(cmd)
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
		return fmt.Errorf("project transport plan stop order stop id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-stop-order-stops]", "project-transport-plan-stop,transport-order-stop,project-transport-plan,transport-order,created-by")
	query.Set("include", "created-by")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-stop-order-stops/"+id, query)
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

	details := buildProjectTransportPlanStopOrderStopDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanStopOrderStopDetails(cmd, details)
}

func parseProjectTransportPlanStopOrderStopsShowOptions(cmd *cobra.Command) (projectTransportPlanStopOrderStopsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStopOrderStopsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanStopOrderStopDetails(resp jsonAPISingleResponse) projectTransportPlanStopOrderStopDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectTransportPlanStopOrderStopDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan-stop"]; ok && rel.Data != nil {
		details.ProjectTransportPlanStopID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["transport-order-stop"]; ok && rel.Data != nil {
		details.TransportOrderStopID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["transport-order"]; ok && rel.Data != nil {
		details.TransportOrderID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(user.Attributes, "name")
			details.CreatedByEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return details
}

func renderProjectTransportPlanStopOrderStopDetails(cmd *cobra.Command, details projectTransportPlanStopOrderStopDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanStopID != "" {
		fmt.Fprintf(out, "Project Transport Plan Stop: %s\n", details.ProjectTransportPlanStopID)
	}
	if details.TransportOrderStopID != "" {
		fmt.Fprintf(out, "Transport Order Stop: %s\n", details.TransportOrderStopID)
	}
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan: %s\n", details.ProjectTransportPlanID)
	}
	if details.TransportOrderID != "" {
		fmt.Fprintf(out, "Transport Order: %s\n", details.TransportOrderID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By Name: %s\n", details.CreatedByName)
	}
	if details.CreatedByEmail != "" {
		fmt.Fprintf(out, "Created By Email: %s\n", details.CreatedByEmail)
	}

	return nil
}
