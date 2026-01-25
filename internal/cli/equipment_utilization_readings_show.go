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

type equipmentUtilizationReadingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentUtilizationReadingDetails struct {
	ID             string `json:"id"`
	EquipmentID    string `json:"equipment_id,omitempty"`
	BusinessUnitID string `json:"business_unit_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	ReportedAt     string `json:"reported_at,omitempty"`
	Odometer       string `json:"odometer,omitempty"`
	Hourmeter      string `json:"hourmeter,omitempty"`
	Source         string `json:"source,omitempty"`
	OtherReadings  any    `json:"other_readings,omitempty"`
}

func newEquipmentUtilizationReadingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment utilization reading details",
		Long: `Show the full details of a specific equipment utilization reading.

Output Fields:
  ID            Reading identifier
  Equipment     Equipment ID
  Business Unit Business unit ID
  User          User ID
  Reported At   Reported timestamp
  Odometer      Odometer reading
  Hourmeter     Hourmeter reading
  Source        Reading source
  Other Readings JSON payload

Arguments:
  <id>    The reading ID (required). You can find IDs using the list command.`,
		Example: `  # Show a reading
  xbe view equipment-utilization-readings show 123

  # Get JSON output
  xbe view equipment-utilization-readings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentUtilizationReadingsShow,
	}
	initEquipmentUtilizationReadingsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentUtilizationReadingsCmd.AddCommand(newEquipmentUtilizationReadingsShowCmd())
}

func initEquipmentUtilizationReadingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentUtilizationReadingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentUtilizationReadingsShowOptions(cmd)
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
		return fmt.Errorf("equipment utilization reading id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-utilization-readings]", "odometer,hourmeter,reported-at,source,other-readings,equipment,business-unit,user")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-utilization-readings/"+id, query)
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

	details := buildEquipmentUtilizationReadingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentUtilizationReadingDetails(cmd, details)
}

func parseEquipmentUtilizationReadingsShowOptions(cmd *cobra.Command) (equipmentUtilizationReadingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentUtilizationReadingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentUtilizationReadingDetails(resp jsonAPISingleResponse) equipmentUtilizationReadingDetails {
	attrs := resp.Data.Attributes

	details := equipmentUtilizationReadingDetails{
		ID:            resp.Data.ID,
		ReportedAt:    formatDateTime(stringAttr(attrs, "reported-at")),
		Odometer:      stringAttr(attrs, "odometer"),
		Hourmeter:     stringAttr(attrs, "hourmeter"),
		Source:        stringAttr(attrs, "source"),
		OtherReadings: attrs["other-readings"],
	}

	if rel, ok := resp.Data.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["business-unit"]; ok && rel.Data != nil {
		details.BusinessUnitID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	return details
}

func renderEquipmentUtilizationReadingDetails(cmd *cobra.Command, details equipmentUtilizationReadingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EquipmentID != "" {
		fmt.Fprintf(out, "Equipment: %s\n", details.EquipmentID)
	}
	if details.BusinessUnitID != "" {
		fmt.Fprintf(out, "Business Unit: %s\n", details.BusinessUnitID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}
	if details.ReportedAt != "" {
		fmt.Fprintf(out, "Reported At: %s\n", details.ReportedAt)
	}
	if details.Odometer != "" {
		fmt.Fprintf(out, "Odometer: %s\n", details.Odometer)
	}
	if details.Hourmeter != "" {
		fmt.Fprintf(out, "Hourmeter: %s\n", details.Hourmeter)
	}
	if details.Source != "" {
		fmt.Fprintf(out, "Source: %s\n", details.Source)
	}
	if other := formatJSONValue(details.OtherReadings); other != "" {
		fmt.Fprintf(out, "Other Readings: %s\n", other)
	}

	return nil
}
