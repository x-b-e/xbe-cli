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

type tennaVehiclesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tennaVehicleDetails struct {
	ID                      string `json:"id"`
	VehicleID               string `json:"vehicle_id,omitempty"`
	VehicleNumber           string `json:"vehicle_number,omitempty"`
	SerialNumber            string `json:"serial_number,omitempty"`
	IntegrationIdentifier   string `json:"integration_identifier,omitempty"`
	TrailerSetAt            string `json:"trailer_set_at,omitempty"`
	TractorSetAt            string `json:"tractor_set_at,omitempty"`
	EquipmentSetAt          string `json:"equipment_set_at,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	BrokerName              string `json:"broker_name,omitempty"`
	TruckerID               string `json:"trucker_id,omitempty"`
	TruckerName             string `json:"trucker_name,omitempty"`
	TrailerID               string `json:"trailer_id,omitempty"`
	TrailerNumber           string `json:"trailer_number,omitempty"`
	TractorID               string `json:"tractor_id,omitempty"`
	TractorNumber           string `json:"tractor_number,omitempty"`
	EquipmentID             string `json:"equipment_id,omitempty"`
	EquipmentNickname       string `json:"equipment_nickname,omitempty"`
	AssignedEquipmentSerial string `json:"assigned_equipment_serial_number,omitempty"`
}

func newTennaVehiclesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show Tenna vehicle details",
		Long: `Show the full details of a Tenna vehicle.

Output Fields:
  ID                 Tenna vehicle identifier
  Vehicle ID         Tenna vehicle source identifier
  Vehicle Number     Tenna vehicle number
  Serial Number      Tenna vehicle serial number
  Integration ID     Integration identifier
  Trailer Set At     Trailer assignment timestamp
  Tractor Set At     Tractor assignment timestamp
  Equipment Set At   Equipment assignment timestamp
  Trailer            Assigned trailer number or ID
  Tractor            Assigned tractor number or ID
  Equipment          Assigned equipment nickname or ID
  Assigned Serial    Assigned equipment serial number
  Trucker            Trucker name or ID
  Broker             Broker name or ID

Arguments:
  <id>  Tenna vehicle ID (required). Find IDs using the list command.`,
		Example: `  # Show Tenna vehicle details
  xbe view tenna-vehicles show 123

  # Output as JSON
  xbe view tenna-vehicles show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTennaVehiclesShow,
	}
	initTennaVehiclesShowFlags(cmd)
	return cmd
}

func init() {
	tennaVehiclesCmd.AddCommand(newTennaVehiclesShowCmd())
}

func initTennaVehiclesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTennaVehiclesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTennaVehiclesShowOptions(cmd)
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
		return fmt.Errorf("tenna vehicle id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tenna-vehicles]", strings.Join([]string{
		"vehicle-id",
		"vehicle-number",
		"serial-number",
		"integration-identifier",
		"trailer-set-at",
		"tractor-set-at",
		"equipment-set-at",
		"broker",
		"trucker",
		"tractor",
		"trailer",
		"equipment",
	}, ","))
	query.Set("include", "broker,trucker,tractor,trailer,equipment")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[tractors]", "number")
	query.Set("fields[trailers]", "number")
	query.Set("fields[equipment]", "nickname,serial-number")

	body, _, err := client.Get(cmd.Context(), "/v1/tenna-vehicles/"+id, query)
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

	details := buildTennaVehicleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTennaVehicleDetails(cmd, details)
}

func parseTennaVehiclesShowOptions(cmd *cobra.Command) (tennaVehiclesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tennaVehiclesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTennaVehicleDetails(resp jsonAPISingleResponse) tennaVehicleDetails {
	row := tennaVehicleRowFromSingle(resp)

	details := tennaVehicleDetails{
		ID:                      row.ID,
		VehicleID:               row.VehicleID,
		VehicleNumber:           row.VehicleNumber,
		SerialNumber:            row.SerialNumber,
		IntegrationIdentifier:   row.IntegrationIdentifier,
		TrailerSetAt:            row.TrailerSetAt,
		TractorSetAt:            row.TractorSetAt,
		EquipmentSetAt:          row.EquipmentSetAt,
		BrokerID:                row.BrokerID,
		BrokerName:              row.BrokerName,
		TruckerID:               row.TruckerID,
		TruckerName:             row.TruckerName,
		TrailerID:               row.TrailerID,
		TrailerNumber:           row.TrailerNumber,
		TractorID:               row.TractorID,
		TractorNumber:           row.TractorNumber,
		EquipmentID:             row.EquipmentID,
		EquipmentNickname:       row.EquipmentNickname,
		AssignedEquipmentSerial: row.AssignedEquipmentSerial,
	}

	return details
}

func renderTennaVehicleDetails(cmd *cobra.Command, details tennaVehicleDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.VehicleID != "" {
		fmt.Fprintf(out, "Vehicle ID: %s\n", details.VehicleID)
	}
	if details.VehicleNumber != "" {
		fmt.Fprintf(out, "Vehicle Number: %s\n", details.VehicleNumber)
	}
	if details.SerialNumber != "" {
		fmt.Fprintf(out, "Serial Number: %s\n", details.SerialNumber)
	}
	if details.IntegrationIdentifier != "" {
		fmt.Fprintf(out, "Integration ID: %s\n", details.IntegrationIdentifier)
	}
	if details.TrailerSetAt != "" {
		fmt.Fprintf(out, "Trailer Set At: %s\n", details.TrailerSetAt)
	}
	if details.TractorSetAt != "" {
		fmt.Fprintf(out, "Tractor Set At: %s\n", details.TractorSetAt)
	}
	if details.EquipmentSetAt != "" {
		fmt.Fprintf(out, "Equipment Set At: %s\n", details.EquipmentSetAt)
	}
	if details.TrailerID != "" || details.TrailerNumber != "" {
		fmt.Fprintf(out, "Trailer: %s\n", formatRelated(details.TrailerNumber, details.TrailerID))
	}
	if details.TractorID != "" || details.TractorNumber != "" {
		fmt.Fprintf(out, "Tractor: %s\n", formatRelated(details.TractorNumber, details.TractorID))
	}
	if details.EquipmentID != "" || details.EquipmentNickname != "" {
		fmt.Fprintf(out, "Equipment: %s\n", formatRelated(details.EquipmentNickname, details.EquipmentID))
	}
	if details.AssignedEquipmentSerial != "" {
		fmt.Fprintf(out, "Assigned Serial: %s\n", details.AssignedEquipmentSerial)
	}
	if details.TruckerID != "" || details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker: %s\n", formatRelated(details.TruckerName, details.TruckerID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}

	return nil
}
