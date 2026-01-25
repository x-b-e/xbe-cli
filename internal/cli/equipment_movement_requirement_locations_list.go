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

type equipmentMovementRequirementLocationsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	Broker    string
	Name      string
	Near      string
	UsedAfter string
	Q         string
}

type equipmentMovementRequirementLocationRow struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	Latitude      string `json:"latitude,omitempty"`
	Longitude     string `json:"longitude,omitempty"`
	DistanceMiles string `json:"distance_miles,omitempty"`
	BrokerID      string `json:"broker_id,omitempty"`
	BrokerName    string `json:"broker,omitempty"`
}

func newEquipmentMovementRequirementLocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement requirement locations",
		Long: `List equipment movement requirement locations with filtering and pagination.

Output Columns:
  ID        Location identifier
  NAME      Location name
  LAT       Latitude
  LNG       Longitude
  DIST MI   Distance in miles (when --near is used)
  BROKER    Broker name or ID

Filters:
  --broker      Filter by broker ID (comma-separated for multiple)
  --name        Filter by name (exact match)
  --near        Filter by proximity (lat|lng|miles)
  --used-after  Filter by last use time (ISO 8601)
  --q           Full-text search`,
		Example: `  # List locations
  xbe view equipment-movement-requirement-locations list

  # Filter by broker
  xbe view equipment-movement-requirement-locations list --broker 123

  # Filter by name
  xbe view equipment-movement-requirement-locations list --name "Main Yard"

  # Filter by proximity
  xbe view equipment-movement-requirement-locations list --near "37.7749|-122.4194|10"

  # Output as JSON
  xbe view equipment-movement-requirement-locations list --json`,
		RunE: runEquipmentMovementRequirementLocationsList,
	}
	initEquipmentMovementRequirementLocationsListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementRequirementLocationsCmd.AddCommand(newEquipmentMovementRequirementLocationsListCmd())
}

func initEquipmentMovementRequirementLocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("name", "", "Filter by name (exact match)")
	cmd.Flags().String("near", "", "Filter by proximity (lat|lng|miles)")
	cmd.Flags().String("used-after", "", "Filter by last use time (ISO 8601)")
	cmd.Flags().String("q", "", "Full-text search")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementRequirementLocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementRequirementLocationsListOptions(cmd)
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
	query.Set("fields[equipment-movement-requirement-locations]", "name,latitude,longitude,distance-from-coordinates-miles,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

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

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[near]", opts.Near)
	setFilterIfPresent(query, "filter[used-after]", opts.UsedAfter)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-requirement-locations", query)
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

	rows := buildEquipmentMovementRequirementLocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementRequirementLocationsTable(cmd, rows)
}

func parseEquipmentMovementRequirementLocationsListOptions(cmd *cobra.Command) (equipmentMovementRequirementLocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	near, _ := cmd.Flags().GetString("near")
	usedAfter, _ := cmd.Flags().GetString("used-after")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementRequirementLocationsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		Broker:    broker,
		Name:      name,
		Near:      near,
		UsedAfter: usedAfter,
		Q:         q,
	}, nil
}

func buildEquipmentMovementRequirementLocationRows(resp jsonAPIResponse) []equipmentMovementRequirementLocationRow {
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		if inc.Attributes == nil {
			continue
		}
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	rows := make([]equipmentMovementRequirementLocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentMovementRequirementLocationRow{
			ID:            resource.ID,
			Name:          stringAttr(resource.Attributes, "name"),
			Latitude:      stringAttr(resource.Attributes, "latitude"),
			Longitude:     stringAttr(resource.Attributes, "longitude"),
			DistanceMiles: stringAttr(resource.Attributes, "distance-from-coordinates-miles"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(attrs, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func buildEquipmentMovementRequirementLocationRowFromSingle(resp jsonAPISingleResponse) equipmentMovementRequirementLocationRow {
	row := equipmentMovementRequirementLocationRow{
		ID:            resp.Data.ID,
		Name:          stringAttr(resp.Data.Attributes, "name"),
		Latitude:      stringAttr(resp.Data.Attributes, "latitude"),
		Longitude:     stringAttr(resp.Data.Attributes, "longitude"),
		DistanceMiles: stringAttr(resp.Data.Attributes, "distance-from-coordinates-miles"),
	}

	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		if inc.Attributes == nil {
			continue
		}
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(attrs, "company-name")
		}
	}

	return row
}

func renderEquipmentMovementRequirementLocationsTable(cmd *cobra.Command, rows []equipmentMovementRequirementLocationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement requirement locations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tLAT\tLNG\tDIST MI\tBROKER")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" {
			broker = row.BrokerID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			row.Latitude,
			row.Longitude,
			row.DistanceMiles,
			broker,
		)
	}
	return writer.Flush()
}
