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

type doJobSitesCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	Name                           string
	CustomerID                     string
	Notes                          string
	PhoneNumber                    string
	ContactName                    string
	IsActive                       bool
	DefaultTimeCardApprovalProcess string
	Address                        string
	AddressLatitude                string
	AddressLongitude               string
	SkipGeocoding                  bool
}

func newDoJobSitesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new job site",
		Long: `Create a new job site.

Required flags:
  --name        Job site name
  --customer    Customer ID (required)

Optional flags:
  --notes                              Notes about the job site
  --phone-number                       Contact phone number
  --contact-name                       Contact person name
  --active                             Set as active (default: true)
  --default-time-card-approval-process Default approval process (admin, field)
  --address                            Full address (will be geocoded)
  --address-latitude                   Address latitude (use with --skip-geocoding)
  --address-longitude                  Address longitude (use with --skip-geocoding)
  --skip-geocoding                     Skip geocoding the address`,
		Example: `  # Create a job site with address
  xbe do job-sites create --name "Main Street Project" --customer 123 \
    --address "123 Main St, Chicago, IL 60601"

  # Create with contact info
  xbe do job-sites create --name "Highway 42 Site" --customer 123 \
    --address "456 Highway 42, Springfield, IL" \
    --contact-name "John Smith" --phone-number "555-1234"`,
		RunE: runDoJobSitesCreate,
	}
	initDoJobSitesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobSitesCmd.AddCommand(newDoJobSitesCreateCmd())
}

func initDoJobSitesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Job site name (required)")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("notes", "", "Notes about the job site")
	cmd.Flags().String("phone-number", "", "Contact phone number")
	cmd.Flags().String("contact-name", "", "Contact person name")
	cmd.Flags().Bool("active", true, "Set as active")
	cmd.Flags().String("default-time-card-approval-process", "", "Default time card approval process (admin, field)")
	cmd.Flags().String("address", "", "Full address (will be geocoded)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("customer")
}

func runDoJobSitesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobSitesCreateOptions(cmd)
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
		"name":      opts.Name,
		"is-active": opts.IsActive,
	}

	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.PhoneNumber != "" {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if opts.ContactName != "" {
		attributes["contact-name"] = opts.ContactName
	}
	if opts.DefaultTimeCardApprovalProcess != "" {
		attributes["default-time-card-approval-process"] = opts.DefaultTimeCardApprovalProcess
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
	if opts.SkipGeocoding {
		attributes["skip-geocoding"] = true
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-sites",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-sites", jsonBody)
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

	row := jobSiteRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job site %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoJobSitesCreateOptions(cmd *cobra.Command) (doJobSitesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	customerID, _ := cmd.Flags().GetString("customer")
	notes, _ := cmd.Flags().GetString("notes")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	contactName, _ := cmd.Flags().GetString("contact-name")
	isActive, _ := cmd.Flags().GetBool("active")
	defaultApproval, _ := cmd.Flags().GetString("default-time-card-approval-process")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobSitesCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		Name:                           name,
		CustomerID:                     customerID,
		Notes:                          notes,
		PhoneNumber:                    phoneNumber,
		ContactName:                    contactName,
		IsActive:                       isActive,
		DefaultTimeCardApprovalProcess: defaultApproval,
		Address:                        address,
		AddressLatitude:                addressLatitude,
		AddressLongitude:               addressLongitude,
		SkipGeocoding:                  skipGeocoding,
	}, nil
}

func jobSiteRowFromSingle(resp jsonAPISingleResponse) jobSiteRow {
	return jobSiteRow{
		ID:     resp.Data.ID,
		Name:   stringAttr(resp.Data.Attributes, "name"),
		Active: boolAttr(resp.Data.Attributes, "is-active"),
	}
}
