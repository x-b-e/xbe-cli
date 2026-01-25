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

type hosDayRegulationSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type hosDayRegulationSetDetails struct {
	ID                      string `json:"id"`
	BrokerID                string `json:"broker_id,omitempty"`
	HosDayID                string `json:"hos_day_id,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	RegulationSetCode       string `json:"regulation_set_code,omitempty"`
	TimeZoneID              string `json:"time_zone_id,omitempty"`
	AvailabilityBreakdown   any    `json:"availability_breakdown,omitempty"`
	Recap                   any    `json:"recap,omitempty"`
	CycleAvailabilities     any    `json:"cycle_availabilities,omitempty"`
	AdverseDrivingAvailable bool   `json:"adverse_driving_available"`
	AdverseDrivingApplied   bool   `json:"adverse_driving_applied"`
	CreatedAt               string `json:"created_at,omitempty"`
	UpdatedAt               string `json:"updated_at,omitempty"`
}

func newHosDayRegulationSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show HOS day regulation set details",
		Long: `Show the full details of a HOS day regulation set.

Output Fields:
  ID                 Regulation set identifier
  Broker             Broker ID (if present)
  HOS Day            HOS day ID (if present)
  Driver             Driver (user) ID
  Regulation Set     Regulation set code
  Time Zone          Time zone identifier
  Adverse Available  Adverse driving availability (true/false)
  Adverse Applied    Adverse driving applied (true/false)
  Availability       Availability breakdown (JSON)
  Recap              Recap details (JSON)
  Cycle Availability Cycle availability details (JSON)
  Created At         Record creation timestamp
  Updated At         Record update timestamp

Arguments:
  <id>    The regulation set ID (required). You can find IDs using the list command.`,
		Example: `  # Show a HOS day regulation set
  xbe view hos-day-regulation-sets show 123

  # Output as JSON
  xbe view hos-day-regulation-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHosDayRegulationSetsShow,
	}
	initHosDayRegulationSetsShowFlags(cmd)
	return cmd
}

func init() {
	hosDayRegulationSetsCmd.AddCommand(newHosDayRegulationSetsShowCmd())
}

func initHosDayRegulationSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosDayRegulationSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseHosDayRegulationSetsShowOptions(cmd)
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
		return fmt.Errorf("hos day regulation set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[hos-day-regulation-sets]", "regulation-set-code,time-zone-id,availability-breakdown,recap,cycle-availabilities,adverse-driving-available,adverse-driving-applied,created-at,updated-at,broker,hos-day,user")

	body, _, err := client.Get(cmd.Context(), "/v1/hos-day-regulation-sets/"+id, query)
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

	details := buildHosDayRegulationSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHosDayRegulationSetDetails(cmd, details)
}

func parseHosDayRegulationSetsShowOptions(cmd *cobra.Command) (hosDayRegulationSetsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return hosDayRegulationSetsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return hosDayRegulationSetsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return hosDayRegulationSetsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return hosDayRegulationSetsShowOptions{}, err
	}

	return hosDayRegulationSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHosDayRegulationSetDetails(resp jsonAPISingleResponse) hosDayRegulationSetDetails {
	attrs := resp.Data.Attributes
	row := buildHosDayRegulationSetRowFromSingle(resp)

	details := hosDayRegulationSetDetails{
		ID:                      resp.Data.ID,
		BrokerID:                row.BrokerID,
		HosDayID:                row.HosDayID,
		UserID:                  row.UserID,
		RegulationSetCode:       strings.TrimSpace(stringAttr(attrs, "regulation-set-code")),
		TimeZoneID:              strings.TrimSpace(stringAttr(attrs, "time-zone-id")),
		AvailabilityBreakdown:   attrs["availability-breakdown"],
		Recap:                   attrs["recap"],
		CycleAvailabilities:     attrs["cycle-availabilities"],
		AdverseDrivingAvailable: boolAttr(attrs, "adverse-driving-available"),
		AdverseDrivingApplied:   boolAttr(attrs, "adverse-driving-applied"),
		CreatedAt:               formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:               formatDateTime(stringAttr(attrs, "updated-at")),
	}

	return details
}

func renderHosDayRegulationSetDetails(cmd *cobra.Command, details hosDayRegulationSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.HosDayID != "" {
		fmt.Fprintf(out, "HOS Day: %s\n", details.HosDayID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "Driver: %s\n", details.UserID)
	}
	if details.RegulationSetCode != "" {
		fmt.Fprintf(out, "Regulation Set: %s\n", details.RegulationSetCode)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", details.TimeZoneID)
	}
	fmt.Fprintf(out, "Adverse Available: %t\n", details.AdverseDrivingAvailable)
	fmt.Fprintf(out, "Adverse Applied: %t\n", details.AdverseDrivingApplied)
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.AvailabilityBreakdown != nil {
		fmt.Fprintln(out, "Availability Breakdown:")
		fmt.Fprintln(out, formatHosDayRegulationSetJSON(details.AvailabilityBreakdown))
	}
	if details.Recap != nil {
		fmt.Fprintln(out, "Recap:")
		fmt.Fprintln(out, formatHosDayRegulationSetJSON(details.Recap))
	}
	if details.CycleAvailabilities != nil {
		fmt.Fprintln(out, "Cycle Availabilities:")
		fmt.Fprintln(out, formatHosDayRegulationSetJSON(details.CycleAvailabilities))
	}

	return nil
}

func formatHosDayRegulationSetJSON(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}
