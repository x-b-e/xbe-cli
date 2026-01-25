package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectTransportLocationEventTypesListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	ProjectTransportLocation  string
	ProjectTransportEventType string
}

type projectTransportLocationEventTypeRow struct {
	ID                            string `json:"id"`
	ProjectTransportLocationID    string `json:"project_transport_location_id,omitempty"`
	ProjectTransportLocation      string `json:"project_transport_location,omitempty"`
	ProjectTransportEventTypeID   string `json:"project_transport_event_type_id,omitempty"`
	ProjectTransportEventType     string `json:"project_transport_event_type,omitempty"`
	ProjectTransportEventTypeCode string `json:"project_transport_event_type_code,omitempty"`
}

func newProjectTransportLocationEventTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport location event types",
		Long: `List project transport location event types with filtering and pagination.

Output Columns:
  ID        Location event type identifier
  LOCATION  Project transport location name or address
  EVENT     Project transport event type (code and name)

Filters:
  --project-transport-location   Filter by project transport location ID
  --project-transport-event-type Filter by project transport event type ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List location event types
  xbe view project-transport-location-event-types list

  # Filter by location
  xbe view project-transport-location-event-types list --project-transport-location 123

  # Filter by event type
  xbe view project-transport-location-event-types list --project-transport-event-type 456

  # JSON output
  xbe view project-transport-location-event-types list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportLocationEventTypesList,
	}
	initProjectTransportLocationEventTypesListFlags(cmd)
	return cmd
}

func init() {
	projectTransportLocationEventTypesCmd.AddCommand(newProjectTransportLocationEventTypesListCmd())
}

func initProjectTransportLocationEventTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-location", "", "Filter by project transport location ID")
	cmd.Flags().String("project-transport-event-type", "", "Filter by project transport event type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportLocationEventTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportLocationEventTypesListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-location-event-types]", "project-transport-location,project-transport-event-type")
	query.Set("fields[project-transport-locations]", "name,address-full")
	query.Set("fields[project-transport-event-types]", "code,name")
	query.Set("include", "project-transport-location,project-transport-event-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project-transport-location]", opts.ProjectTransportLocation)
	setFilterIfPresent(query, "filter[project-transport-event-type]", opts.ProjectTransportEventType)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-location-event-types", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildProjectTransportLocationEventTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportLocationEventTypesTable(cmd, rows)
}

func parseProjectTransportLocationEventTypesListOptions(cmd *cobra.Command) (projectTransportLocationEventTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	location, _ := cmd.Flags().GetString("project-transport-location")
	eventType, _ := cmd.Flags().GetString("project-transport-event-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportLocationEventTypesListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		ProjectTransportLocation:  location,
		ProjectTransportEventType: eventType,
	}, nil
}

func buildProjectTransportLocationEventTypeRows(resp jsonAPIResponse) []projectTransportLocationEventTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectTransportLocationEventTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportLocationEventTypeRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["project-transport-location"]; ok && rel.Data != nil {
			row.ProjectTransportLocationID = rel.Data.ID
			if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				name := stringAttr(location.Attributes, "name")
				address := stringAttr(location.Attributes, "address-full")
				row.ProjectTransportLocation = firstNonEmpty(name, address)
			}
		}

		if rel, ok := resource.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
			row.ProjectTransportEventTypeID = rel.Data.ID
			if eventType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectTransportEventType = strings.TrimSpace(stringAttr(eventType.Attributes, "name"))
				row.ProjectTransportEventTypeCode = strings.TrimSpace(stringAttr(eventType.Attributes, "code"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderProjectTransportLocationEventTypesTable(cmd *cobra.Command, rows []projectTransportLocationEventTypeRow) error {
	out := cmd.OutOrStdout()

	writer := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(writer, "ID\tLOCATION\tEVENT TYPE")

	for _, row := range rows {
		locationDisplay := firstNonEmpty(row.ProjectTransportLocation, row.ProjectTransportLocationID)
		eventTypeDisplay := formatProjectTransportEventTypeDisplay(row.ProjectTransportEventTypeCode, row.ProjectTransportEventType)
		fmt.Fprintf(writer, "%s\t%s\t%s\n", row.ID, locationDisplay, eventTypeDisplay)
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func formatProjectTransportEventTypeDisplay(code, name string) string {
	if code != "" && name != "" {
		return fmt.Sprintf("%s - %s", code, name)
	}
	return firstNonEmpty(name, code)
}
