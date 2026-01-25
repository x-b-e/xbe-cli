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

type digitalFleetTrucksShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type digitalFleetTruckDetails struct {
	ID                    string `json:"id"`
	TruckID               string `json:"truck_id,omitempty"`
	TruckNumber           string `json:"truck_number,omitempty"`
	IsActive              bool   `json:"is_active,omitempty"`
	IntegrationIdentifier string `json:"integration_identifier,omitempty"`
	TrailerSetAt          string `json:"trailer_set_at,omitempty"`
	TractorSetAt          string `json:"tractor_set_at,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	BrokerName            string `json:"broker_name,omitempty"`
	TruckerID             string `json:"trucker_id,omitempty"`
	TruckerName           string `json:"trucker_name,omitempty"`
	TrailerID             string `json:"trailer_id,omitempty"`
	TrailerNumber         string `json:"trailer_number,omitempty"`
	TractorID             string `json:"tractor_id,omitempty"`
	TractorNumber         string `json:"tractor_number,omitempty"`
}

func newDigitalFleetTrucksShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show digital fleet truck details",
		Long: `Show the full details of a digital fleet truck.

Output Fields:
  ID               Digital fleet truck identifier
  Truck ID         Digital fleet truck source identifier
  Truck Number     Digital fleet truck number
  Active           Active status
  Integration ID   Integration identifier
  Trailer Set At   Trailer assignment timestamp
  Tractor Set At   Tractor assignment timestamp
  Trailer          Assigned trailer number or ID
  Tractor          Assigned tractor number or ID
  Trucker          Trucker name or ID
  Broker           Broker name or ID

Arguments:
  <id>  Digital fleet truck ID (required). Find IDs using the list command.`,
		Example: `  # Show digital fleet truck details
  xbe view digital-fleet-trucks show 123

  # Output as JSON
  xbe view digital-fleet-trucks show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDigitalFleetTrucksShow,
	}
	initDigitalFleetTrucksShowFlags(cmd)
	return cmd
}

func init() {
	digitalFleetTrucksCmd.AddCommand(newDigitalFleetTrucksShowCmd())
}

func initDigitalFleetTrucksShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDigitalFleetTrucksShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDigitalFleetTrucksShowOptions(cmd)
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
		return fmt.Errorf("digital fleet truck id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[digital-fleet-trucks]", strings.Join([]string{
		"truck-id",
		"truck-number",
		"is-active",
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

	body, _, err := client.Get(cmd.Context(), "/v1/digital-fleet-trucks/"+id, query)
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

	details := buildDigitalFleetTruckDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDigitalFleetTruckDetails(cmd, details)
}

func parseDigitalFleetTrucksShowOptions(cmd *cobra.Command) (digitalFleetTrucksShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return digitalFleetTrucksShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDigitalFleetTruckDetails(resp jsonAPISingleResponse) digitalFleetTruckDetails {
	row := digitalFleetTruckRowFromSingle(resp)

	details := digitalFleetTruckDetails{
		ID:                    row.ID,
		TruckID:               row.TruckID,
		TruckNumber:           row.TruckNumber,
		IsActive:              row.IsActive,
		IntegrationIdentifier: row.IntegrationIdentifier,
		TrailerSetAt:          row.TrailerSetAt,
		TractorSetAt:          row.TractorSetAt,
		BrokerID:              row.BrokerID,
		BrokerName:            row.BrokerName,
		TruckerID:             row.TruckerID,
		TruckerName:           row.TruckerName,
		TrailerID:             row.TrailerID,
		TrailerNumber:         row.TrailerNumber,
		TractorID:             row.TractorID,
		TractorNumber:         row.TractorNumber,
	}

	return details
}

func renderDigitalFleetTruckDetails(cmd *cobra.Command, details digitalFleetTruckDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckID != "" {
		fmt.Fprintf(out, "Truck ID: %s\n", details.TruckID)
	}
	if details.TruckNumber != "" {
		fmt.Fprintf(out, "Truck Number: %s\n", details.TruckNumber)
	}
	fmt.Fprintf(out, "Active: %s\n", formatBool(details.IsActive))
	if details.IntegrationIdentifier != "" {
		fmt.Fprintf(out, "Integration ID: %s\n", details.IntegrationIdentifier)
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
