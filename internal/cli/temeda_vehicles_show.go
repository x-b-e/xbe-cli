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

type temedaVehiclesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type temedaVehicleDetails struct {
	ID                    string `json:"id"`
	VehicleID             string `json:"vehicle_id,omitempty"`
	VehicleNumber         string `json:"vehicle_number,omitempty"`
	IntegrationIdentifier string `json:"integration_identifier,omitempty"`
	TrailerSetAt          string `json:"trailer_set_at,omitempty"`
	TractorSetAt          string `json:"tractor_set_at,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	TruckerID             string `json:"trucker_id,omitempty"`
	TrailerID             string `json:"trailer_id,omitempty"`
	TractorID             string `json:"tractor_id,omitempty"`
}

func newTemedaVehiclesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show temeda vehicle details",
		Long: `Show the full details of a temeda vehicle.

Output Fields:
  ID                 Temeda vehicle identifier
  VEHICLE NUMBER     Temeda vehicle number
  VEHICLE ID         Temeda vehicle external ID
  INTEGRATION ID     Integration identifier
  TRAILER SET AT     Trailer assignment timestamp
  TRACTOR SET AT     Tractor assignment timestamp
  BROKER ID          Broker ID
  TRUCKER ID         Trucker ID
  TRAILER ID         Trailer ID
  TRACTOR ID         Tractor ID

Arguments:
  <id>    The temeda vehicle ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a temeda vehicle
  xbe view temeda-vehicles show 123

  # Output as JSON
  xbe view temeda-vehicles show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTemedaVehiclesShow,
	}
	initTemedaVehiclesShowFlags(cmd)
	return cmd
}

func init() {
	temedaVehiclesCmd.AddCommand(newTemedaVehiclesShowCmd())
}

func initTemedaVehiclesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTemedaVehiclesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTemedaVehiclesShowOptions(cmd)
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
		return fmt.Errorf("temeda vehicle id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[temeda-vehicles]", "vehicle-id,vehicle-number,integration-identifier,trailer-set-at,tractor-set-at,broker,trucker,trailer,tractor")
	query.Set("include", "broker,trucker,trailer,tractor")

	body, _, err := client.Get(cmd.Context(), "/v1/temeda-vehicles/"+id, query)
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

	details := buildTemedaVehicleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTemedaVehicleDetails(cmd, details)
}

func parseTemedaVehiclesShowOptions(cmd *cobra.Command) (temedaVehiclesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return temedaVehiclesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return temedaVehiclesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return temedaVehiclesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return temedaVehiclesShowOptions{}, err
	}

	return temedaVehiclesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTemedaVehicleDetails(resp jsonAPISingleResponse) temedaVehicleDetails {
	attrs := resp.Data.Attributes
	return temedaVehicleDetails{
		ID:                    resp.Data.ID,
		VehicleID:             stringAttr(attrs, "vehicle-id"),
		VehicleNumber:         stringAttr(attrs, "vehicle-number"),
		IntegrationIdentifier: stringAttr(attrs, "integration-identifier"),
		TrailerSetAt:          formatDateTime(stringAttr(attrs, "trailer-set-at")),
		TractorSetAt:          formatDateTime(stringAttr(attrs, "tractor-set-at")),
		BrokerID:              relationshipIDFromMap(resp.Data.Relationships, "broker"),
		TruckerID:             relationshipIDFromMap(resp.Data.Relationships, "trucker"),
		TrailerID:             relationshipIDFromMap(resp.Data.Relationships, "trailer"),
		TractorID:             relationshipIDFromMap(resp.Data.Relationships, "tractor"),
	}
}

func renderTemedaVehicleDetails(cmd *cobra.Command, details temedaVehicleDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.VehicleNumber != "" {
		fmt.Fprintf(out, "Vehicle Number: %s\n", details.VehicleNumber)
	}
	if details.VehicleID != "" {
		fmt.Fprintf(out, "Vehicle ID: %s\n", details.VehicleID)
	}
	if details.IntegrationIdentifier != "" {
		fmt.Fprintf(out, "Integration Identifier: %s\n", details.IntegrationIdentifier)
	}
	if details.TrailerSetAt != "" {
		fmt.Fprintf(out, "Trailer Set At: %s\n", details.TrailerSetAt)
	}
	if details.TractorSetAt != "" {
		fmt.Fprintf(out, "Tractor Set At: %s\n", details.TractorSetAt)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer ID: %s\n", details.TrailerID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor ID: %s\n", details.TractorID)
	}

	return nil
}
