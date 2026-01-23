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

type parkingSitesListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Broker         string
	Trucker        string
	Trailer        string
	Tractor        string
	ParkedType     string
	ParkedID       string
	ActiveStartMin string
	ActiveStartMax string
	ActiveEndMin   string
	ActiveEndMax   string
}

type parkingSiteRow struct {
	ID            string `json:"id"`
	IsActive      bool   `json:"is_active"`
	ActiveStartAt string `json:"active_start_at,omitempty"`
	ActiveEndAt   string `json:"active_end_at,omitempty"`
	ParkedType    string `json:"parked_type,omitempty"`
	ParkedID      string `json:"parked_id,omitempty"`
	BrokerID      string `json:"broker_id,omitempty"`
	TruckerID     string `json:"trucker_id,omitempty"`
	Address       string `json:"address,omitempty"`
}

func newParkingSitesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List parking sites",
		Long: `List parking sites.

Output Columns:
  ID            Parking site identifier
  ACTIVE        Whether the site is active
  PARKED        Type and ID of parked item
  BROKER        Broker ID
  TRUCKER       Trucker ID

Filters:
  --broker              Filter by broker ID
  --trucker             Filter by trucker ID
  --trailer             Filter by trailer ID
  --tractor             Filter by tractor ID
  --parked-type         Filter by parked type
  --parked-id           Filter by parked ID
  --active-start-min    Filter by minimum active start time
  --active-start-max    Filter by maximum active start time
  --active-end-min      Filter by minimum active end time
  --active-end-max      Filter by maximum active end time`,
		Example: `  # List all parking sites
  xbe view parking-sites list

  # Filter by broker
  xbe view parking-sites list --broker 123

  # Filter by trucker
  xbe view parking-sites list --trucker 456

  # Output as JSON
  xbe view parking-sites list --json`,
		RunE: runParkingSitesList,
	}
	initParkingSitesListFlags(cmd)
	return cmd
}

func init() {
	parkingSitesCmd.AddCommand(newParkingSitesListCmd())
}

func initParkingSitesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("parked-type", "", "Filter by parked type")
	cmd.Flags().String("parked-id", "", "Filter by parked ID")
	cmd.Flags().String("active-start-min", "", "Filter by minimum active start time")
	cmd.Flags().String("active-start-max", "", "Filter by maximum active start time")
	cmd.Flags().String("active-end-min", "", "Filter by minimum active end time")
	cmd.Flags().String("active-end-max", "", "Filter by maximum active end time")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runParkingSitesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseParkingSitesListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[active_start_at_min]", opts.ActiveStartMin)
	setFilterIfPresent(query, "filter[active_start_at_max]", opts.ActiveStartMax)
	setFilterIfPresent(query, "filter[active_end_at_min]", opts.ActiveEndMin)
	setFilterIfPresent(query, "filter[active_end_at_max]", opts.ActiveEndMax)

	// Handle polymorphic parked filter
	if opts.ParkedType != "" && opts.ParkedID != "" {
		query.Set("filter[parked]", opts.ParkedType+"|"+opts.ParkedID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/parking-sites", query)
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

	rows := buildParkingSiteRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderParkingSitesTable(cmd, rows)
}

func parseParkingSitesListOptions(cmd *cobra.Command) (parkingSitesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	parkedType, _ := cmd.Flags().GetString("parked-type")
	parkedID, _ := cmd.Flags().GetString("parked-id")
	activeStartMin, _ := cmd.Flags().GetString("active-start-min")
	activeStartMax, _ := cmd.Flags().GetString("active-start-max")
	activeEndMin, _ := cmd.Flags().GetString("active-end-min")
	activeEndMax, _ := cmd.Flags().GetString("active-end-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return parkingSitesListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Broker:         broker,
		Trucker:        trucker,
		Trailer:        trailer,
		Tractor:        tractor,
		ParkedType:     parkedType,
		ParkedID:       parkedID,
		ActiveStartMin: activeStartMin,
		ActiveStartMax: activeStartMax,
		ActiveEndMin:   activeEndMin,
		ActiveEndMax:   activeEndMax,
	}, nil
}

func buildParkingSiteRows(resp jsonAPIResponse) []parkingSiteRow {
	rows := make([]parkingSiteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := parkingSiteRow{
			ID:            resource.ID,
			IsActive:      boolAttr(resource.Attributes, "is-active"),
			ActiveStartAt: stringAttr(resource.Attributes, "active-start-at"),
			ActiveEndAt:   stringAttr(resource.Attributes, "active-end-at"),
			ParkedType:    stringAttr(resource.Attributes, "parked-type"),
		}

		if rel, ok := resource.Relationships["parked"]; ok && rel.Data != nil {
			row.ParkedType = rel.Data.Type
			row.ParkedID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderParkingSitesTable(cmd *cobra.Command, rows []parkingSiteRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No parking sites found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tACTIVE\tPARKED\tBROKER\tTRUCKER")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		parked := ""
		if row.ParkedType != "" && row.ParkedID != "" {
			parked = row.ParkedType + "/" + row.ParkedID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			active,
			parked,
			row.BrokerID,
			row.TruckerID,
		)
	}
	return writer.Flush()
}
