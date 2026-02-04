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

type projectTransportPlanEventTimesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanEventTimeDetails struct {
	ID                          string `json:"id"`
	ProjectTransportPlanEventID string `json:"project_transport_plan_event_id,omitempty"`
	ProjectTransportPlanID      string `json:"project_transport_plan_id,omitempty"`
	ProjectID                   string `json:"project_id,omitempty"`
	BrokerID                    string `json:"broker_id,omitempty"`
	ChangedByID                 string `json:"changed_by_id,omitempty"`
	Kind                        string `json:"kind,omitempty"`
	StartAt                     string `json:"start_at,omitempty"`
	EndAt                       string `json:"end_at,omitempty"`
	At                          string `json:"at,omitempty"`
	StartAtConfidence           string `json:"start_at_confidence,omitempty"`
	EndAtConfidence             string `json:"end_at_confidence,omitempty"`
	IsAppOwned                  bool   `json:"is_app_owned"`
	IsManualOverride            bool   `json:"is_manual_override"`
}

func newProjectTransportPlanEventTimesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan event time details",
		Long: `Show the full details of a project transport plan event time.

Output Fields:
  ID
  Project Transport Plan Event ID
  Project Transport Plan ID
  Project ID
  Broker ID
  Kind
  Start At
  End At
  At (deprecated)
  Start At Confidence
  End At Confidence
  Is App Owned
  Is Manual Override
  Changed By ID
Arguments:
  <id>    The project transport plan event time ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan event time
  xbe view project-transport-plan-event-times show 123

  # JSON output
  xbe view project-transport-plan-event-times show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanEventTimesShow,
	}
	initProjectTransportPlanEventTimesShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventTimesCmd.AddCommand(newProjectTransportPlanEventTimesShowCmd())
}

func initProjectTransportPlanEventTimesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventTimesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanEventTimesShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan event time id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-event-times]", "start-at,end-at,at,kind,start-at-confidence,end-at-confidence,is-app-owned,is-manual-override,project-transport-plan-event,project-transport-plan,project,broker,changed-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-event-times/"+id, query)
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

	details := buildProjectTransportPlanEventTimeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanEventTimeDetails(cmd, details)
}

func parseProjectTransportPlanEventTimesShowOptions(cmd *cobra.Command) (projectTransportPlanEventTimesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventTimesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanEventTimeDetails(resp jsonAPISingleResponse) projectTransportPlanEventTimeDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return projectTransportPlanEventTimeDetails{
		ID:                          resource.ID,
		ProjectTransportPlanEventID: relationshipIDFromMap(resource.Relationships, "project-transport-plan-event"),
		ProjectTransportPlanID:      relationshipIDFromMap(resource.Relationships, "project-transport-plan"),
		ProjectID:                   relationshipIDFromMap(resource.Relationships, "project"),
		BrokerID:                    relationshipIDFromMap(resource.Relationships, "broker"),
		ChangedByID:                 relationshipIDFromMap(resource.Relationships, "changed-by"),
		Kind:                        stringAttr(attrs, "kind"),
		StartAt:                     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                       formatDateTime(stringAttr(attrs, "end-at")),
		At:                          formatDateTime(stringAttr(attrs, "at")),
		StartAtConfidence:           stringAttr(attrs, "start-at-confidence"),
		EndAtConfidence:             stringAttr(attrs, "end-at-confidence"),
		IsAppOwned:                  boolAttr(attrs, "is-app-owned"),
		IsManualOverride:            boolAttr(attrs, "is-manual-override"),
	}
}

func renderProjectTransportPlanEventTimeDetails(cmd *cobra.Command, details projectTransportPlanEventTimeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanEventID != "" {
		fmt.Fprintf(out, "Project Transport Plan Event ID: %s\n", details.ProjectTransportPlanEventID)
	}
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.At != "" {
		fmt.Fprintf(out, "At (deprecated): %s\n", details.At)
	}
	if details.StartAtConfidence != "" {
		fmt.Fprintf(out, "Start At Confidence: %s\n", details.StartAtConfidence)
	}
	if details.EndAtConfidence != "" {
		fmt.Fprintf(out, "End At Confidence: %s\n", details.EndAtConfidence)
	}
	fmt.Fprintf(out, "Is App Owned: %t\n", details.IsAppOwned)
	fmt.Fprintf(out, "Is Manual Override: %t\n", details.IsManualOverride)
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By ID: %s\n", details.ChangedByID)
	}

	return nil
}
