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

type projectBidLocationsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	Project   string
	Near      string
	StateCode string
	County    string
}

type projectBidLocationRow struct {
	ID                string `json:"id"`
	Name              string `json:"name,omitempty"`
	ProjectID         string `json:"project_id,omitempty"`
	ProjectName       string `json:"project_name,omitempty"`
	BrokerID          string `json:"broker_id,omitempty"`
	BrokerName        string `json:"broker_name,omitempty"`
	StateCode         string `json:"state_code,omitempty"`
	County            string `json:"county,omitempty"`
	CentroidLatitude  string `json:"centroid_latitude,omitempty"`
	CentroidLongitude string `json:"centroid_longitude,omitempty"`
}

func newProjectBidLocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project bid locations",
		Long: `List project bid locations with filtering and pagination.

Output Columns:
  ID        Project bid location identifier
  NAME      Location name
  PROJECT   Project name or ID
  BROKER    Broker name or ID
  STATE     Address state code
  COUNTY    Address county
  LAT       Centroid latitude
  LNG       Centroid longitude

Filters:
  --project     Filter by project ID
  --near        Filter by proximity (lat|lng|miles)
  --state-code  Filter by state code
  --county      Filter by county name

Global flags (see xbe --help): --json, --base-url, --token, --no-auth, --limit, --offset, --sort`,
		Example: `  # List project bid locations
  xbe view project-bid-locations list

  # Filter by project
  xbe view project-bid-locations list --project 123

  # Filter by proximity
  xbe view project-bid-locations list --near "38.8977|-77.0365|5"

  # Output as JSON
  xbe view project-bid-locations list --json`,
		RunE: runProjectBidLocationsList,
	}
	initProjectBidLocationsListFlags(cmd)
	return cmd
}

func init() {
	projectBidLocationsCmd.AddCommand(newProjectBidLocationsListCmd())
}

func initProjectBidLocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("near", "", "Filter by proximity (lat|lng|miles)")
	cmd.Flags().String("state-code", "", "Filter by state code")
	cmd.Flags().String("county", "", "Filter by county name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectBidLocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectBidLocationsListOptions(cmd)
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
	query.Set("fields[project-bid-locations]", "name,centroid-latitude,centroid-longitude,address-state-code,address-county,project,broker")
	query.Set("fields[projects]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "project,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "name")
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[near]", opts.Near)
	setFilterIfPresent(query, "filter[state-code]", opts.StateCode)
	setFilterIfPresent(query, "filter[county]", opts.County)

	body, _, err := client.Get(cmd.Context(), "/v1/project-bid-locations", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildProjectBidLocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectBidLocationsTable(cmd, rows)
}

func parseProjectBidLocationsListOptions(cmd *cobra.Command) (projectBidLocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	near, _ := cmd.Flags().GetString("near")
	stateCode, _ := cmd.Flags().GetString("state-code")
	county, _ := cmd.Flags().GetString("county")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectBidLocationsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		Project:   project,
		Near:      near,
		StateCode: stateCode,
		County:    county,
	}, nil
}

func buildProjectBidLocationRows(resp jsonAPIResponse) []projectBidLocationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectBidLocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectBidLocationRow{
			ID:                resource.ID,
			Name:              stringAttr(resource.Attributes, "name"),
			StateCode:         stringAttr(resource.Attributes, "address-state-code"),
			County:            stringAttr(resource.Attributes, "address-county"),
			CentroidLatitude:  stringAttr(resource.Attributes, "centroid-latitude"),
			CentroidLongitude: stringAttr(resource.Attributes, "centroid-longitude"),
		}

		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
			if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectName = stringAttr(project.Attributes, "name")
			}
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderProjectBidLocationsTable(cmd *cobra.Command, rows []projectBidLocationRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tNAME\tPROJECT\tBROKER\tSTATE\tCOUNTY\tLAT\tLNG")
	for _, row := range rows {
		projectLabel := firstNonEmpty(row.ProjectName, row.ProjectID)
		brokerLabel := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Name,
			projectLabel,
			brokerLabel,
			row.StateCode,
			row.County,
			row.CentroidLatitude,
			row.CentroidLongitude,
		)
	}

	return w.Flush()
}

func projectBidLocationRowFromSingle(resp jsonAPISingleResponse) projectBidLocationRow {
	row := projectBidLocationRow{
		ID:                resp.Data.ID,
		Name:              stringAttr(resp.Data.Attributes, "name"),
		StateCode:         stringAttr(resp.Data.Attributes, "address-state-code"),
		County:            stringAttr(resp.Data.Attributes, "address-county"),
		CentroidLatitude:  stringAttr(resp.Data.Attributes, "centroid-latitude"),
		CentroidLongitude: stringAttr(resp.Data.Attributes, "centroid-longitude"),
	}
	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	return row
}
