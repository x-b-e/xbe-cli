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

type tractorOdometerReadingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tractorOdometerReadingDetails struct {
	ID              string `json:"id"`
	DateSequence    string `json:"date_sequence,omitempty"`
	ReadingOn       string `json:"reading_on,omitempty"`
	ReadingTime     string `json:"reading_time,omitempty"`
	StateCode       string `json:"state_code,omitempty"`
	Value           string `json:"value,omitempty"`
	TractorID       string `json:"tractor_id,omitempty"`
	DriverDayID     string `json:"driver_day_id,omitempty"`
	UnitOfMeasureID string `json:"unit_of_measure_id,omitempty"`
	CreatedByID     string `json:"created_by_id,omitempty"`
}

func newTractorOdometerReadingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tractor odometer reading details",
		Long: `Show the full details of a tractor odometer reading.

Output Fields:
  ID
  Date Sequence
  Reading On
  Reading Time
  State Code
  Value
  Tractor ID
  Driver Day ID
  Unit of Measure ID
  Created By (user ID)

Arguments:
  <id>    The odometer reading ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a tractor odometer reading
  xbe view tractor-odometer-readings show 123

  # Output as JSON
  xbe view tractor-odometer-readings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTractorOdometerReadingsShow,
	}
	initTractorOdometerReadingsShowFlags(cmd)
	return cmd
}

func init() {
	tractorOdometerReadingsCmd.AddCommand(newTractorOdometerReadingsShowCmd())
}

func initTractorOdometerReadingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTractorOdometerReadingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTractorOdometerReadingsShowOptions(cmd)
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
		return fmt.Errorf("tractor odometer reading id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tractor-odometer-readings]", "date-sequence,reading-on,reading-time,state-code,value,tractor,unit-of-measure,driver-day,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/tractor-odometer-readings/"+id, query)
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

	details := buildTractorOdometerReadingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTractorOdometerReadingDetails(cmd, details)
}

func parseTractorOdometerReadingsShowOptions(cmd *cobra.Command) (tractorOdometerReadingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tractorOdometerReadingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTractorOdometerReadingDetails(resp jsonAPISingleResponse) tractorOdometerReadingDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := tractorOdometerReadingDetails{
		ID:           resource.ID,
		DateSequence: stringAttr(attrs, "date-sequence"),
		ReadingOn:    formatDate(stringAttr(attrs, "reading-on")),
		ReadingTime:  formatTime(stringAttr(attrs, "reading-time")),
		StateCode:    stringAttr(attrs, "state-code"),
		Value:        stringAttr(attrs, "value"),
	}

	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		details.TractorID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		details.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderTractorOdometerReadingDetails(cmd *cobra.Command, details tractorOdometerReadingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DateSequence != "" {
		fmt.Fprintf(out, "Date Sequence: %s\n", details.DateSequence)
	}
	if details.ReadingOn != "" {
		fmt.Fprintf(out, "Reading On: %s\n", details.ReadingOn)
	}
	if details.ReadingTime != "" {
		fmt.Fprintf(out, "Reading Time: %s\n", details.ReadingTime)
	}
	if details.StateCode != "" {
		fmt.Fprintf(out, "State Code: %s\n", details.StateCode)
	}
	if details.Value != "" {
		fmt.Fprintf(out, "Value: %s\n", details.Value)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor ID: %s\n", details.TractorID)
	}
	if details.DriverDayID != "" {
		fmt.Fprintf(out, "Driver Day ID: %s\n", details.DriverDayID)
	}
	if details.UnitOfMeasureID != "" {
		fmt.Fprintf(out, "Unit of Measure ID: %s\n", details.UnitOfMeasureID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}

	return nil
}
