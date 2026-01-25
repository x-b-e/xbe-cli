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

type digitalFleetTrucksListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	Broker                string
	Trucker               string
	Tractor               string
	Trailer               string
	HasTractor            string
	HasTrailer            string
	AssignedAtMin         string
	TrailerSetAtMin       string
	TrailerSetAtMax       string
	IsTrailerSetAt        string
	TractorSetAtMin       string
	TractorSetAtMax       string
	IsTractorSetAt        string
	IntegrationIdentifier string
	IsActive              string
	CreatedAtMin          string
	CreatedAtMax          string
	IsCreatedAt           string
	UpdatedAtMin          string
	UpdatedAtMax          string
	IsUpdatedAt           string
}

type digitalFleetTruckRow struct {
	ID                    string `json:"id"`
	TruckID               string `json:"truck_id,omitempty"`
	TruckNumber           string `json:"truck_number,omitempty"`
	IsActive              bool   `json:"is_active,omitempty"`
	IntegrationIdentifier string `json:"integration_identifier,omitempty"`
	TrailerSetAt          string `json:"trailer_set_at,omitempty"`
	TractorSetAt          string `json:"tractor_set_at,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	BrokerName            string `json:"broker_name,omitempty"`
	TruckerID             string `json:"trucker_id,omitempty"`
	TruckerName           string `json:"trucker_name,omitempty"`
	TrailerID             string `json:"trailer_id,omitempty"`
	TrailerNumber         string `json:"trailer_number,omitempty"`
	TractorID             string `json:"tractor_id,omitempty"`
	TractorNumber         string `json:"tractor_number,omitempty"`
}

func newDigitalFleetTrucksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List digital fleet trucks",
		Long: `List digital fleet trucks with filtering and pagination.

Output Columns:
  ID        Digital fleet truck identifier
  TRUCK #   Digital fleet truck number
  TRUCK ID  Digital fleet truck source identifier
  ACTIVE    Active status
  TRACTOR   Assigned tractor number or ID
  TRAILER   Assigned trailer number or ID
  TRUCKER   Trucker name or ID
  BROKER    Broker name or ID

Filters:
  --broker                 Filter by broker ID
  --trucker                Filter by trucker ID
  --tractor                Filter by tractor ID
  --trailer                Filter by trailer ID
  --has-tractor            Filter by tractor assignment (true/false)
  --has-trailer            Filter by trailer assignment (true/false)
  --assigned-at-min        Filter by assigned-at timestamp (ISO 8601)
  --integration-identifier Filter by integration identifier
  --is-active              Filter by active status (true/false)
  --trailer-set-at-min     Filter by minimum trailer set timestamp (ISO 8601)
  --trailer-set-at-max     Filter by maximum trailer set timestamp (ISO 8601)
  --is-trailer-set-at      Filter by has trailer set timestamp (true/false)
  --tractor-set-at-min     Filter by minimum tractor set timestamp (ISO 8601)
  --tractor-set-at-max     Filter by maximum tractor set timestamp (ISO 8601)
  --is-tractor-set-at      Filter by has tractor set timestamp (true/false)
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --is-created-at          Filter by has created-at (true/false)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)
  --is-updated-at          Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List digital fleet trucks
  xbe view digital-fleet-trucks list

  # Filter by broker
  xbe view digital-fleet-trucks list --broker 123

  # Filter by trailer assignment
  xbe view digital-fleet-trucks list --has-trailer true

  # Filter by active status
  xbe view digital-fleet-trucks list --is-active true

  # Output as JSON
  xbe view digital-fleet-trucks list --json`,
		Args: cobra.NoArgs,
		RunE: runDigitalFleetTrucksList,
	}
	initDigitalFleetTrucksListFlags(cmd)
	return cmd
}

func init() {
	digitalFleetTrucksCmd.AddCommand(newDigitalFleetTrucksListCmd())
}

func initDigitalFleetTrucksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("has-tractor", "", "Filter by tractor assignment (true/false)")
	cmd.Flags().String("has-trailer", "", "Filter by trailer assignment (true/false)")
	cmd.Flags().String("assigned-at-min", "", "Filter by assigned-at timestamp (ISO 8601)")
	cmd.Flags().String("integration-identifier", "", "Filter by integration identifier")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("trailer-set-at-min", "", "Filter by minimum trailer set timestamp (ISO 8601)")
	cmd.Flags().String("trailer-set-at-max", "", "Filter by maximum trailer set timestamp (ISO 8601)")
	cmd.Flags().String("is-trailer-set-at", "", "Filter by has trailer set timestamp (true/false)")
	cmd.Flags().String("tractor-set-at-min", "", "Filter by minimum tractor set timestamp (ISO 8601)")
	cmd.Flags().String("tractor-set-at-max", "", "Filter by maximum tractor set timestamp (ISO 8601)")
	cmd.Flags().String("is-tractor-set-at", "", "Filter by has tractor set timestamp (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDigitalFleetTrucksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDigitalFleetTrucksListOptions(cmd)
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
	query.Set("fields[digital-fleet-trucks]", strings.Join([]string{
		"truck-id",
		"truck-number",
		"is-active",
		"integration-identifier",
		"trailer-set-at",
		"tractor-set-at",
		"broker",
		"trucker",
		"tractor",
		"trailer",
	}, ","))
	query.Set("include", "broker,trucker,tractor,trailer")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[tractors]", "number")
	query.Set("fields[trailers]", "number")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[has_tractor]", opts.HasTractor)
	setFilterIfPresent(query, "filter[has_trailer]", opts.HasTrailer)
	setFilterIfPresent(query, "filter[assigned_at_min]", opts.AssignedAtMin)
	setFilterIfPresent(query, "filter[integration_identifier]", opts.IntegrationIdentifier)
	setFilterIfPresent(query, "filter[is_active]", opts.IsActive)
	setFilterIfPresent(query, "filter[trailer_set_at_min]", opts.TrailerSetAtMin)
	setFilterIfPresent(query, "filter[trailer_set_at_max]", opts.TrailerSetAtMax)
	setFilterIfPresent(query, "filter[is_trailer_set_at]", opts.IsTrailerSetAt)
	setFilterIfPresent(query, "filter[tractor_set_at_min]", opts.TractorSetAtMin)
	setFilterIfPresent(query, "filter[tractor_set_at_max]", opts.TractorSetAtMax)
	setFilterIfPresent(query, "filter[is_tractor_set_at]", opts.IsTractorSetAt)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/digital-fleet-trucks", query)
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

	rows := buildDigitalFleetTruckRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDigitalFleetTrucksTable(cmd, rows)
}

func parseDigitalFleetTrucksListOptions(cmd *cobra.Command) (digitalFleetTrucksListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	tractor, _ := cmd.Flags().GetString("tractor")
	trailer, _ := cmd.Flags().GetString("trailer")
	hasTractor, _ := cmd.Flags().GetString("has-tractor")
	hasTrailer, _ := cmd.Flags().GetString("has-trailer")
	assignedAtMin, _ := cmd.Flags().GetString("assigned-at-min")
	integrationIdentifier, _ := cmd.Flags().GetString("integration-identifier")
	isActive, _ := cmd.Flags().GetString("is-active")
	trailerSetAtMin, _ := cmd.Flags().GetString("trailer-set-at-min")
	trailerSetAtMax, _ := cmd.Flags().GetString("trailer-set-at-max")
	isTrailerSetAt, _ := cmd.Flags().GetString("is-trailer-set-at")
	tractorSetAtMin, _ := cmd.Flags().GetString("tractor-set-at-min")
	tractorSetAtMax, _ := cmd.Flags().GetString("tractor-set-at-max")
	isTractorSetAt, _ := cmd.Flags().GetString("is-tractor-set-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return digitalFleetTrucksListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		Broker:                broker,
		Trucker:               trucker,
		Tractor:               tractor,
		Trailer:               trailer,
		HasTractor:            hasTractor,
		HasTrailer:            hasTrailer,
		AssignedAtMin:         assignedAtMin,
		IntegrationIdentifier: integrationIdentifier,
		IsActive:              isActive,
		TrailerSetAtMin:       trailerSetAtMin,
		TrailerSetAtMax:       trailerSetAtMax,
		IsTrailerSetAt:        isTrailerSetAt,
		TractorSetAtMin:       tractorSetAtMin,
		TractorSetAtMax:       tractorSetAtMax,
		IsTractorSetAt:        isTractorSetAt,
		CreatedAtMin:          createdAtMin,
		CreatedAtMax:          createdAtMax,
		IsCreatedAt:           isCreatedAt,
		UpdatedAtMin:          updatedAtMin,
		UpdatedAtMax:          updatedAtMax,
		IsUpdatedAt:           isUpdatedAt,
	}, nil
}

func buildDigitalFleetTruckRows(resp jsonAPIResponse) []digitalFleetTruckRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]digitalFleetTruckRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDigitalFleetTruckRow(resource, included))
	}
	return rows
}

func digitalFleetTruckRowFromSingle(resp jsonAPISingleResponse) digitalFleetTruckRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildDigitalFleetTruckRow(resp.Data, included)
}

func buildDigitalFleetTruckRow(resource jsonAPIResource, included map[string]jsonAPIResource) digitalFleetTruckRow {
	attrs := resource.Attributes
	row := digitalFleetTruckRow{
		ID:                    resource.ID,
		TruckID:               stringAttr(attrs, "truck-id"),
		TruckNumber:           stringAttr(attrs, "truck-number"),
		IsActive:              boolAttr(attrs, "is-active"),
		IntegrationIdentifier: stringAttr(attrs, "integration-identifier"),
		TrailerSetAt:          formatDateTime(stringAttr(attrs, "trailer-set-at")),
		TractorSetAt:          formatDateTime(stringAttr(attrs, "tractor-set-at")),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = stringAttr(trucker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		row.TrailerID = rel.Data.ID
		if trailer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TrailerNumber = stringAttr(trailer.Attributes, "number")
		}
	}

	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		row.TractorID = rel.Data.ID
		if tractor, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TractorNumber = stringAttr(tractor.Attributes, "number")
		}
	}

	return row
}

func renderDigitalFleetTrucksTable(cmd *cobra.Command, rows []digitalFleetTruckRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No digital fleet trucks found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRUCK #\tTRUCK ID\tACTIVE\tTRACTOR\tTRAILER\tTRUCKER\tBROKER")
	for _, row := range rows {
		tractor := firstNonEmpty(row.TractorNumber, row.TractorID)
		trailer := firstNonEmpty(row.TrailerNumber, row.TrailerID)
		trucker := firstNonEmpty(row.TruckerName, row.TruckerID)
		broker := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.TruckNumber, 18),
			truncateString(row.TruckID, 14),
			formatBool(row.IsActive),
			truncateString(tractor, 18),
			truncateString(trailer, 18),
			truncateString(trucker, 20),
			truncateString(broker, 20),
		)
	}
	return writer.Flush()
}
