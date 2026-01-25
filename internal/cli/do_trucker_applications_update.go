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

type doTruckerApplicationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string

	Name                        string
	CompanyAddress              string
	Broker                      string
	User                        string
	CompanyAddressPlaceID       string
	CompanyAddressPlusCode      string
	SkipCompanyAddressGeocoding string
	HasUnionDrivers             string
	EstimatedTrailerCapacity    string
	Notes                       string
	ReferralCode                string
	ViaDumptruckloadsdotcom     string
	Status                      string
}

func newDoTruckerApplicationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker application",
		Long: `Update a trucker application.

Arguments:
  <id>  The trucker application ID (required).

Optional flags:
  --name                          Company name
  --company-address               Company address
  --broker                        Broker ID
  --user                          User ID
  --company-address-place-id      Google Place ID
  --company-address-plus-code     Plus code
  --skip-company-address-geocoding Skip geocoding (true/false)
  --has-union-drivers             Has union drivers (true/false)
  --estimated-trailer-capacity    Estimated trailer capacity
  --notes                         Notes
  --referral-code                 Referral code
  --status                        Status: pending, reviewing, denied, approved
  --via-dumptruckloadsdotcom       Via dumptruckloadsdotcom (true/false, broker members only)`,
		Example: `  # Update a trucker application
  xbe do trucker-applications update 123 --status reviewing

  # Update company details
  xbe do trucker-applications update 123 --name "New Name" --company-address "456 Oak Ave"

  # Output JSON
  xbe do trucker-applications update 123 --status denied --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerApplicationsUpdate,
	}
	initDoTruckerApplicationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerApplicationsCmd.AddCommand(newDoTruckerApplicationsUpdateCmd())
}

func initDoTruckerApplicationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Company name")
	cmd.Flags().String("company-address", "", "Company address")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("company-address-place-id", "", "Google Place ID")
	cmd.Flags().String("company-address-plus-code", "", "Plus code")
	cmd.Flags().String("skip-company-address-geocoding", "", "Skip geocoding (true/false)")
	cmd.Flags().String("has-union-drivers", "", "Has union drivers (true/false)")
	cmd.Flags().String("estimated-trailer-capacity", "", "Estimated trailer capacity")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("referral-code", "", "Referral code")
	cmd.Flags().String("status", "", "Status: pending, reviewing, denied, approved")
	cmd.Flags().String("via-dumptruckloadsdotcom", "", "Via dumptruckloadsdotcom (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerApplicationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerApplicationsUpdateOptions(cmd, args)
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

	if opts.ID == "" {
		err := fmt.Errorf("trucker application id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["company-name"] = opts.Name
	}
	if opts.CompanyAddress != "" {
		attributes["company-address"] = opts.CompanyAddress
	}
	if opts.CompanyAddressPlaceID != "" {
		attributes["company-address-place-id"] = opts.CompanyAddressPlaceID
	}
	if opts.CompanyAddressPlusCode != "" {
		attributes["company-address-plus-code"] = opts.CompanyAddressPlusCode
	}
	if opts.SkipCompanyAddressGeocoding != "" {
		attributes["skip-company-address-geocoding"] = opts.SkipCompanyAddressGeocoding == "true"
	}
	if opts.HasUnionDrivers != "" {
		attributes["has-union-drivers"] = opts.HasUnionDrivers == "true"
	}
	if opts.EstimatedTrailerCapacity != "" {
		attributes["estimated-trailer-capacity"] = opts.EstimatedTrailerCapacity
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.ReferralCode != "" {
		attributes["referral-code"] = opts.ReferralCode
	}
	if opts.ViaDumptruckloadsdotcom != "" {
		attributes["via-dumptruckloadsdotcom"] = opts.ViaDumptruckloadsdotcom == "true"
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}

	relationships := map[string]any{}
	if opts.Broker != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]string{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if opts.User != "" {
		relationships["user"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.User,
			},
		}
	}

	data := map[string]any{
		"type": "trucker-applications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-applications/"+opts.ID, jsonBody)
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

	details := buildTruckerApplicationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker application %s (%s)\n", details.ID, details.CompanyName)
	return renderTruckerApplicationDetails(cmd, details)
}

func parseDoTruckerApplicationsUpdateOptions(cmd *cobra.Command, args []string) (doTruckerApplicationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	companyAddress, _ := cmd.Flags().GetString("company-address")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	companyAddressPlaceID, _ := cmd.Flags().GetString("company-address-place-id")
	companyAddressPlusCode, _ := cmd.Flags().GetString("company-address-plus-code")
	skipCompanyAddressGeocoding, _ := cmd.Flags().GetString("skip-company-address-geocoding")
	hasUnionDrivers, _ := cmd.Flags().GetString("has-union-drivers")
	estimatedTrailerCapacity, _ := cmd.Flags().GetString("estimated-trailer-capacity")
	notes, _ := cmd.Flags().GetString("notes")
	referralCode, _ := cmd.Flags().GetString("referral-code")
	viaDumptruckloadsdotcom, _ := cmd.Flags().GetString("via-dumptruckloadsdotcom")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerApplicationsUpdateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ID:                          args[0],
		Name:                        name,
		CompanyAddress:              companyAddress,
		Broker:                      broker,
		User:                        user,
		CompanyAddressPlaceID:       companyAddressPlaceID,
		CompanyAddressPlusCode:      companyAddressPlusCode,
		SkipCompanyAddressGeocoding: skipCompanyAddressGeocoding,
		HasUnionDrivers:             hasUnionDrivers,
		EstimatedTrailerCapacity:    estimatedTrailerCapacity,
		Notes:                       notes,
		ReferralCode:                referralCode,
		ViaDumptruckloadsdotcom:     viaDumptruckloadsdotcom,
		Status:                      status,
	}, nil
}
