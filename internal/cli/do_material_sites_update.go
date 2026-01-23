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

type doMaterialSitesUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	Name                 string
	PhoneNumber          string
	CbChannel            string
	HoursDescription     string
	Notes                string
	ColorHex             string
	OperatingStatus      string
	IsTicketMaker        bool
	HasScale             bool
	CanBeJobMaterialSite bool
	CanBeStartSite       bool
	IsPortable           bool
	IsProductive         bool
	Address              string
	AddressLatitude      string
	AddressLongitude     string
	SkipGeocoding        bool
	MaterialSupplierID   string
}

func newDoMaterialSitesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material site",
		Long: `Update a material site.

Optional flags:
  --name                       Material site name
  --material-supplier          Material supplier ID
  --phone-number               Contact phone number
  --cb-channel                 CB radio channel
  --hours-description          Operating hours description
  --notes                      Notes about the site
  --color-hex                  Display color (hex format)
  --operating-status           Operating status (active, inactive)
  --is-ticket-maker            Site can create tickets
  --no-is-ticket-maker         Site cannot create tickets
  --has-scale                  Site has a scale
  --no-has-scale               Site does not have a scale
  --can-be-job-material-site   Can be used as job material site
  --can-be-start-site          Can be used as start site
  --is-portable                Site is portable
  --is-productive              Site is productive
  --address                    Full address (will be geocoded)
  --address-latitude           Address latitude
  --address-longitude          Address longitude
  --skip-geocoding             Skip geocoding the address`,
		Example: `  # Update material site name
  xbe do material-sites update 123 --name "New Name"

  # Update operating status
  xbe do material-sites update 123 --operating-status "inactive"

  # Enable scale
  xbe do material-sites update 123 --has-scale`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialSitesUpdate,
	}
	initDoMaterialSitesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSitesCmd.AddCommand(newDoMaterialSitesUpdateCmd())
}

func initDoMaterialSitesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Material site name")
	cmd.Flags().String("material-supplier", "", "Material supplier ID")
	cmd.Flags().String("phone-number", "", "Contact phone number")
	cmd.Flags().String("cb-channel", "", "CB radio channel")
	cmd.Flags().String("hours-description", "", "Operating hours description")
	cmd.Flags().String("notes", "", "Notes about the site")
	cmd.Flags().String("color-hex", "", "Display color (hex format)")
	cmd.Flags().String("operating-status", "", "Operating status (active, inactive)")
	cmd.Flags().Bool("is-ticket-maker", false, "Site can create tickets")
	cmd.Flags().Bool("no-is-ticket-maker", false, "Site cannot create tickets")
	cmd.Flags().Bool("has-scale", false, "Site has a scale")
	cmd.Flags().Bool("no-has-scale", false, "Site does not have a scale")
	cmd.Flags().Bool("can-be-job-material-site", false, "Can be used as job material site")
	cmd.Flags().Bool("no-can-be-job-material-site", false, "Cannot be used as job material site")
	cmd.Flags().Bool("can-be-start-site", false, "Can be used as start site")
	cmd.Flags().Bool("no-can-be-start-site", false, "Cannot be used as start site")
	cmd.Flags().Bool("is-portable", false, "Site is portable")
	cmd.Flags().Bool("no-is-portable", false, "Site is not portable")
	cmd.Flags().Bool("is-productive", false, "Site is productive")
	cmd.Flags().Bool("no-is-productive", false, "Site is not productive")
	cmd.Flags().String("address", "", "Full address (will be geocoded)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSitesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSitesUpdateOptions(cmd, args)
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
	var relationships map[string]any

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("phone-number") {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if cmd.Flags().Changed("cb-channel") {
		attributes["cb-channel"] = opts.CbChannel
	}
	if cmd.Flags().Changed("hours-description") {
		attributes["hours-description"] = opts.HoursDescription
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("color-hex") {
		attributes["color-hex"] = opts.ColorHex
	}
	if cmd.Flags().Changed("operating-status") {
		attributes["operating-status"] = opts.OperatingStatus
	}
	if cmd.Flags().Changed("is-ticket-maker") {
		attributes["is-ticket-maker"] = true
	}
	if cmd.Flags().Changed("no-is-ticket-maker") {
		attributes["is-ticket-maker"] = false
	}
	if cmd.Flags().Changed("has-scale") {
		attributes["has-scale"] = true
	}
	if cmd.Flags().Changed("no-has-scale") {
		attributes["has-scale"] = false
	}
	if cmd.Flags().Changed("can-be-job-material-site") {
		attributes["can-be-job-material-site"] = true
	}
	if cmd.Flags().Changed("no-can-be-job-material-site") {
		attributes["can-be-job-material-site"] = false
	}
	if cmd.Flags().Changed("can-be-start-site") {
		attributes["can-be-start-site"] = true
	}
	if cmd.Flags().Changed("no-can-be-start-site") {
		attributes["can-be-start-site"] = false
	}
	if cmd.Flags().Changed("is-portable") {
		attributes["is-portable"] = true
	}
	if cmd.Flags().Changed("no-is-portable") {
		attributes["is-portable"] = false
	}
	if cmd.Flags().Changed("is-productive") {
		attributes["is-productive"] = true
	}
	if cmd.Flags().Changed("no-is-productive") {
		attributes["is-productive"] = false
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

	if cmd.Flags().Changed("material-supplier") {
		relationships = map[string]any{
			"material-supplier": map[string]any{
				"data": map[string]any{
					"type": "material-suppliers",
					"id":   opts.MaterialSupplierID,
				},
			},
		}
	}

	if len(attributes) == 0 && relationships == nil {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "material-sites",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if relationships != nil {
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-sites/"+opts.ID, jsonBody)
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

	row := materialSiteRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material site %s\n", row.ID)
	return nil
}

func parseDoMaterialSitesUpdateOptions(cmd *cobra.Command, args []string) (doMaterialSitesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	cbChannel, _ := cmd.Flags().GetString("cb-channel")
	hoursDescription, _ := cmd.Flags().GetString("hours-description")
	notes, _ := cmd.Flags().GetString("notes")
	colorHex, _ := cmd.Flags().GetString("color-hex")
	operatingStatus, _ := cmd.Flags().GetString("operating-status")
	isTicketMaker, _ := cmd.Flags().GetBool("is-ticket-maker")
	hasScale, _ := cmd.Flags().GetBool("has-scale")
	canBeJobMaterialSite, _ := cmd.Flags().GetBool("can-be-job-material-site")
	canBeStartSite, _ := cmd.Flags().GetBool("can-be-start-site")
	isPortable, _ := cmd.Flags().GetBool("is-portable")
	isProductive, _ := cmd.Flags().GetBool("is-productive")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSitesUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		Name:                 name,
		MaterialSupplierID:   materialSupplierID,
		PhoneNumber:          phoneNumber,
		CbChannel:            cbChannel,
		HoursDescription:     hoursDescription,
		Notes:                notes,
		ColorHex:             colorHex,
		OperatingStatus:      operatingStatus,
		IsTicketMaker:        isTicketMaker,
		HasScale:             hasScale,
		CanBeJobMaterialSite: canBeJobMaterialSite,
		CanBeStartSite:       canBeStartSite,
		IsPortable:           isPortable,
		IsProductive:         isProductive,
		Address:              address,
		AddressLatitude:      addressLatitude,
		AddressLongitude:     addressLongitude,
		SkipGeocoding:        skipGeocoding,
	}, nil
}
