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

type doCrewRatesCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
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

func newDoCrewRatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a crew rate",
		Long: `Create a crew rate.

Required:
  --price-per-unit               Price per unit
  --start-on                     Start date (YYYY-MM-DD)
  --is-active                    Active status (true/false)
  --broker                       Broker ID
  --resource-type/--resource-id OR --resource-classification-type/--resource-classification-id OR --craft-class

Optional:
  --description                  Description
  --end-on                       End date (YYYY-MM-DD)
  --resource-type                Resource type (Laborer, Equipment)
  --resource-id                  Resource ID
  --resource-classification-type Resource classification type (LaborClassification, EquipmentClassification)
  --resource-classification-id   Resource classification ID
  --craft-class                  Craft class ID`,
		Example: `  # Create with a resource classification
  xbe do crew-rates create --price-per-unit 75.00 --start-on 2025-01-01 --is-active true \
    --broker 123 --resource-classification-type LaborClassification --resource-classification-id 456

  # Create for a specific resource
  xbe do crew-rates create --price-per-unit 90.00 --start-on 2025-01-01 --is-active true \
    --broker 123 --resource-type Equipment --resource-id 789

  # Create for a craft class
  xbe do crew-rates create --price-per-unit 60.00 --start-on 2025-01-01 --is-active true \
    --broker 123 --craft-class 321`,
		Args: cobra.NoArgs,
		RunE: runDoCrewRatesCreate,
	}
	initDoCrewRatesCreateFlags(cmd)
	return cmd
}

func init() {
	doCrewRatesCmd.AddCommand(newDoCrewRatesCreateCmd())
}

func initDoCrewRatesCreateFlags(cmd *cobra.Command) {
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

func runDoCrewRatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCrewRatesCreateOptions(cmd)
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

	if opts.PricePerUnit == "" {
		err := fmt.Errorf("--price-per-unit is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.StartOn == "" {
		err := fmt.Errorf("--start-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.IsActive == "" {
		err := fmt.Errorf("--is-active is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	isActive, err := parseCrewRateBool(opts.IsActive, "is-active")
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	resourceProvided := opts.ResourceType != "" || opts.ResourceID != ""
	resourceClassProvided := opts.ResourceClassificationType != "" || opts.ResourceClassificationID != ""
	craftClassProvided := opts.CraftClass != ""

	if !resourceProvided && !resourceClassProvided && !craftClassProvided {
		err := fmt.Errorf("at least one of --resource-type/--resource-id, --resource-classification-type/--resource-classification-id, or --craft-class is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if resourceProvided && (opts.ResourceType == "" || opts.ResourceID == "") {
		err := fmt.Errorf("--resource-type and --resource-id must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if resourceClassProvided && (opts.ResourceClassificationType == "" || opts.ResourceClassificationID == "") {
		err := fmt.Errorf("--resource-classification-type and --resource-classification-id must be provided together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"price-per-unit": opts.PricePerUnit,
		"start-on":       opts.StartOn,
		"is-active":      isActive,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.EndOn != "" {
		attributes["end-on"] = opts.EndOn
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	if resourceProvided {
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
	if resourceClassProvided {
		resourceClassificationType, err := parseCrewRateResourceClassificationType(opts.ResourceClassificationType)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["resource-classification"] = map[string]any{
			"data": map[string]any{
				"type": resourceClassificationType,
				"id":   opts.ResourceClassificationID,
			},
		}
	}
	if craftClassProvided {
		relationships["craft-class"] = map[string]any{
			"data": map[string]any{
				"type": "craft-classes",
				"id":   opts.CraftClass,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "crew-rates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/crew-rates", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created crew rate %s\n", row.ID)
	return nil
}

func parseDoCrewRatesCreateOptions(cmd *cobra.Command) (doCrewRatesCreateOptions, error) {
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

	return doCrewRatesCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
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
