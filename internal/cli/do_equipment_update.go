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

type doEquipmentUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
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
	TractorID                 string
	TrailerID                 string
}

func newDoEquipmentUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update equipment",
		Long: `Update equipment.

Optional flags:
  --nickname                   Equipment nickname
  --is-active                  Whether equipment is active
  --mobilization-method        Mobilization method
  --serial-number              Serial number
  --manufacturer-name          Manufacturer name
  --model-description          Model description
  --year                       Year of manufacture
  --description                Description
  --is-off-road                Whether equipment is off-road
  --group-name                 Group name
  --weight-lbs                 Weight in pounds
  --color-hex                  Color hex code
  --is-available               Whether equipment is available
  --equipment-classification   Equipment classification ID
  --tractor                    Tractor ID to link
  --trailer                    Trailer ID to link`,
		Example: `  # Update nickname
  xbe do equipment update 123 --nickname "New Excavator"

  # Update active status
  xbe do equipment update 123 --is-active false

  # Update multiple fields
  xbe do equipment update 123 --serial-number "SN-99999" --year "2021"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentUpdate,
	}
	initDoEquipmentUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentCmd.AddCommand(newDoEquipmentUpdateCmd())
}

func initDoEquipmentUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("nickname", "", "Equipment nickname")
	cmd.Flags().Bool("is-active", false, "Whether equipment is active")
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
	cmd.Flags().Bool("is-available", false, "Whether equipment is available")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("tractor", "", "Tractor ID to link")
	cmd.Flags().String("trailer", "", "Trailer ID to link")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("nickname") {
		attributes["nickname"] = opts.Nickname
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}
	if cmd.Flags().Changed("mobilization-method") {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}
	if cmd.Flags().Changed("serial-number") {
		attributes["serial-number"] = opts.SerialNumber
	}
	if cmd.Flags().Changed("manufacturer-name") {
		attributes["manufacturer-name"] = opts.ManufacturerName
	}
	if cmd.Flags().Changed("model-description") {
		attributes["model-description"] = opts.ModelDescription
	}
	if cmd.Flags().Changed("year") {
		attributes["year"] = opts.Year
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("is-off-road") {
		attributes["is-off-road"] = opts.IsOffRoad
	}
	if cmd.Flags().Changed("group-name") {
		attributes["group-name"] = opts.GroupName
	}
	if cmd.Flags().Changed("weight-lbs") {
		attributes["weight-lbs"] = opts.WeightLbs
	}
	if cmd.Flags().Changed("color-hex") {
		attributes["color-hex"] = opts.ColorHex
	}
	if cmd.Flags().Changed("is-available") {
		attributes["is-available"] = opts.IsAvailable
	}

	if cmd.Flags().Changed("equipment-classification") {
		relationships["equipment-classification"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassificationID,
			},
		}
	}

	if cmd.Flags().Changed("tractor") {
		if opts.TractorID == "" {
			relationships["tractor"] = map[string]any{"data": nil}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]any{
					"type": "tractors",
					"id":   opts.TractorID,
				},
			}
		}
	}

	if cmd.Flags().Changed("trailer") {
		if opts.TrailerID == "" {
			relationships["trailer"] = map[string]any{"data": nil}
		} else {
			relationships["trailer"] = map[string]any{
				"data": map[string]any{
					"type": "trailers",
					"id":   opts.TrailerID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "equipment",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment %s\n", row.ID)
	return nil
}

func parseDoEquipmentUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentUpdateOptions, error) {
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
	tractorID, _ := cmd.Flags().GetString("tractor")
	trailerID, _ := cmd.Flags().GetString("trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
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
		TractorID:                 tractorID,
		TrailerID:                 trailerID,
	}, nil
}
