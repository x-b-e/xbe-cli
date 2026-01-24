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

type projectTransportLocationsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Broker                       string
	ProjectTransportOrganization string
	NearestProjectOfficeCached   string
	Name                         string
	Q                            string
	Near                         string
	ExternalTmsCompanyID         string
	ExternalIdentificationValue  string
}

type projectTransportLocationRow struct {
	ID                               string `json:"id"`
	Name                             string `json:"name"`
	ExternalTmsCompanyID             string `json:"external_tms_company_id,omitempty"`
	GeocodingMethod                  string `json:"geocoding_method,omitempty"`
	AddressCity                      string `json:"address_city,omitempty"`
	AddressStateCode                 string `json:"address_state_code,omitempty"`
	IsActive                         bool   `json:"is_active"`
	IsValidForStop                   bool   `json:"is_valid_for_stop"`
	DistanceInMiles                  any    `json:"distance_in_miles,omitempty"`
	BrokerID                         string `json:"broker_id,omitempty"`
	BrokerName                       string `json:"broker_name,omitempty"`
	ProjectTransportOrganizationID   string `json:"project_transport_organization_id,omitempty"`
	ProjectTransportOrganizationName string `json:"project_transport_organization_name,omitempty"`
}

func newProjectTransportLocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport locations",
		Long: `List project transport locations with filtering and pagination.

Project transport locations represent pickup, delivery, and staging locations
used in transport planning.

Output Columns:
  ID          Location identifier
  NAME        Location name
  CITY        City
  STATE       State/region code
  ACTIVE      Whether the location is active
  VALID STOP  Whether the location is valid for stops
  DIST MI     Distance in miles (when using --near)
  BROKER      Broker name
  PTO         Project transport organization name

Filters:
  --broker                         Filter by broker ID
  --project-transport-organization Filter by project transport organization ID
  --nearest-project-office-cached  Filter by nearest project office ID
  --name                           Filter by name (partial match)
  --q                              General search
  --near                           Filter by proximity (lat|lng|miles)
  --external-tms-company-id        Filter by external TMS company ID
  --external-identification-value  Filter by external identification value`,
		Example: `  # List project transport locations
  xbe view project-transport-locations list

  # Filter by broker
  xbe view project-transport-locations list --broker 123

  # Search by name
  xbe view project-transport-locations list --name "North Yard"

  # Find locations near a point (latitude|longitude|radius_miles)
  xbe view project-transport-locations list --near "41.8781|-87.6298|10"

  # Output as JSON
  xbe view project-transport-locations list --json`,
		RunE: runProjectTransportLocationsList,
	}
	initProjectTransportLocationsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportLocationsCmd.AddCommand(newProjectTransportLocationsListCmd())
}

func initProjectTransportLocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("project-transport-organization", "", "Filter by project transport organization ID")
	cmd.Flags().String("nearest-project-office-cached", "", "Filter by nearest project office ID")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("q", "", "General search")
	cmd.Flags().String("near", "", "Filter by proximity (lat|lng|miles)")
	cmd.Flags().String("external-tms-company-id", "", "Filter by external TMS company ID")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportLocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportLocationsListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[project-transport-locations]", "name,external-tms-company-id,geocoding-method,address-city,address-state-code,is-active,is-valid-for-stop,distance-in-miles,broker,project-transport-organization")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[project-transport-organizations]", "name")
	query.Set("include", "broker,project-transport-organization")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[project-transport-organization]", opts.ProjectTransportOrganization)
	setFilterIfPresent(query, "filter[nearest-project-office-cached]", opts.NearestProjectOfficeCached)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[near]", opts.Near)
	setFilterIfPresent(query, "filter[external-tms-company-id]", opts.ExternalTmsCompanyID)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-locations", query)
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

	rows := buildProjectTransportLocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportLocationsTable(cmd, rows)
}

func parseProjectTransportLocationsListOptions(cmd *cobra.Command) (projectTransportLocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	projectTransportOrganization, _ := cmd.Flags().GetString("project-transport-organization")
	nearestProjectOfficeCached, _ := cmd.Flags().GetString("nearest-project-office-cached")
	name, _ := cmd.Flags().GetString("name")
	q, _ := cmd.Flags().GetString("q")
	near, _ := cmd.Flags().GetString("near")
	externalTmsCompanyID, _ := cmd.Flags().GetString("external-tms-company-id")
	externalIdentificationValue, _ := cmd.Flags().GetString("external-identification-value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportLocationsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Broker:                       broker,
		ProjectTransportOrganization: projectTransportOrganization,
		NearestProjectOfficeCached:   nearestProjectOfficeCached,
		Name:                         name,
		Q:                            q,
		Near:                         near,
		ExternalTmsCompanyID:         externalTmsCompanyID,
		ExternalIdentificationValue:  externalIdentificationValue,
	}, nil
}

func buildProjectTransportLocationRows(resp jsonAPIResponse) []projectTransportLocationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectTransportLocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportLocationRow{
			ID:                   resource.ID,
			Name:                 stringAttr(attrs, "name"),
			ExternalTmsCompanyID: stringAttr(attrs, "external-tms-company-id"),
			GeocodingMethod:      stringAttr(attrs, "geocoding-method"),
			AddressCity:          stringAttr(attrs, "address-city"),
			AddressStateCode:     stringAttr(attrs, "address-state-code"),
			IsActive:             boolAttr(attrs, "is-active"),
			IsValidForStop:       boolAttr(attrs, "is-valid-for-stop"),
			DistanceInMiles:      attrs["distance-in-miles"],
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(broker.Attributes, "company-name")
			}
		}

		if rel, ok := resource.Relationships["project-transport-organization"]; ok && rel.Data != nil {
			row.ProjectTransportOrganizationID = rel.Data.ID
			if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectTransportOrganizationName = stringAttr(org.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderProjectTransportLocationsTable(cmd *cobra.Command, rows []projectTransportLocationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport locations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tCITY\tSTATE\tACTIVE\tVALID STOP\tDIST MI\tBROKER\tPTO")
	for _, row := range rows {
		distance := formatDistanceMiles(row.DistanceInMiles)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.AddressCity, 18),
			truncateString(row.AddressStateCode, 6),
			row.IsActive,
			row.IsValidForStop,
			distance,
			truncateString(row.BrokerName, 20),
			truncateString(row.ProjectTransportOrganizationName, 20),
		)
	}
	return writer.Flush()
}

func formatDistanceMiles(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case float64:
		return fmt.Sprintf("%.2f", typed)
	case float32:
		return fmt.Sprintf("%.2f", typed)
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case string:
		return strings.TrimSpace(typed)
	default:
		return fmt.Sprintf("%v", typed)
	}
}
