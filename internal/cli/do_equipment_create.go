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

type doEquipmentCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	Nickname                  string
	IsActive                  bool
	MobilizationMethod        string
	SerialNumber              string
	ManufacturerName          string
	ModelDescription          string
	Year                      string
	Description               string
	IsOffRoad                 bool
	GroupName                 string
	WeightLbs                 string
	ColorHex                  string
	IsAvailable               bool
	EquipmentClassificationID string
	OrganizationType          string
	OrganizationID            string
	TractorID                 string
	TrailerID                 string
}

func newDoEquipmentCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create new equipment",
		Long: `Create new equipment.

Required flags:
  --nickname                   Equipment nickname (required)
  --equipment-classification   Equipment classification ID (required)
  --organization-type          Organization type (e.g., brokers, customers) (required)
  --organization-id            Organization ID (required)

Optional flags:
  --is-active              Whether equipment is active (default true)
  --mobilization-method    Mobilization method
  --serial-number          Serial number
  --manufacturer-name      Manufacturer name
  --model-description      Model description
  --year                   Year of manufacture
  --description            Description
  --is-off-road            Whether equipment is off-road
  --group-name             Group name
  --weight-lbs             Weight in pounds
  --color-hex              Color hex code
  --is-available           Whether equipment is available
  --tractor                Tractor ID to link
  --trailer                Trailer ID to link`,
		Example: `  # Create equipment with required fields
  xbe do equipment create \
    --nickname "Excavator 1" \
    --equipment-classification 123 \
    --organization-type brokers \
    --organization-id 456

  # Create equipment with optional fields
  xbe do equipment create \
    --nickname "Excavator 1" \
    --equipment-classification 123 \
    --organization-type brokers \
    --organization-id 456 \
    --serial-number "SN-12345" \
    --manufacturer-name "Caterpillar" \
    --year "2020"`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentCreate,
	}
	initDoEquipmentCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentCmd.AddCommand(newDoEquipmentCreateCmd())
}

func initDoEquipmentCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("nickname", "", "Equipment nickname (required)")
	cmd.Flags().Bool("is-active", true, "Whether equipment is active")
	cmd.Flags().String("mobilization-method", "", "Mobilization method")
	cmd.Flags().String("serial-number", "", "Serial number")
	cmd.Flags().String("manufacturer-name", "", "Manufacturer name")
	cmd.Flags().String("model-description", "", "Model description")
	cmd.Flags().String("year", "", "Year of manufacture")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().Bool("is-off-road", false, "Whether equipment is off-road")
	cmd.Flags().String("group-name", "", "Group name")
	cmd.Flags().String("weight-lbs", "", "Weight in pounds")
	cmd.Flags().String("color-hex", "", "Color hex code")
	cmd.Flags().Bool("is-available", true, "Whether equipment is available")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID (required)")
	cmd.Flags().String("organization-type", "", "Organization type (e.g., brokers, customers) (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("tractor", "", "Tractor ID to link")
	cmd.Flags().String("trailer", "", "Trailer ID to link")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentCreateOptions(cmd)
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

	if opts.Nickname == "" {
		err := fmt.Errorf("--nickname is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.EquipmentClassificationID == "" {
		err := fmt.Errorf("--equipment-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"nickname":  opts.Nickname,
		"is-active": opts.IsActive,
	}

	if opts.MobilizationMethod != "" {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}
	if opts.SerialNumber != "" {
		attributes["serial-number"] = opts.SerialNumber
	}
	if opts.ManufacturerName != "" {
		attributes["manufacturer-name"] = opts.ManufacturerName
	}
	if opts.ModelDescription != "" {
		attributes["model-description"] = opts.ModelDescription
	}
	if opts.Year != "" {
		attributes["year"] = opts.Year
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("is-off-road") {
		attributes["is-off-road"] = opts.IsOffRoad
	}
	if opts.GroupName != "" {
		attributes["group-name"] = opts.GroupName
	}
	if opts.WeightLbs != "" {
		attributes["weight-lbs"] = opts.WeightLbs
	}
	if opts.ColorHex != "" {
		attributes["color-hex"] = opts.ColorHex
	}
	if cmd.Flags().Changed("is-available") {
		attributes["is-available"] = opts.IsAvailable
	}

	relationships := map[string]any{
		"equipment-classification": map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassificationID,
			},
		},
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	if opts.TractorID != "" {
		relationships["tractor"] = map[string]any{
			"data": map[string]any{
				"type": "tractors",
				"id":   opts.TractorID,
			},
		}
	}

	if opts.TrailerID != "" {
		relationships["trailer"] = map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.TrailerID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment", jsonBody)
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

	row := buildEquipmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment %s\n", row.ID)
	return nil
}

func parseDoEquipmentCreateOptions(cmd *cobra.Command) (doEquipmentCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	nickname, _ := cmd.Flags().GetString("nickname")
	isActive, _ := cmd.Flags().GetBool("is-active")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	serialNumber, _ := cmd.Flags().GetString("serial-number")
	manufacturerName, _ := cmd.Flags().GetString("manufacturer-name")
	modelDescription, _ := cmd.Flags().GetString("model-description")
	year, _ := cmd.Flags().GetString("year")
	description, _ := cmd.Flags().GetString("description")
	isOffRoad, _ := cmd.Flags().GetBool("is-off-road")
	groupName, _ := cmd.Flags().GetString("group-name")
	weightLbs, _ := cmd.Flags().GetString("weight-lbs")
	colorHex, _ := cmd.Flags().GetString("color-hex")
	isAvailable, _ := cmd.Flags().GetBool("is-available")
	equipmentClassificationID, _ := cmd.Flags().GetString("equipment-classification")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	tractorID, _ := cmd.Flags().GetString("tractor")
	trailerID, _ := cmd.Flags().GetString("trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		Nickname:                  nickname,
		IsActive:                  isActive,
		MobilizationMethod:        mobilizationMethod,
		SerialNumber:              serialNumber,
		ManufacturerName:          manufacturerName,
		ModelDescription:          modelDescription,
		Year:                      year,
		Description:               description,
		IsOffRoad:                 isOffRoad,
		GroupName:                 groupName,
		WeightLbs:                 weightLbs,
		ColorHex:                  colorHex,
		IsAvailable:               isAvailable,
		EquipmentClassificationID: equipmentClassificationID,
		OrganizationType:          organizationType,
		OrganizationID:            organizationID,
		TractorID:                 tractorID,
		TrailerID:                 trailerID,
	}, nil
}

func buildEquipmentRowFromSingle(resp jsonAPISingleResponse) equipmentRow {
	attrs := resp.Data.Attributes

	row := equipmentRow{
		ID:           resp.Data.ID,
		Nickname:     stringAttr(attrs, "nickname"),
		SerialNumber: stringAttr(attrs, "serial-number"),
		IsActive:     boolAttr(attrs, "is-active"),
	}

	if rel, ok := resp.Data.Relationships["equipment-classification"]; ok && rel.Data != nil {
		row.EquipmentClassificationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}
