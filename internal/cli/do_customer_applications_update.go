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

type doCustomerApplicationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string

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

func newDoCustomerApplicationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer application",
		Long: `Update a customer application.

Optional:
  --company-name                         Applicant company name
  --company-address                      Company address
  --company-address-latitude             Latitude
  --company-address-longitude            Longitude
  --company-address-place-id             Google Place ID
  --company-address-plus-code            Plus code
  --skip-company-address-geocoding       Skip geocoding (true/false)
  --company-url                          Company website URL
  --notes                                Notes
  --requires-union-drivers               Requires union drivers (true/false)
  --is-trucking-company                  Is trucking company (true/false)
  --estimated-annual-material-transport-spend Estimated annual spend
  --status                               Status (pending, reviewing, denied, approved)
  --broker                               Broker ID
  --user                                 User ID
  --job-types                            Job type IDs (comma-separated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update status
  xbe do customer-applications update 123 --status reviewing

  # Update company address
  xbe do customer-applications update 123 --company-address "456 Market St, Richmond, VA"

  # Update job types
  xbe do customer-applications update 123 --job-types "1,2"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerApplicationsUpdate,
	}
	initDoCustomerApplicationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerApplicationsCmd.AddCommand(newDoCustomerApplicationsUpdateCmd())
}

func initDoCustomerApplicationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("company-name", "", "Applicant company name")
	cmd.Flags().String("company-address", "", "Company address")
	cmd.Flags().String("company-address-latitude", "", "Latitude")
	cmd.Flags().String("company-address-longitude", "", "Longitude")
	cmd.Flags().String("company-address-place-id", "", "Google Place ID")
	cmd.Flags().String("company-address-plus-code", "", "Plus code")
	cmd.Flags().String("skip-company-address-geocoding", "", "Skip geocoding (true/false)")
	cmd.Flags().String("company-url", "", "Company website URL")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("requires-union-drivers", "", "Requires union drivers (true/false)")
	cmd.Flags().String("is-trucking-company", "", "Is trucking company (true/false)")
	cmd.Flags().String("estimated-annual-material-transport-spend", "", "Estimated annual spend")
	cmd.Flags().String("status", "", "Status (pending, reviewing, denied, approved)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("job-types", "", "Job type IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerApplicationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerApplicationsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("company-name") {
		attributes["company-name"] = opts.CompanyName
	}
	if cmd.Flags().Changed("company-address") {
		attributes["company-address"] = opts.CompanyAddress
	}
	if cmd.Flags().Changed("company-address-latitude") {
		attributes["company-address-latitude"] = opts.CompanyAddressLatitude
	}
	if cmd.Flags().Changed("company-address-longitude") {
		attributes["company-address-longitude"] = opts.CompanyAddressLongitude
	}
	if cmd.Flags().Changed("company-address-place-id") {
		attributes["company-address-place-id"] = opts.CompanyAddressPlaceID
	}
	if cmd.Flags().Changed("company-address-plus-code") {
		attributes["company-address-plus-code"] = opts.CompanyAddressPlusCode
	}
	if cmd.Flags().Changed("skip-company-address-geocoding") {
		attributes["skip-company-address-geocoding"] = opts.SkipCompanyAddressGeocoding == "true"
	}
	if cmd.Flags().Changed("company-url") {
		attributes["company-url"] = opts.CompanyURL
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("requires-union-drivers") {
		attributes["requires-union-drivers"] = opts.RequiresUnionDrivers == "true"
	}
	if cmd.Flags().Changed("is-trucking-company") {
		attributes["is-trucking-company"] = opts.IsTruckingCompany == "true"
	}
	if cmd.Flags().Changed("estimated-annual-material-transport-spend") {
		attributes["estimated-annual-material-transport-spend"] = opts.EstimatedAnnualMaterialSpend
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}

	if cmd.Flags().Changed("broker") {
		if strings.TrimSpace(opts.Broker) == "" {
			relationships["broker"] = map[string]any{"data": nil}
		} else {
			relationships["broker"] = map[string]any{
				"data": map[string]any{
					"type": "brokers",
					"id":   opts.Broker,
				},
			}
		}
	}
	if cmd.Flags().Changed("user") {
		if strings.TrimSpace(opts.User) == "" {
			relationships["user"] = map[string]any{"data": nil}
		} else {
			relationships["user"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.User,
				},
			}
		}
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

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "customer-applications",
			"id":   opts.ID,
		},
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-applications/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer application %s\n", row.ID)
	return nil
}

func parseDoCustomerApplicationsUpdateOptions(cmd *cobra.Command, args []string) (doCustomerApplicationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	companyName, _ := cmd.Flags().GetString("company-name")
	companyAddress, _ := cmd.Flags().GetString("company-address")
	companyAddressLatitude, _ := cmd.Flags().GetString("company-address-latitude")
	companyAddressLongitude, _ := cmd.Flags().GetString("company-address-longitude")
	companyAddressPlaceID, _ := cmd.Flags().GetString("company-address-place-id")
	companyAddressPlusCode, _ := cmd.Flags().GetString("company-address-plus-code")
	skipCompanyAddressGeocoding, _ := cmd.Flags().GetString("skip-company-address-geocoding")
	companyURL, _ := cmd.Flags().GetString("company-url")
	notes, _ := cmd.Flags().GetString("notes")
	requiresUnionDrivers, _ := cmd.Flags().GetString("requires-union-drivers")
	isTruckingCompany, _ := cmd.Flags().GetString("is-trucking-company")
	estimatedAnnualMaterialSpend, _ := cmd.Flags().GetString("estimated-annual-material-transport-spend")
	status, _ := cmd.Flags().GetString("status")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	jobTypes, _ := cmd.Flags().GetString("job-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerApplicationsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
		CompanyName:                  companyName,
		CompanyAddress:               companyAddress,
		CompanyAddressLatitude:       companyAddressLatitude,
		CompanyAddressLongitude:      companyAddressLongitude,
		CompanyAddressPlaceID:        companyAddressPlaceID,
		CompanyAddressPlusCode:       companyAddressPlusCode,
		SkipCompanyAddressGeocoding:  skipCompanyAddressGeocoding,
		CompanyURL:                   companyURL,
		Notes:                        notes,
		RequiresUnionDrivers:         requiresUnionDrivers,
		IsTruckingCompany:            isTruckingCompany,
		EstimatedAnnualMaterialSpend: estimatedAnnualMaterialSpend,
		Status:                       status,
		Broker:                       broker,
		User:                         user,
		JobTypes:                     jobTypes,
	}, nil
}
