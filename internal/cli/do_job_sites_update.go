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

type doJobSitesUpdateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	ID                             string
	Name                           string
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

func newDoJobSitesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job site",
		Long: `Update a job site.

Optional flags:
  --name                               Job site name
  --notes                              Notes about the job site
  --phone-number                       Contact phone number
  --contact-name                       Contact person name
  --active                             Set as active
  --no-active                          Set as inactive
  --default-time-card-approval-process Default approval process (admin, field)
  --address                            Full address (will be geocoded)
  --address-latitude                   Address latitude
  --address-longitude                  Address longitude
  --skip-geocoding                     Skip geocoding the address`,
		Example: `  # Update job site name
  xbe do job-sites update 123 --name "New Name"

  # Deactivate a job site
  xbe do job-sites update 123 --no-active

  # Update address
  xbe do job-sites update 123 --address "456 Oak Ave, Springfield, IL"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobSitesUpdate,
	}
	initDoJobSitesUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobSitesCmd.AddCommand(newDoJobSitesUpdateCmd())
}

func initDoJobSitesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Job site name")
	cmd.Flags().String("notes", "", "Notes about the job site")
	cmd.Flags().String("phone-number", "", "Contact phone number")
	cmd.Flags().String("contact-name", "", "Contact person name")
	cmd.Flags().Bool("active", false, "Set as active")
	cmd.Flags().Bool("no-active", false, "Set as inactive")
	cmd.Flags().String("default-time-card-approval-process", "", "Default time card approval process (admin, field)")
	cmd.Flags().String("address", "", "Full address (will be geocoded)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobSitesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobSitesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("phone-number") {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if cmd.Flags().Changed("contact-name") {
		attributes["contact-name"] = opts.ContactName
	}
	if cmd.Flags().Changed("active") {
		attributes["is-active"] = true
	}
	if cmd.Flags().Changed("no-active") {
		attributes["is-active"] = false
	}
	if cmd.Flags().Changed("default-time-card-approval-process") {
		attributes["default-time-card-approval-process"] = opts.DefaultTimeCardApprovalProcess
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
	if cmd.Flags().Changed("skip-geocoding") {
		attributes["skip-geocoding"] = opts.SkipGeocoding
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-sites",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-sites/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job site %s\n", row.ID)
	return nil
}

func parseDoJobSitesUpdateOptions(cmd *cobra.Command, args []string) (doJobSitesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
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

	return doJobSitesUpdateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		ID:                             args[0],
		Name:                           name,
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
