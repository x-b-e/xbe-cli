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

type doCrewRatesUpdateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ID                         string
	Description                string
	PricePerUnit               string
	StartOn                    string
	EndOn                      string
	IsActive                   string
	Broker                     string
	ResourceType               string
	ResourceID                 string
	ResourceClassificationType string
	ResourceClassificationID   string
	CraftClass                 string
}

func newDoCrewRatesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a crew rate",
		Long: `Update a crew rate.

Optional:
  --description                  Description
  --price-per-unit               Price per unit
  --start-on                     Start date (YYYY-MM-DD)
  --end-on                       End date (YYYY-MM-DD)
  --is-active                    Active status (true/false)
  --broker                       Broker ID
  --resource-type                Resource type (Laborer, Equipment)
  --resource-id                  Resource ID
  --resource-classification-type Resource classification type (LaborClassification, EquipmentClassification)
  --resource-classification-id   Resource classification ID
  --craft-class                  Craft class ID`,
		Example: `  # Update price and end date
  xbe do crew-rates update 123 --price-per-unit 85.00 --end-on 2025-12-31

  # Update active status
  xbe do crew-rates update 123 --is-active false

  # Update resource association
  xbe do crew-rates update 123 --resource-type Equipment --resource-id 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCrewRatesUpdate,
	}
	initDoCrewRatesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCrewRatesCmd.AddCommand(newDoCrewRatesUpdateCmd())
}

func initDoCrewRatesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("price-per-unit", "", "Price per unit")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("is-active", "", "Active status (true/false)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("resource-type", "", "Resource type (Laborer, Equipment)")
	cmd.Flags().String("resource-id", "", "Resource ID")
	cmd.Flags().String("resource-classification-type", "", "Resource classification type (LaborClassification, EquipmentClassification)")
	cmd.Flags().String("resource-classification-id", "", "Resource classification ID")
	cmd.Flags().String("craft-class", "", "Craft class ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCrewRatesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCrewRatesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("price-per-unit") {
		attributes["price-per-unit"] = opts.PricePerUnit
	}
	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
	}
	if cmd.Flags().Changed("end-on") {
		attributes["end-on"] = opts.EndOn
	}
	if cmd.Flags().Changed("is-active") {
		if opts.IsActive == "" {
			err := fmt.Errorf("--is-active must be true or false")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		isActive, err := parseCrewRateBool(opts.IsActive, "is-active")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["is-active"] = isActive
	}

	if cmd.Flags().Changed("broker") {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}

	resourceTypeChanged := cmd.Flags().Changed("resource-type")
	resourceIDChanged := cmd.Flags().Changed("resource-id")
	if resourceTypeChanged || resourceIDChanged {
		if opts.ResourceType == "" || opts.ResourceID == "" {
			err := fmt.Errorf("--resource-type and --resource-id must be provided together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		resourceType, err := parseCrewRateResourceType(opts.ResourceType)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["resource"] = map[string]any{
			"data": map[string]any{
				"type": resourceType,
				"id":   opts.ResourceID,
			},
		}
	}

	resourceClassTypeChanged := cmd.Flags().Changed("resource-classification-type")
	resourceClassIDChanged := cmd.Flags().Changed("resource-classification-id")
	if resourceClassTypeChanged || resourceClassIDChanged {
		if opts.ResourceClassificationType == "" || opts.ResourceClassificationID == "" {
			err := fmt.Errorf("--resource-classification-type and --resource-classification-id must be provided together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		resourceClassType, err := parseCrewRateResourceClassificationType(opts.ResourceClassificationType)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["resource-classification"] = map[string]any{
			"data": map[string]any{
				"type": resourceClassType,
				"id":   opts.ResourceClassificationID,
			},
		}
	}

	if cmd.Flags().Changed("craft-class") {
		relationships["craft-class"] = map[string]any{
			"data": map[string]any{
				"type": "craft-classes",
				"id":   opts.CraftClass,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "crew-rates",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/crew-rates/"+opts.ID, jsonBody)
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

	row := buildCrewRateRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated crew rate %s\n", row.ID)
	return nil
}

func parseDoCrewRatesUpdateOptions(cmd *cobra.Command, args []string) (doCrewRatesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	pricePerUnit, _ := cmd.Flags().GetString("price-per-unit")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	isActive, _ := cmd.Flags().GetString("is-active")
	broker, _ := cmd.Flags().GetString("broker")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	resourceClassificationType, _ := cmd.Flags().GetString("resource-classification-type")
	resourceClassificationID, _ := cmd.Flags().GetString("resource-classification-id")
	craftClass, _ := cmd.Flags().GetString("craft-class")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCrewRatesUpdateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ID:                         args[0],
		Description:                description,
		PricePerUnit:               pricePerUnit,
		StartOn:                    startOn,
		EndOn:                      endOn,
		IsActive:                   isActive,
		Broker:                     broker,
		ResourceType:               resourceType,
		ResourceID:                 resourceID,
		ResourceClassificationType: resourceClassificationType,
		ResourceClassificationID:   resourceClassificationID,
		CraftClass:                 craftClass,
	}, nil
}
