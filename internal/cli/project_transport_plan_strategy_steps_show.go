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

type projectTransportPlanStrategyStepsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanStrategyStepDetails struct {
	ID        string `json:"id"`
	Position  int    `json:"position"`
	Strategy  string `json:"strategy_id,omitempty"`
	EventType string `json:"event_type_id,omitempty"`
}

func newProjectTransportPlanStrategyStepsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan strategy step details",
		Long: `Show the full details of a project transport plan strategy step.

Output Fields:
  ID
  Position
  Strategy ID
  Event Type ID

Arguments:
  <id>    The project transport plan strategy step ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan strategy step
  xbe view project-transport-plan-strategy-steps show 123

  # Output as JSON
  xbe view project-transport-plan-strategy-steps show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanStrategyStepsShow,
	}
	initProjectTransportPlanStrategyStepsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStrategyStepsCmd.AddCommand(newProjectTransportPlanStrategyStepsShowCmd())
}

func initProjectTransportPlanStrategyStepsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStrategyStepsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanStrategyStepsShowOptions(cmd)
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
		return fmt.Errorf("project transport plan strategy step id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-strategy-steps]", "position,strategy,event-type")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-strategy-steps/"+id, query)
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

	details := buildProjectTransportPlanStrategyStepDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanStrategyStepDetails(cmd, details)
}

func parseProjectTransportPlanStrategyStepsShowOptions(cmd *cobra.Command) (projectTransportPlanStrategyStepsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStrategyStepsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanStrategyStepDetails(resp jsonAPISingleResponse) projectTransportPlanStrategyStepDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := projectTransportPlanStrategyStepDetails{
		ID:       resource.ID,
		Position: intAttr(attrs, "position"),
	}

	if rel, ok := resource.Relationships["strategy"]; ok && rel.Data != nil {
		details.Strategy = rel.Data.ID
	}
	if rel, ok := resource.Relationships["event-type"]; ok && rel.Data != nil {
		details.EventType = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanStrategyStepDetails(cmd *cobra.Command, details projectTransportPlanStrategyStepDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Position: %d\n", details.Position)
	if details.Strategy != "" {
		fmt.Fprintf(out, "Strategy ID: %s\n", details.Strategy)
	}
	if details.EventType != "" {
		fmt.Fprintf(out, "Event Type ID: %s\n", details.EventType)
	}

	return nil
}
