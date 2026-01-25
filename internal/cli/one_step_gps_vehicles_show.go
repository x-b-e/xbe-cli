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

type oneStepGpsVehiclesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type oneStepGpsVehicleDetails struct {
	ID                                       string `json:"id"`
	VehicleID                                string `json:"vehicle_id,omitempty"`
	VehicleNumber                            string `json:"vehicle_number,omitempty"`
	IntegrationIdentifier                    string `json:"integration_identifier,omitempty"`
	TrailerSetAt                             string `json:"trailer_set_at,omitempty"`
	TractorSetAt                             string `json:"tractor_set_at,omitempty"`
	SkipTrailerIsNotAlreadyMatchedValidation bool   `json:"skip_trailer_is_not_already_matched_validation,omitempty"`
	SkipTractorIsNotAlreadyMatchedValidation bool   `json:"skip_tractor_is_not_already_matched_validation,omitempty"`
	BrokerID                                 string `json:"broker_id,omitempty"`
	TruckerID                                string `json:"trucker_id,omitempty"`
	TrailerID                                string `json:"trailer_id,omitempty"`
	TractorID                                string `json:"tractor_id,omitempty"`
}

func newOneStepGpsVehiclesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show One Step GPS vehicle details",
		Long: `Show the full details of a One Step GPS vehicle.

Output Fields:
  ID                 One Step GPS vehicle identifier
  VEHICLE NUMBER     One Step GPS vehicle number
  VEHICLE ID         One Step GPS vehicle external ID
  INTEGRATION ID     Integration identifier
  TRAILER SET AT     Trailer assignment timestamp
  TRACTOR SET AT     Tractor assignment timestamp
  SKIP TRAILER VALIDATION  Skip trailer match validation
  SKIP TRACTOR VALIDATION  Skip tractor match validation
  BROKER ID          Broker ID
  TRUCKER ID         Trucker ID
  TRAILER ID         Trailer ID
  TRACTOR ID         Tractor ID

Arguments:
  <id>    The One Step GPS vehicle ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a One Step GPS vehicle
  xbe view one-step-gps-vehicles show 123

  # Output as JSON
  xbe view one-step-gps-vehicles show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOneStepGpsVehiclesShow,
	}
	initOneStepGpsVehiclesShowFlags(cmd)
	return cmd
}

func init() {
	oneStepGpsVehiclesCmd.AddCommand(newOneStepGpsVehiclesShowCmd())
}

func initOneStepGpsVehiclesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOneStepGpsVehiclesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOneStepGpsVehiclesShowOptions(cmd)
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
		return fmt.Errorf("One Step GPS vehicle id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[one-step-gps-vehicles]", "vehicle-id,vehicle-number,integration-identifier,trailer-set-at,tractor-set-at,skip-trailer-is-not-already-matched-validation,skip-tractor-is-not-already-matched-validation,broker,trucker,trailer,tractor")
	query.Set("include", "broker,trucker,trailer,tractor")

	body, _, err := client.Get(cmd.Context(), "/v1/one-step-gps-vehicles/"+id, query)
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

	details := buildOneStepGpsVehicleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOneStepGpsVehicleDetails(cmd, details)
}

func parseOneStepGpsVehiclesShowOptions(cmd *cobra.Command) (oneStepGpsVehiclesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return oneStepGpsVehiclesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return oneStepGpsVehiclesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return oneStepGpsVehiclesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return oneStepGpsVehiclesShowOptions{}, err
	}

	return oneStepGpsVehiclesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOneStepGpsVehicleDetails(resp jsonAPISingleResponse) oneStepGpsVehicleDetails {
	attrs := resp.Data.Attributes
	return oneStepGpsVehicleDetails{
		ID:                                       resp.Data.ID,
		VehicleID:                                stringAttr(attrs, "vehicle-id"),
		VehicleNumber:                            stringAttr(attrs, "vehicle-number"),
		IntegrationIdentifier:                    stringAttr(attrs, "integration-identifier"),
		TrailerSetAt:                             formatDateTime(stringAttr(attrs, "trailer-set-at")),
		TractorSetAt:                             formatDateTime(stringAttr(attrs, "tractor-set-at")),
		SkipTrailerIsNotAlreadyMatchedValidation: boolAttr(attrs, "skip-trailer-is-not-already-matched-validation"),
		SkipTractorIsNotAlreadyMatchedValidation: boolAttr(attrs, "skip-tractor-is-not-already-matched-validation"),
		BrokerID:                                 relationshipIDFromMap(resp.Data.Relationships, "broker"),
		TruckerID:                                relationshipIDFromMap(resp.Data.Relationships, "trucker"),
		TrailerID:                                relationshipIDFromMap(resp.Data.Relationships, "trailer"),
		TractorID:                                relationshipIDFromMap(resp.Data.Relationships, "tractor"),
	}
}

func renderOneStepGpsVehicleDetails(cmd *cobra.Command, details oneStepGpsVehicleDetails) error {
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
	if details.SkipTrailerIsNotAlreadyMatchedValidation {
		fmt.Fprintf(out, "Skip Trailer Validation: %t\n", details.SkipTrailerIsNotAlreadyMatchedValidation)
	}
	if details.SkipTractorIsNotAlreadyMatchedValidation {
		fmt.Fprintf(out, "Skip Tractor Validation: %t\n", details.SkipTractorIsNotAlreadyMatchedValidation)
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
