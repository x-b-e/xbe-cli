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

type doCustomerApplicationsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	CompanyName    string
	Broker         string
	User           string
	Status         string
	CompanyURL     string
	Notes          string
	CompanyAddress string

	CompanyAddressLatitude  string
	CompanyAddressLongitude string
	CompanyAddressPlaceID   string
	CompanyAddressPlusCode  string

	SkipCompanyAddressGeocoding  string
	RequiresUnionDrivers         string
	IsTruckingCompany            string
	EstimatedAnnualMaterialSpend string
	JobTypes                     string
}

func newDoCustomerApplicationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer application",
		Long: `Create a new customer application.

Required:
  --company-name             Applicant company name
  --company-address          Company address
  --requires-union-drivers   Requires union drivers (true/false)
  --is-trucking-company      Is trucking company (true/false)
  --broker                   Broker ID
  --user                     User ID

Optional:
  --status                                 Status (pending, reviewing, denied, approved)
  --company-url                             Company website URL
  --notes                                   Notes
  --company-address-latitude                Latitude
  --company-address-longitude               Longitude
  --company-address-place-id                Google Place ID
  --company-address-plus-code               Plus code
  --skip-company-address-geocoding          Skip geocoding (true/false)
  --estimated-annual-material-transport-spend Estimated annual spend
  --job-types                               Job type IDs (comma-separated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a customer application
  xbe do customer-applications create \
    --company-name "Acme Paving" \
    --company-address "123 Main St, Richmond, VA" \
    --requires-union-drivers false \
    --is-trucking-company true \
    --broker 123 \
    --user 456

  # Create with job types
  xbe do customer-applications create \
    --company-name "Acme Paving" \
    --company-address "123 Main St, Richmond, VA" \
    --requires-union-drivers false \
    --is-trucking-company true \
    --broker 123 \
    --user 456 \
    --job-types "1,2"`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerApplicationsCreate,
	}
	initDoCustomerApplicationsCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerApplicationsCmd.AddCommand(newDoCustomerApplicationsCreateCmd())
}

func initDoCustomerApplicationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("company-name", "", "Applicant company name (required)")
	cmd.Flags().String("company-address", "", "Company address (required)")
	cmd.Flags().String("requires-union-drivers", "", "Requires union drivers (true/false, required)")
	cmd.Flags().String("is-trucking-company", "", "Is trucking company (true/false, required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("status", "", "Status (pending, reviewing, denied, approved)")
	cmd.Flags().String("company-url", "", "Company website URL")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("company-address-latitude", "", "Latitude")
	cmd.Flags().String("company-address-longitude", "", "Longitude")
	cmd.Flags().String("company-address-place-id", "", "Google Place ID")
	cmd.Flags().String("company-address-plus-code", "", "Plus code")
	cmd.Flags().String("skip-company-address-geocoding", "", "Skip geocoding (true/false)")
	cmd.Flags().String("estimated-annual-material-transport-spend", "", "Estimated annual spend")
	cmd.Flags().String("job-types", "", "Job type IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("company-name")
	_ = cmd.MarkFlagRequired("company-address")
	_ = cmd.MarkFlagRequired("requires-union-drivers")
	_ = cmd.MarkFlagRequired("is-trucking-company")
	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("user")
}

func runDoCustomerApplicationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerApplicationsCreateOptions(cmd)
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

	if opts.CompanyName == "" {
		err := fmt.Errorf("--company-name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.CompanyAddress == "" {
		err := fmt.Errorf("--company-address is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.RequiresUnionDrivers == "" {
		err := fmt.Errorf("--requires-union-drivers is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.IsTruckingCompany == "" {
		err := fmt.Errorf("--is-trucking-company is required")
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
		"company-name":           opts.CompanyName,
		"company-address":        opts.CompanyAddress,
		"requires-union-drivers": opts.RequiresUnionDrivers == "true",
		"is-trucking-company":    opts.IsTruckingCompany == "true",
	}

	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.CompanyURL != "" {
		attributes["company-url"] = opts.CompanyURL
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.CompanyAddressLatitude != "" {
		attributes["company-address-latitude"] = opts.CompanyAddressLatitude
	}
	if opts.CompanyAddressLongitude != "" {
		attributes["company-address-longitude"] = opts.CompanyAddressLongitude
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
	if opts.EstimatedAnnualMaterialSpend != "" {
		attributes["estimated-annual-material-transport-spend"] = opts.EstimatedAnnualMaterialSpend
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	if cmd.Flags().Changed("job-types") {
		if strings.TrimSpace(opts.JobTypes) == "" {
			relationships["job-types"] = map[string]any{"data": []any{}}
		} else {
			ids := strings.Split(opts.JobTypes, ",")
			data := make([]map[string]any, 0, len(ids))
			for _, id := range ids {
				trimmed := strings.TrimSpace(id)
				if trimmed == "" {
					continue
				}
				data = append(data, map[string]any{
					"type": "job-types",
					"id":   trimmed,
				})
			}
			relationships["job-types"] = map[string]any{"data": data}
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-applications",
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

	body, _, err := client.Post(cmd.Context(), "/v1/customer-applications", jsonBody)
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

	row := buildCustomerApplicationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer application %s\n", row.ID)
	return nil
}

func parseDoCustomerApplicationsCreateOptions(cmd *cobra.Command) (doCustomerApplicationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	companyName, _ := cmd.Flags().GetString("company-name")
	companyAddress, _ := cmd.Flags().GetString("company-address")
	requiresUnionDrivers, _ := cmd.Flags().GetString("requires-union-drivers")
	isTruckingCompany, _ := cmd.Flags().GetString("is-trucking-company")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	status, _ := cmd.Flags().GetString("status")
	companyURL, _ := cmd.Flags().GetString("company-url")
	notes, _ := cmd.Flags().GetString("notes")
	companyAddressLatitude, _ := cmd.Flags().GetString("company-address-latitude")
	companyAddressLongitude, _ := cmd.Flags().GetString("company-address-longitude")
	companyAddressPlaceID, _ := cmd.Flags().GetString("company-address-place-id")
	companyAddressPlusCode, _ := cmd.Flags().GetString("company-address-plus-code")
	skipCompanyAddressGeocoding, _ := cmd.Flags().GetString("skip-company-address-geocoding")
	estimatedAnnualMaterialSpend, _ := cmd.Flags().GetString("estimated-annual-material-transport-spend")
	jobTypes, _ := cmd.Flags().GetString("job-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerApplicationsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		CompanyName:                  companyName,
		CompanyAddress:               companyAddress,
		RequiresUnionDrivers:         requiresUnionDrivers,
		IsTruckingCompany:            isTruckingCompany,
		Broker:                       broker,
		User:                         user,
		Status:                       status,
		CompanyURL:                   companyURL,
		Notes:                        notes,
		CompanyAddressLatitude:       companyAddressLatitude,
		CompanyAddressLongitude:      companyAddressLongitude,
		CompanyAddressPlaceID:        companyAddressPlaceID,
		CompanyAddressPlusCode:       companyAddressPlusCode,
		SkipCompanyAddressGeocoding:  skipCompanyAddressGeocoding,
		EstimatedAnnualMaterialSpend: estimatedAnnualMaterialSpend,
		JobTypes:                     jobTypes,
	}, nil
}

func buildCustomerApplicationRowFromSingle(resp jsonAPISingleResponse) customerApplicationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	row := customerApplicationRow{
		ID:          resp.Data.ID,
		CompanyName: strings.TrimSpace(stringAttr(resp.Data.Attributes, "company-name")),
		Status:      stringAttr(resp.Data.Attributes, "status"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			if row.UserName == "" {
				row.UserName = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
			}
		}
	}

	return row
}
