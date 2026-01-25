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

type hosDayRegulationSetsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Broker       string
	HosDay       string
	User         string
	Driver       string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type hosDayRegulationSetRow struct {
	ID                      string `json:"id"`
	BrokerID                string `json:"broker_id,omitempty"`
	HosDayID                string `json:"hos_day_id,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	RegulationSetCode       string `json:"regulation_set_code,omitempty"`
	TimeZoneID              string `json:"time_zone_id,omitempty"`
	AdverseDrivingAvailable bool   `json:"adverse_driving_available"`
	AdverseDrivingApplied   bool   `json:"adverse_driving_applied"`
}

func newHosDayRegulationSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HOS day regulation sets",
		Long: `List HOS day regulation sets with filtering and pagination.

Output Columns:
  ID                 Regulation set identifier
  BROKER             Broker ID (if present)
  HOS DAY            HOS day ID (if present)
  DRIVER             Driver (user) ID
  REG SET            Regulation set code
  TIME ZONE          Time zone identifier
  ADVERSE AVAIL      Adverse driving available (true/false)
  ADVERSE APPLIED    Adverse driving applied (true/false)

Filters:
  --broker            Filter by broker ID
  --hos-day           Filter by HOS day ID
  --user              Filter by user ID
  --driver            Filter by driver (alias for user) ID
  --created-at-min    Filter by created-at on/after (ISO 8601)
  --created-at-max    Filter by created-at on/before (ISO 8601)
  --updated-at-min    Filter by updated-at on/after (ISO 8601)
  --updated-at-max    Filter by updated-at on/before (ISO 8601)`,
		Example: `  # List HOS day regulation sets
  xbe view hos-day-regulation-sets list

  # Filter by driver
  xbe view hos-day-regulation-sets list --driver 123

  # Filter by broker
  xbe view hos-day-regulation-sets list --broker 456

  # Output as JSON
  xbe view hos-day-regulation-sets list --json`,
		RunE: runHosDayRegulationSetsList,
	}
	initHosDayRegulationSetsListFlags(cmd)
	return cmd
}

func init() {
	hosDayRegulationSetsCmd.AddCommand(newHosDayRegulationSetsListCmd())
}

func initHosDayRegulationSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order (prefix with - for descending)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("hos-day", "", "Filter by HOS day ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("driver", "", "Filter by driver (alias for user) ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosDayRegulationSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHosDayRegulationSetsListOptions(cmd)
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
	query.Set("fields[hos-day-regulation-sets]", "regulation-set-code,time-zone-id,adverse-driving-available,adverse-driving-applied,broker,hos-day,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[hos-day]", opts.HosDay)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-day-regulation-sets", query)
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

	rows := buildHosDayRegulationSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHosDayRegulationSetsTable(cmd, rows)
}

func parseHosDayRegulationSetsListOptions(cmd *cobra.Command) (hosDayRegulationSetsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	hosDay, err := cmd.Flags().GetString("hos-day")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	user, err := cmd.Flags().GetString("user")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	driver, err := cmd.Flags().GetString("driver")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	createdAtMin, err := cmd.Flags().GetString("created-at-min")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	createdAtMax, err := cmd.Flags().GetString("created-at-max")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	updatedAtMin, err := cmd.Flags().GetString("updated-at-min")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	updatedAtMax, err := cmd.Flags().GetString("updated-at-max")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return hosDayRegulationSetsListOptions{}, err
	}

	return hosDayRegulationSetsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Broker:       broker,
		HosDay:       hosDay,
		User:         user,
		Driver:       driver,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildHosDayRegulationSetRows(resp jsonAPIResponse) []hosDayRegulationSetRow {
	rows := make([]hosDayRegulationSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildHosDayRegulationSetRow(resource))
	}
	return rows
}

func buildHosDayRegulationSetRowFromSingle(resp jsonAPISingleResponse) hosDayRegulationSetRow {
	return buildHosDayRegulationSetRow(resp.Data)
}

func buildHosDayRegulationSetRow(resource jsonAPIResource) hosDayRegulationSetRow {
	attrs := resource.Attributes
	row := hosDayRegulationSetRow{
		ID:                      resource.ID,
		RegulationSetCode:       strings.TrimSpace(stringAttr(attrs, "regulation-set-code")),
		TimeZoneID:              strings.TrimSpace(stringAttr(attrs, "time-zone-id")),
		AdverseDrivingAvailable: boolAttr(attrs, "adverse-driving-available"),
		AdverseDrivingApplied:   boolAttr(attrs, "adverse-driving-applied"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		row.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}

func renderHosDayRegulationSetsTable(cmd *cobra.Command, rows []hosDayRegulationSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No HOS day regulation sets found.")
		return nil
	}

	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tBROKER\tHOS DAY\tDRIVER\tREG SET\tTIME ZONE\tADVERSE AVAIL\tADVERSE APPLIED")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%t\t%t\n",
			row.ID,
			row.BrokerID,
			row.HosDayID,
			row.UserID,
			row.RegulationSetCode,
			row.TimeZoneID,
			row.AdverseDrivingAvailable,
			row.AdverseDrivingApplied,
		)
	}

	return w.Flush()
}
