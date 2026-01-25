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

type teletracNavmanVehiclesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type teletracNavmanVehicleDetails struct {
	ID            string `json:"id"`
	VehicleID     string `json:"vehicle_id,omitempty"`
	VehicleNumber string `json:"vehicle_number,omitempty"`
	IntegrationID string `json:"integration_identifier,omitempty"`
	TrailerSetAt  string `json:"trailer_set_at,omitempty"`
	TractorSetAt  string `json:"tractor_set_at,omitempty"`
	BrokerID      string `json:"broker_id,omitempty"`
	BrokerName    string `json:"broker_name,omitempty"`
	TruckerID     string `json:"trucker_id,omitempty"`
	TruckerName   string `json:"trucker_name,omitempty"`
	TrailerID     string `json:"trailer_id,omitempty"`
	TrailerNumber string `json:"trailer_number,omitempty"`
	TractorID     string `json:"tractor_id,omitempty"`
	TractorNumber string `json:"tractor_number,omitempty"`
}

func newTeletracNavmanVehiclesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show Teletrac Navman vehicle details",
		Long: `Show the full details of a Teletrac Navman vehicle.

Output Fields:
  ID               Teletrac Navman vehicle identifier
  Vehicle ID       Teletrac Navman vehicle source identifier
  Vehicle Number   Teletrac Navman vehicle number
  Integration ID   Integration identifier
  Trailer Set At   Trailer assignment timestamp
  Tractor Set At   Tractor assignment timestamp
  Trailer          Assigned trailer number or ID
  Tractor          Assigned tractor number or ID
  Trucker          Trucker name or ID
  Broker           Broker name or ID

Arguments:
  <id>  Teletrac Navman vehicle ID (required). Find IDs using the list command.`,
		Example: `  # Show Teletrac Navman vehicle details
  xbe view teletrac-navman-vehicles show 123

  # Output as JSON
  xbe view teletrac-navman-vehicles show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTeletracNavmanVehiclesShow,
	}
	initTeletracNavmanVehiclesShowFlags(cmd)
	return cmd
}

func init() {
	teletracNavmanVehiclesCmd.AddCommand(newTeletracNavmanVehiclesShowCmd())
}

func initTeletracNavmanVehiclesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTeletracNavmanVehiclesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTeletracNavmanVehiclesShowOptions(cmd)
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
		return fmt.Errorf("teletrac navman vehicle id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[teletrac-navman-vehicles]", strings.Join([]string{
		"vehicle-id",
		"vehicle-number",
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

	body, _, err := client.Get(cmd.Context(), "/v1/teletrac-navman-vehicles/"+id, query)
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

	details := buildTeletracNavmanVehicleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTeletracNavmanVehicleDetails(cmd, details)
}

func parseTeletracNavmanVehiclesShowOptions(cmd *cobra.Command) (teletracNavmanVehiclesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return teletracNavmanVehiclesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTeletracNavmanVehicleDetails(resp jsonAPISingleResponse) teletracNavmanVehicleDetails {
	row := teletracNavmanVehicleRowFromSingle(resp)

	details := teletracNavmanVehicleDetails{
		ID:            row.ID,
		VehicleID:     row.VehicleID,
		VehicleNumber: row.VehicleNumber,
		IntegrationID: row.IntegrationIdentifier,
		TrailerSetAt:  row.TrailerSetAt,
		TractorSetAt:  row.TractorSetAt,
		BrokerID:      row.BrokerID,
		BrokerName:    row.BrokerName,
		TruckerID:     row.TruckerID,
		TruckerName:   row.TruckerName,
		TrailerID:     row.TrailerID,
		TrailerNumber: row.TrailerNumber,
		TractorID:     row.TractorID,
		TractorNumber: row.TractorNumber,
	}

	return details
}

func renderTeletracNavmanVehicleDetails(cmd *cobra.Command, details teletracNavmanVehicleDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.VehicleID != "" {
		fmt.Fprintf(out, "Vehicle ID: %s\n", details.VehicleID)
	}
	if details.VehicleNumber != "" {
		fmt.Fprintf(out, "Vehicle Number: %s\n", details.VehicleNumber)
	}
	if details.IntegrationID != "" {
		fmt.Fprintf(out, "Integration ID: %s\n", details.IntegrationID)
	}
	if details.TrailerSetAt != "" {
		fmt.Fprintf(out, "Trailer Set At: %s\n", details.TrailerSetAt)
	}
	if details.TractorSetAt != "" {
		fmt.Fprintf(out, "Tractor Set At: %s\n", details.TractorSetAt)
	}
	if details.TrailerID != "" || details.TrailerNumber != "" {
		fmt.Fprintf(out, "Trailer: %s\n", formatRelated(details.TrailerNumber, details.TrailerID))
	}
	if details.TractorID != "" || details.TractorNumber != "" {
		fmt.Fprintf(out, "Tractor: %s\n", formatRelated(details.TractorNumber, details.TractorID))
	}
	if details.TruckerID != "" || details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker: %s\n", formatRelated(details.TruckerName, details.TruckerID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}

	return nil
}
