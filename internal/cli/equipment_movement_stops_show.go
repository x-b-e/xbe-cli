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

type equipmentMovementStopsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementStopDetails struct {
	ID                 string   `json:"id"`
	TripID             string   `json:"trip_id,omitempty"`
	TripJobNumber      string   `json:"trip_job_number,omitempty"`
	LocationID         string   `json:"location_id,omitempty"`
	LocationName       string   `json:"location,omitempty"`
	ScheduledArrivalAt string   `json:"scheduled_arrival_at,omitempty"`
	SequencePosition   string   `json:"sequence_position,omitempty"`
	SequenceIndex      string   `json:"sequence_index,omitempty"`
	StopCompletionID   string   `json:"stop_completion_id,omitempty"`
	StopRequirementIDs []string `json:"stop_requirement_ids,omitempty"`
}

func newEquipmentMovementStopsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement stop details",
		Long: `Show the full details of an equipment movement stop.

Output Fields:
  ID
  Trip ID
  Trip Job Number
  Location ID
  Location Name
  Scheduled Arrival At
  Sequence Position
  Sequence Index
  Stop Completion ID
  Stop Requirement IDs

Arguments:
  <id>    The stop ID (required). You can find IDs using the list command.`,
		Example: `  # Show a stop
  xbe view equipment-movement-stops show 123

  # Get JSON output
  xbe view equipment-movement-stops show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementStopsShow,
	}
	initEquipmentMovementStopsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementStopsCmd.AddCommand(newEquipmentMovementStopsShowCmd())
}

func initEquipmentMovementStopsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementStopsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentMovementStopsShowOptions(cmd)
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
		return fmt.Errorf("equipment movement stop id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "trip,location")
	query.Set("fields[equipment-movement-stops]", "sequence-position,scheduled-arrival-at,sequence-index,trip,location,stop-completion,stop-requirements")
	query.Set("fields[equipment-movement-trips]", "job-number")
	query.Set("fields[equipment-movement-requirement-locations]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-stops/"+id, query)
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

	details := buildEquipmentMovementStopDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementStopDetails(cmd, details)
}

func parseEquipmentMovementStopsShowOptions(cmd *cobra.Command) (equipmentMovementStopsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementStopsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementStopDetails(resp jsonAPISingleResponse) equipmentMovementStopDetails {
	details := equipmentMovementStopDetails{
		ID:                 resp.Data.ID,
		ScheduledArrivalAt: formatDateTime(stringAttr(resp.Data.Attributes, "scheduled-arrival-at")),
		SequencePosition:   stringAttr(resp.Data.Attributes, "sequence-position"),
		SequenceIndex:      stringAttr(resp.Data.Attributes, "sequence-index"),
	}

	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		if inc.Attributes == nil {
			continue
		}
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	if rel, ok := resp.Data.Relationships["trip"]; ok && rel.Data != nil {
		details.TripID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TripJobNumber = stringAttr(attrs, "job-number")
		}
	}
	if rel, ok := resp.Data.Relationships["location"]; ok && rel.Data != nil {
		details.LocationID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LocationName = stringAttr(attrs, "name")
		}
	}
	if rel, ok := resp.Data.Relationships["stop-completion"]; ok && rel.Data != nil {
		details.StopCompletionID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["stop-requirements"]; ok {
		details.StopRequirementIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderEquipmentMovementStopDetails(cmd *cobra.Command, details equipmentMovementStopDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TripJobNumber != "" {
		fmt.Fprintf(out, "Trip Job Number: %s\n", details.TripJobNumber)
	}
	if details.TripID != "" {
		fmt.Fprintf(out, "Trip ID: %s\n", details.TripID)
	}
	if details.LocationName != "" {
		fmt.Fprintf(out, "Location: %s\n", details.LocationName)
	}
	if details.LocationID != "" {
		fmt.Fprintf(out, "Location ID: %s\n", details.LocationID)
	}
	if details.ScheduledArrivalAt != "" {
		fmt.Fprintf(out, "Scheduled Arrival At: %s\n", details.ScheduledArrivalAt)
	}
	if details.SequencePosition != "" {
		fmt.Fprintf(out, "Sequence Position: %s\n", details.SequencePosition)
	}
	if details.SequenceIndex != "" {
		fmt.Fprintf(out, "Sequence Index: %s\n", details.SequenceIndex)
	}
	if details.StopCompletionID != "" {
		fmt.Fprintf(out, "Stop Completion ID: %s\n", details.StopCompletionID)
	}
	if len(details.StopRequirementIDs) > 0 {
		fmt.Fprintf(out, "Stop Requirement IDs: %s\n", strings.Join(details.StopRequirementIDs, ", "))
	}

	return nil
}
