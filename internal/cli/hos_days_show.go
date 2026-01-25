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

type hosDaysShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type hosDayDetails struct {
	ID                              string   `json:"id"`
	DriverID                        string   `json:"driver_id,omitempty"`
	BrokerID                        string   `json:"broker_id,omitempty"`
	LatestHosAvailabilitySnapshotID string   `json:"latest_hos_availability_snapshot_id,omitempty"`
	HosAvailabilitySnapshotIDs      []string `json:"hos_availability_snapshot_ids,omitempty"`
	HosEventIDs                     []string `json:"hos_event_ids,omitempty"`
	ServiceDate                     string   `json:"service_date,omitempty"`
	RegulationSetCode               string   `json:"regulation_set_code,omitempty"`
	TimeZoneID                      string   `json:"time_zone_id,omitempty"`
}

func newHosDaysShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show HOS day details",
		Long: `Show the full details of a specific HOS day.

Output Fields:
  ID
  Driver
  Broker
  Service Date
  Regulation Set Code
  Time Zone ID
  Latest HOS Availability Snapshot
  HOS Availability Snapshots
  HOS Events

Arguments:
  <id>  HOS day ID (required). Find IDs using the list command.`,
		Example: `  # View an HOS day by ID
  xbe view hos-days show 123

  # Get JSON output
  xbe view hos-days show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHosDaysShow,
	}
	initHosDaysShowFlags(cmd)
	return cmd
}

func init() {
	hosDaysCmd.AddCommand(newHosDaysShowCmd())
}

func initHosDaysShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosDaysShow(cmd *cobra.Command, args []string) error {
	opts, err := parseHosDaysShowOptions(cmd)
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
		return fmt.Errorf("hos day id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[hos-days]", "service-date,regulation-set-code,time-zone-id,driver,broker,latest-hos-availability-snapshot,hos-availability-snapshots,hos-events")

	body, _, err := client.Get(cmd.Context(), "/v1/hos-days/"+id, query)
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

	details := buildHosDayDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHosDayDetails(cmd, details)
}

func parseHosDaysShowOptions(cmd *cobra.Command) (hosDaysShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosDaysShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHosDayDetails(resp jsonAPISingleResponse) hosDayDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := hosDayDetails{
		ID:                resource.ID,
		ServiceDate:       formatDate(stringAttr(attrs, "service-date")),
		RegulationSetCode: stringAttr(attrs, "regulation-set-code"),
		TimeZoneID:        stringAttr(attrs, "time-zone-id"),
	}

	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["latest-hos-availability-snapshot"]; ok && rel.Data != nil {
		details.LatestHosAvailabilitySnapshotID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-availability-snapshots"]; ok {
		details.HosAvailabilitySnapshotIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["hos-events"]; ok {
		details.HosEventIDs = relationshipIDList(rel)
	}

	return details
}

func renderHosDayDetails(cmd *cobra.Command, details hosDayDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver: %s\n", details.DriverID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.ServiceDate != "" {
		fmt.Fprintf(out, "Service Date: %s\n", details.ServiceDate)
	}
	if details.RegulationSetCode != "" {
		fmt.Fprintf(out, "Regulation Set Code: %s\n", details.RegulationSetCode)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	if details.LatestHosAvailabilitySnapshotID != "" {
		fmt.Fprintf(out, "Latest HOS Availability Snapshot: %s\n", details.LatestHosAvailabilitySnapshotID)
	}
	if len(details.HosAvailabilitySnapshotIDs) > 0 {
		fmt.Fprintf(out, "HOS Availability Snapshots: %s\n", strings.Join(details.HosAvailabilitySnapshotIDs, ", "))
	}
	if len(details.HosEventIDs) > 0 {
		fmt.Fprintf(out, "HOS Events: %s\n", strings.Join(details.HosEventIDs, ", "))
	}

	return nil
}
