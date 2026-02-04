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

type projectTransportPlanEventLocationPredictionAutopsiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanEventLocationPredictionAutopsyDetails struct {
	ID                          string `json:"id"`
	Status                      string `json:"status,omitempty"`
	Error                       string `json:"error,omitempty"`
	ProjectTransportPlanEventID string `json:"project_transport_plan_event_id,omitempty"`
	AutopsyContext              any    `json:"autopsy_context,omitempty"`
	LLMOutput                   any    `json:"llm_output,omitempty"`
}

func newProjectTransportPlanEventLocationPredictionAutopsiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan event location prediction autopsy details",
		Long: `Show the full details of a project transport plan event location prediction autopsy.

Output Fields:
  ID
  Status
  Error
  Project Transport Plan Event ID
  Autopsy Context
  LLM Output

Arguments:
  <id>    The autopsy ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an autopsy
  xbe view project-transport-plan-event-location-prediction-autopsies show 123

  # Output as JSON
  xbe view project-transport-plan-event-location-prediction-autopsies show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanEventLocationPredictionAutopsiesShow,
	}
	initProjectTransportPlanEventLocationPredictionAutopsiesShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventLocationPredictionAutopsiesCmd.AddCommand(newProjectTransportPlanEventLocationPredictionAutopsiesShowCmd())
}

func initProjectTransportPlanEventLocationPredictionAutopsiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventLocationPredictionAutopsiesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanEventLocationPredictionAutopsiesShowOptions(cmd)
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
		return fmt.Errorf("project transport plan event location prediction autopsy id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-event-location-prediction-autopsies]", "status,error,autopsy-context,llm-output,project-transport-plan-event")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-event-location-prediction-autopsies/"+id, query)
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

	details := buildProjectTransportPlanEventLocationPredictionAutopsyDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanEventLocationPredictionAutopsyDetails(cmd, details)
}

func parseProjectTransportPlanEventLocationPredictionAutopsiesShowOptions(cmd *cobra.Command) (projectTransportPlanEventLocationPredictionAutopsiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventLocationPredictionAutopsiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanEventLocationPredictionAutopsyDetails(resp jsonAPISingleResponse) projectTransportPlanEventLocationPredictionAutopsyDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectTransportPlanEventLocationPredictionAutopsyDetails{
		ID:             resource.ID,
		Status:         stringAttr(attrs, "status"),
		Error:          stringAttr(attrs, "error"),
		AutopsyContext: attrs["autopsy-context"],
		LLMOutput:      attrs["llm-output"],
	}

	if rel, ok := resource.Relationships["project-transport-plan-event"]; ok && rel.Data != nil {
		details.ProjectTransportPlanEventID = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanEventLocationPredictionAutopsyDetails(cmd *cobra.Command, details projectTransportPlanEventLocationPredictionAutopsyDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Error != "" {
		fmt.Fprintf(out, "Error: %s\n", details.Error)
	}
	if details.ProjectTransportPlanEventID != "" {
		fmt.Fprintf(out, "Project Transport Plan Event ID: %s\n", details.ProjectTransportPlanEventID)
	}

	if details.AutopsyContext != nil {
		pretty := formatJSONValue(details.AutopsyContext)
		if pretty != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Autopsy Context:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, pretty)
		}
	}

	if details.LLMOutput != nil {
		pretty := formatJSONValue(details.LLMOutput)
		if pretty != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "LLM Output:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, pretty)
		}
	}

	return nil
}
