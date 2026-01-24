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

type projectTransportLocationEventTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportLocationEventTypeDetails struct {
	ID                              string `json:"id"`
	ProjectTransportLocationID      string `json:"project_transport_location_id,omitempty"`
	ProjectTransportLocationName    string `json:"project_transport_location_name,omitempty"`
	ProjectTransportLocationAddress string `json:"project_transport_location_address,omitempty"`
	ProjectTransportEventTypeID     string `json:"project_transport_event_type_id,omitempty"`
	ProjectTransportEventTypeName   string `json:"project_transport_event_type_name,omitempty"`
	ProjectTransportEventTypeCode   string `json:"project_transport_event_type_code,omitempty"`
}

func newProjectTransportLocationEventTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport location event type details",
		Long: `Show the full details of a project transport location event type.

Output Fields:
  ID                          Location event type identifier
  Project Transport Location  Location name or address
  Location Address            Full address (if available)
  Project Transport Event Type Event type name
  Event Type Code             Event type code

Arguments:
  <id>    Project transport location event type ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a location event type
  xbe view project-transport-location-event-types show 123

  # JSON output
  xbe view project-transport-location-event-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportLocationEventTypesShow,
	}
	initProjectTransportLocationEventTypesShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportLocationEventTypesCmd.AddCommand(newProjectTransportLocationEventTypesShowCmd())
}

func initProjectTransportLocationEventTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportLocationEventTypesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportLocationEventTypesShowOptions(cmd)
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
		return fmt.Errorf("project transport location event type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-location-event-types]", "project-transport-location,project-transport-event-type")
	query.Set("fields[project-transport-locations]", "name,address-full")
	query.Set("fields[project-transport-event-types]", "code,name")
	query.Set("include", "project-transport-location,project-transport-event-type")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-location-event-types/"+id, query)
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

	details := buildProjectTransportLocationEventTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportLocationEventTypeDetails(cmd, details)
}

func parseProjectTransportLocationEventTypesShowOptions(cmd *cobra.Command) (projectTransportLocationEventTypesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportLocationEventTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportLocationEventTypeDetails(resp jsonAPISingleResponse) projectTransportLocationEventTypeDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectTransportLocationEventTypeDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["project-transport-location"]; ok && rel.Data != nil {
		details.ProjectTransportLocationID = rel.Data.ID
		if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportLocationName = strings.TrimSpace(stringAttr(location.Attributes, "name"))
			details.ProjectTransportLocationAddress = strings.TrimSpace(stringAttr(location.Attributes, "address-full"))
		}
	}

	if rel, ok := resp.Data.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
		details.ProjectTransportEventTypeID = rel.Data.ID
		if eventType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportEventTypeName = strings.TrimSpace(stringAttr(eventType.Attributes, "name"))
			details.ProjectTransportEventTypeCode = strings.TrimSpace(stringAttr(eventType.Attributes, "code"))
		}
	}

	return details
}

func renderProjectTransportLocationEventTypeDetails(cmd *cobra.Command, details projectTransportLocationEventTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	locationDisplay := firstNonEmpty(details.ProjectTransportLocationName, details.ProjectTransportLocationAddress)
	writeLabelWithID(out, "Project Transport Location", locationDisplay, details.ProjectTransportLocationID)
	if details.ProjectTransportLocationName != "" && details.ProjectTransportLocationAddress != "" {
		fmt.Fprintf(out, "Location Address: %s\n", details.ProjectTransportLocationAddress)
	}

	eventTypeDisplay := firstNonEmpty(details.ProjectTransportEventTypeName, details.ProjectTransportEventTypeCode)
	writeLabelWithID(out, "Project Transport Event Type", eventTypeDisplay, details.ProjectTransportEventTypeID)
	if details.ProjectTransportEventTypeCode != "" && details.ProjectTransportEventTypeName != "" {
		fmt.Fprintf(out, "Event Type Code: %s\n", details.ProjectTransportEventTypeCode)
	}

	return nil
}
