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

type doJobProductionPlanLocationsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	JobProductionPlanID  string
	SegmentID            string
	Name                 string
	SiteKind             string
	IsStartSiteCandidate bool
	Address              string
	AddressLatitude      string
	AddressLongitude     string
	AddressPlaceID       string
	AddressPlusCode      string
	SkipGeocoding        bool
}

func newDoJobProductionPlanLocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan location",
		Long: `Create a job production plan location.

Required flags:
  --job-production-plan   Job production plan ID
  --name                  Location name

Optional flags:
  --site-kind                Site kind (job_site, other)
  --is-start-site-candidate  Mark as start site candidate
  --segment                  Job production plan segment ID
  --address                  Full address (will be geocoded)
  --address-latitude         Address latitude
  --address-longitude        Address longitude
  --address-place-id         Google Place ID
  --address-plus-code        Plus code
  --skip-address-geocoding   Skip geocoding the address

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a job production plan location
  xbe do job-production-plan-locations create \
    --job-production-plan 123 \
    --name "Job Site" \
    --site-kind job_site \
    --address "100 Main St, Chicago, IL" \
    --address-latitude 41.8781 \
    --address-longitude -87.6298 \
    --skip-address-geocoding

  # Create an "other" location
  xbe do job-production-plan-locations create \
    --job-production-plan 123 \
    --name "Staging Yard" \
    --site-kind other \
    --is-start-site-candidate`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanLocationsCreate,
	}
	initDoJobProductionPlanLocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanLocationsCmd.AddCommand(newDoJobProductionPlanLocationsCreateCmd())
}

func initDoJobProductionPlanLocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("segment", "", "Job production plan segment ID")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("site-kind", "", "Site kind (job_site, other)")
	cmd.Flags().Bool("is-start-site-candidate", false, "Mark as start site candidate")
	cmd.Flags().String("address", "", "Full address (will be geocoded)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().String("address-place-id", "", "Google Place ID")
	cmd.Flags().String("address-plus-code", "", "Plus code")
	cmd.Flags().Bool("skip-address-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan")
	_ = cmd.MarkFlagRequired("name")
}

func runDoJobProductionPlanLocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanLocationsCreateOptions(cmd)
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

	attributes := map[string]any{
		"name": opts.Name,
	}

	if opts.SiteKind != "" {
		attributes["site-kind"] = opts.SiteKind
	}
	if cmd.Flags().Changed("is-start-site-candidate") {
		attributes["is-start-site-candidate"] = opts.IsStartSiteCandidate
	}
	if opts.Address != "" {
		attributes["address"] = opts.Address
	}
	if opts.AddressLatitude != "" {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if opts.AddressLongitude != "" {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if opts.AddressPlaceID != "" {
		attributes["address-place-id"] = opts.AddressPlaceID
	}
	if opts.AddressPlusCode != "" {
		attributes["address-plus-code"] = opts.AddressPlusCode
	}
	if opts.SkipGeocoding {
		attributes["skip-address-geocoding"] = true
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
	}

	if opts.SegmentID != "" {
		relationships["segment"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-segments",
				"id":   opts.SegmentID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-locations",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-locations", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan location %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanLocationsCreateOptions(cmd *cobra.Command) (doJobProductionPlanLocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	segmentID, _ := cmd.Flags().GetString("segment")
	name, _ := cmd.Flags().GetString("name")
	siteKind, _ := cmd.Flags().GetString("site-kind")
	isStartSiteCandidate, _ := cmd.Flags().GetBool("is-start-site-candidate")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	addressPlaceID, _ := cmd.Flags().GetString("address-place-id")
	addressPlusCode, _ := cmd.Flags().GetString("address-plus-code")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-address-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanLocationsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		JobProductionPlanID:  jobProductionPlanID,
		SegmentID:            segmentID,
		Name:                 name,
		SiteKind:             siteKind,
		IsStartSiteCandidate: isStartSiteCandidate,
		Address:              address,
		AddressLatitude:      addressLatitude,
		AddressLongitude:     addressLongitude,
		AddressPlaceID:       addressPlaceID,
		AddressPlusCode:      addressPlusCode,
		SkipGeocoding:        skipGeocoding,
	}, nil
}
