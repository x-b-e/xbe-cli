package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanLocationsUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	Name                 string
	SiteKind             string
	IsStartSiteCandidate bool
	NoStartSiteCandidate bool
	SegmentID            string
	Address              string
	AddressLatitude      string
	AddressLongitude     string
	AddressPlaceID       string
	AddressPlusCode      string
	SkipGeocoding        bool
}

func newDoJobProductionPlanLocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan location",
		Long: `Update a job production plan location.

Optional flags:
  --name                      Location name
  --site-kind                 Site kind (job_site, other)
  --is-start-site-candidate   Mark as start site candidate
  --no-is-start-site-candidate Clear start site candidate flag
  --segment                   Job production plan segment ID
  --address                   Full address (will be geocoded)
  --address-latitude          Address latitude
  --address-longitude         Address longitude
  --address-place-id          Google Place ID
  --address-plus-code         Plus code
  --skip-address-geocoding    Skip geocoding the address

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update location name
  xbe do job-production-plan-locations update 123 --name "New Name"

  # Mark as start site candidate
  xbe do job-production-plan-locations update 123 --is-start-site-candidate

  # Update address
  xbe do job-production-plan-locations update 123 \
    --address "456 Oak Ave, Springfield, IL" \
    --address-latitude 39.7817 \
    --address-longitude -89.6501 \
    --skip-address-geocoding`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanLocationsUpdate,
	}
	initDoJobProductionPlanLocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanLocationsCmd.AddCommand(newDoJobProductionPlanLocationsUpdateCmd())
}

func initDoJobProductionPlanLocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("site-kind", "", "Site kind (job_site, other)")
	cmd.Flags().Bool("is-start-site-candidate", false, "Mark as start site candidate")
	cmd.Flags().Bool("no-is-start-site-candidate", false, "Clear start site candidate flag")
	cmd.Flags().String("segment", "", "Job production plan segment ID")
	cmd.Flags().String("address", "", "Full address (will be geocoded)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().String("address-place-id", "", "Google Place ID")
	cmd.Flags().String("address-plus-code", "", "Plus code")
	cmd.Flags().Bool("skip-address-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanLocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanLocationsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("site-kind") {
		attributes["site-kind"] = opts.SiteKind
	}
	if cmd.Flags().Changed("is-start-site-candidate") {
		attributes["is-start-site-candidate"] = true
	}
	if cmd.Flags().Changed("no-is-start-site-candidate") {
		attributes["is-start-site-candidate"] = false
	}
	if cmd.Flags().Changed("address") {
		attributes["address"] = opts.Address
	}
	if cmd.Flags().Changed("address-latitude") {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if cmd.Flags().Changed("address-longitude") {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if cmd.Flags().Changed("address-place-id") {
		attributes["address-place-id"] = opts.AddressPlaceID
	}
	if cmd.Flags().Changed("address-plus-code") {
		attributes["address-plus-code"] = opts.AddressPlusCode
	}
	if cmd.Flags().Changed("skip-address-geocoding") {
		attributes["skip-address-geocoding"] = opts.SkipGeocoding
	}

	if cmd.Flags().Changed("segment") {
		relationships["segment"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-segments",
				"id":   opts.SegmentID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "job-production-plan-locations",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-locations/"+opts.ID, jsonBody)
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

	row := buildJobProductionPlanLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan location %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanLocationsUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanLocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	siteKind, _ := cmd.Flags().GetString("site-kind")
	isStartSiteCandidate, _ := cmd.Flags().GetBool("is-start-site-candidate")
	noStartSiteCandidate, _ := cmd.Flags().GetBool("no-is-start-site-candidate")
	segmentID, _ := cmd.Flags().GetString("segment")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	addressPlaceID, _ := cmd.Flags().GetString("address-place-id")
	addressPlusCode, _ := cmd.Flags().GetString("address-plus-code")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-address-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanLocationsUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		Name:                 name,
		SiteKind:             siteKind,
		IsStartSiteCandidate: isStartSiteCandidate,
		NoStartSiteCandidate: noStartSiteCandidate,
		SegmentID:            segmentID,
		Address:              address,
		AddressLatitude:      addressLatitude,
		AddressLongitude:     addressLongitude,
		AddressPlaceID:       addressPlaceID,
		AddressPlusCode:      addressPlusCode,
		SkipGeocoding:        skipGeocoding,
	}, nil
}
