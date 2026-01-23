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

type doMaterialSitesCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	Name                 string
	MaterialSupplierID   string
	ParentID             string
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
}

func newDoMaterialSitesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new material site",
		Long: `Create a new material site.

Required flags:
  --name               Material site name
  --material-supplier  Material supplier ID (required)

Optional flags:
  --parent                     Parent material site ID
  --phone-number               Contact phone number
  --cb-channel                 CB radio channel
  --hours-description          Operating hours description
  --notes                      Notes about the site
  --color-hex                  Display color (hex format)
  --operating-status           Operating status (active, inactive)
  --is-ticket-maker            Site can create tickets
  --has-scale                  Site has a scale
  --can-be-job-material-site   Can be used as job material site
  --can-be-start-site          Can be used as start site
  --is-portable                Site is portable
  --is-productive              Site is productive
  --address                    Full address (will be geocoded)
  --address-latitude           Address latitude
  --address-longitude          Address longitude
  --skip-geocoding             Skip geocoding the address`,
		Example: `  # Create a basic material site
  xbe do material-sites create --name "Main Plant" --material-supplier 123

  # Create with full details
  xbe do material-sites create --name "Quarry A" --material-supplier 123 \
    --has-scale --is-ticket-maker \
    --address "100 Quarry Rd, Springfield, IL"`,
		RunE: runDoMaterialSitesCreate,
	}
	initDoMaterialSitesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSitesCmd.AddCommand(newDoMaterialSitesCreateCmd())
}

func initDoMaterialSitesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Material site name (required)")
	cmd.Flags().String("material-supplier", "", "Material supplier ID (required)")
	cmd.Flags().String("parent", "", "Parent material site ID")
	cmd.Flags().String("phone-number", "", "Contact phone number")
	cmd.Flags().String("cb-channel", "", "CB radio channel")
	cmd.Flags().String("hours-description", "", "Operating hours description")
	cmd.Flags().String("notes", "", "Notes about the site")
	cmd.Flags().String("color-hex", "", "Display color (hex format)")
	cmd.Flags().String("operating-status", "", "Operating status (active, inactive)")
	cmd.Flags().Bool("is-ticket-maker", false, "Site can create tickets")
	cmd.Flags().Bool("has-scale", false, "Site has a scale")
	cmd.Flags().Bool("can-be-job-material-site", true, "Can be used as job material site")
	cmd.Flags().Bool("can-be-start-site", true, "Can be used as start site")
	cmd.Flags().Bool("is-portable", false, "Site is portable")
	cmd.Flags().Bool("is-productive", true, "Site is productive")
	cmd.Flags().String("address", "", "Full address (will be geocoded)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("material-supplier")
}

func runDoMaterialSitesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSitesCreateOptions(cmd)
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

	if opts.PhoneNumber != "" {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if opts.CbChannel != "" {
		attributes["cb-channel"] = opts.CbChannel
	}
	if opts.HoursDescription != "" {
		attributes["hours-description"] = opts.HoursDescription
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.ColorHex != "" {
		attributes["color-hex"] = opts.ColorHex
	}
	if opts.OperatingStatus != "" {
		attributes["operating-status"] = opts.OperatingStatus
	}
	if cmd.Flags().Changed("is-ticket-maker") {
		attributes["is-ticket-maker"] = opts.IsTicketMaker
	}
	if cmd.Flags().Changed("has-scale") {
		attributes["has-scale"] = opts.HasScale
	}
	if cmd.Flags().Changed("can-be-job-material-site") {
		attributes["can-be-job-material-site"] = opts.CanBeJobMaterialSite
	}
	if cmd.Flags().Changed("can-be-start-site") {
		attributes["can-be-start-site"] = opts.CanBeStartSite
	}
	if cmd.Flags().Changed("is-portable") {
		attributes["is-portable"] = opts.IsPortable
	}
	if cmd.Flags().Changed("is-productive") {
		attributes["is-productive"] = opts.IsProductive
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
		"material-supplier": map[string]any{
			"data": map[string]any{
				"type": "material-suppliers",
				"id":   opts.MaterialSupplierID,
			},
		},
	}

	if opts.ParentID != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.ParentID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-sites",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-sites", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material site %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoMaterialSitesCreateOptions(cmd *cobra.Command) (doMaterialSitesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier")
	parentID, _ := cmd.Flags().GetString("parent")
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

	return doMaterialSitesCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		Name:                 name,
		MaterialSupplierID:   materialSupplierID,
		ParentID:             parentID,
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

func materialSiteRowFromSingle(resp jsonAPISingleResponse) materialSiteRow {
	return materialSiteRow{
		ID:   resp.Data.ID,
		Name: stringAttr(resp.Data.Attributes, "name"),
	}
}
