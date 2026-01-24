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

type doTruckerApplicationsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

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

func newDoTruckerApplicationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trucker application",
		Long: `Create a new trucker application.

Required flags:
  --name             Company name (required)
  --company-address  Company address (required)
  --broker           Broker ID (required)
  --user             User ID (required)

Optional flags:
  --company-address-place-id       Google Place ID
  --company-address-plus-code      Plus code
  --skip-company-address-geocoding Skip geocoding (true/false)
  --has-union-drivers              Has union drivers (true/false)
  --estimated-trailer-capacity     Estimated trailer capacity
  --notes                          Notes
  --referral-code                  Referral code
  --status                         Status: pending, reviewing, denied, approved
  --via-dumptruckloadsdotcom        Via dumptruckloadsdotcom (true/false, broker members only)`,
		Example: `  # Create a trucker application
  xbe do trucker-applications create \\
    --name "Acme Trucking" \\
    --company-address "123 Main St" \\
    --broker 123 \\
    --user 456

  # Create with status
  xbe do trucker-applications create \\
    --name "Acme Trucking" \\
    --company-address "123 Main St" \\
    --broker 123 \\
    --user 456 \\
    --status reviewing

  # Output JSON
  xbe do trucker-applications create --name "Acme Trucking" --company-address "123 Main St" --broker 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTruckerApplicationsCreate,
	}
	initDoTruckerApplicationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerApplicationsCmd.AddCommand(newDoTruckerApplicationsCreateCmd())
}

func initDoTruckerApplicationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Company name (required)")
	cmd.Flags().String("company-address", "", "Company address (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
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

func runDoTruckerApplicationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerApplicationsCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.CompanyAddress == "" {
		err := fmt.Errorf("--company-address is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.User == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"company-name":    opts.Name,
		"company-address": opts.CompanyAddress,
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

	data := map[string]any{
		"type":       "trucker-applications",
		"attributes": attributes,
		"relationships": map[string]any{
			"broker": map[string]any{
				"data": map[string]string{
					"type": "brokers",
					"id":   opts.Broker,
				},
			},
			"user": map[string]any{
				"data": map[string]string{
					"type": "users",
					"id":   opts.User,
				},
			},
		},
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

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-applications", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker application %s (%s)\n", details.ID, details.CompanyName)
	return renderTruckerApplicationDetails(cmd, details)
}

func parseDoTruckerApplicationsCreateOptions(cmd *cobra.Command) (doTruckerApplicationsCreateOptions, error) {
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

	return doTruckerApplicationsCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
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
