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

type jobProductionPlanLocationsListOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	NoAuth                         bool
	Limit                          int
	Offset                         int
	Sort                           string
	JobProductionPlan              string
	Segment                        string
	BrokerTenderJobScheduleShiftID string
}

type jobProductionPlanLocationRow struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	SiteKind             string `json:"site_kind,omitempty"`
	IsStartSiteCandidate bool   `json:"is_start_site_candidate"`
	JobProductionPlanID  string `json:"job_production_plan_id,omitempty"`
	SegmentID            string `json:"segment_id,omitempty"`
	Address              string `json:"address,omitempty"`
	IsAddressFormatted   bool   `json:"is_address_formatted_address"`
	AddressFormatted     string `json:"address_formatted,omitempty"`
	AddressTimeZoneID    string `json:"address_time_zone_id,omitempty"`
	AddressCity          string `json:"address_city,omitempty"`
	AddressStateCode     string `json:"address_state_code,omitempty"`
	AddressLatitude      string `json:"address_latitude,omitempty"`
	AddressLongitude     string `json:"address_longitude,omitempty"`
	AddressPlaceID       string `json:"address_place_id,omitempty"`
	AddressPlusCode      string `json:"address_plus_code,omitempty"`
	SkipAddressGeocoding bool   `json:"skip_address_geocoding"`
}

func newJobProductionPlanLocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan locations",
		Long: `List job production plan locations.

Output Columns:
  ID            Location identifier
  NAME          Location name
  KIND          Site kind (job_site, other)
  START CAND    Start site candidate flag
  JOB PLAN      Job production plan ID
  SEGMENT       Job production plan segment ID

Filters:
  --job-production-plan               Filter by job production plan ID
  --segment                           Filter by job production plan segment ID
  --broker-tender-job-schedule-shift  Filter by broker tender job schedule shift ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan locations
  xbe view job-production-plan-locations list

  # Filter by job production plan
  xbe view job-production-plan-locations list --job-production-plan 123

  # Output as JSON
  xbe view job-production-plan-locations list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanLocationsList,
	}
	initJobProductionPlanLocationsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanLocationsCmd.AddCommand(newJobProductionPlanLocationsListCmd())
}

func initJobProductionPlanLocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("segment", "", "Filter by job production plan segment ID")
	cmd.Flags().String("broker-tender-job-schedule-shift", "", "Filter by broker tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanLocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanLocationsListOptions(cmd)
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
	query.Set("fields[job-production-plan-locations]", "name,site-kind,is-start-site-candidate,job-production-plan,segment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[segment]", opts.Segment)
	setFilterIfPresent(query, "filter[broker-tender-job-schedule-shift]", opts.BrokerTenderJobScheduleShiftID)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-locations", query)
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

	rows := buildJobProductionPlanLocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanLocationsTable(cmd, rows)
}

func parseJobProductionPlanLocationsListOptions(cmd *cobra.Command) (jobProductionPlanLocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	segment, _ := cmd.Flags().GetString("segment")
	brokerTenderJobScheduleShiftID, _ := cmd.Flags().GetString("broker-tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanLocationsListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Limit:                          limit,
		Offset:                         offset,
		Sort:                           sort,
		JobProductionPlan:              jobProductionPlan,
		Segment:                        segment,
		BrokerTenderJobScheduleShiftID: brokerTenderJobScheduleShiftID,
	}, nil
}

func buildJobProductionPlanLocationRows(resp jsonAPIResponse) []jobProductionPlanLocationRow {
	rows := make([]jobProductionPlanLocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanLocationRow(resource))
	}
	return rows
}

func buildJobProductionPlanLocationRow(resource jsonAPIResource) jobProductionPlanLocationRow {
	row := jobProductionPlanLocationRow{
		ID:                   resource.ID,
		Name:                 stringAttr(resource.Attributes, "name"),
		SiteKind:             stringAttr(resource.Attributes, "site-kind"),
		IsStartSiteCandidate: boolAttr(resource.Attributes, "is-start-site-candidate"),
		Address:              stringAttr(resource.Attributes, "address"),
		IsAddressFormatted:   boolAttr(resource.Attributes, "is-address-formatted-address"),
		AddressFormatted:     stringAttr(resource.Attributes, "address-formatted"),
		AddressTimeZoneID:    stringAttr(resource.Attributes, "address-time-zone-id"),
		AddressCity:          stringAttr(resource.Attributes, "address-city"),
		AddressStateCode:     stringAttr(resource.Attributes, "address-state-code"),
		AddressLatitude:      stringAttr(resource.Attributes, "address-latitude"),
		AddressLongitude:     stringAttr(resource.Attributes, "address-longitude"),
		AddressPlaceID:       stringAttr(resource.Attributes, "address-place-id"),
		AddressPlusCode:      stringAttr(resource.Attributes, "address-plus-code"),
		SkipAddressGeocoding: boolAttr(resource.Attributes, "skip-address-geocoding"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["segment"]; ok && rel.Data != nil {
		row.SegmentID = rel.Data.ID
	}

	return row
}

func buildJobProductionPlanLocationRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanLocationRow {
	return buildJobProductionPlanLocationRow(resp.Data)
}

func renderJobProductionPlanLocationsTable(cmd *cobra.Command, rows []jobProductionPlanLocationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan locations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tKIND\tSTART CAND\tJOB PLAN\tSEGMENT")
	for _, row := range rows {
		startCandidate := "no"
		if row.IsStartSiteCandidate {
			startCandidate = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Name,
			row.SiteKind,
			startCandidate,
			row.JobProductionPlanID,
			row.SegmentID,
		)
	}
	return writer.Flush()
}
