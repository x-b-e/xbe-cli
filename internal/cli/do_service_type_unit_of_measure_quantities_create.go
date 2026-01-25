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

type doServiceTypeUnitOfMeasureQuantitiesCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	Quantity                 string
	ExplicitQuantity         string
	ServiceTypeUnitOfMeasure string
	QuantifiesType           string
	QuantifiesID             string
	MaterialType             string
	TrailerClassification    string
}

func newDoServiceTypeUnitOfMeasureQuantitiesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a service type unit of measure quantity",
		Long: `Create a service type unit of measure quantity.

Required flags:
  --service-type-unit-of-measure  Service type unit of measure ID (required)
  --quantifies-type               Quantified resource type (required, e.g., time-cards)
  --quantifies-id                 Quantified resource ID (required)

Optional flags:
  --quantity                      Quantity value
  --explicit-quantity             Explicit quantity value

Relationships:
  --material-type                 Material type ID
  --trailer-classification        Trailer classification ID`,
		Example: `  # Create a quantity for a time card
  xbe do service-type-unit-of-measure-quantities create \
    --service-type-unit-of-measure 123 \
    --quantifies-type time-cards \
    --quantifies-id 456 \
    --quantity 10

  # Create with explicit quantity and material type
  xbe do service-type-unit-of-measure-quantities create \
    --service-type-unit-of-measure 123 \
    --quantifies-type time-cards \
    --quantifies-id 456 \
    --explicit-quantity 12.5 \
    --material-type 789`,
		Args: cobra.NoArgs,
		RunE: runDoServiceTypeUnitOfMeasureQuantitiesCreate,
	}
	initDoServiceTypeUnitOfMeasureQuantitiesCreateFlags(cmd)
	return cmd
}

func init() {
	doServiceTypeUnitOfMeasureQuantitiesCmd.AddCommand(newDoServiceTypeUnitOfMeasureQuantitiesCreateCmd())
}

func initDoServiceTypeUnitOfMeasureQuantitiesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Quantity value")
	cmd.Flags().String("explicit-quantity", "", "Explicit quantity value")
	cmd.Flags().String("service-type-unit-of-measure", "", "Service type unit of measure ID (required)")
	cmd.Flags().String("quantifies-type", "", "Quantified resource type (required)")
	cmd.Flags().String("quantifies-id", "", "Quantified resource ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("service-type-unit-of-measure")
	_ = cmd.MarkFlagRequired("quantifies-type")
	_ = cmd.MarkFlagRequired("quantifies-id")
}

func runDoServiceTypeUnitOfMeasureQuantitiesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoServiceTypeUnitOfMeasureQuantitiesCreateOptions(cmd)
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

	if opts.ServiceTypeUnitOfMeasure == "" {
		err := fmt.Errorf("--service-type-unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.QuantifiesType == "" {
		err := fmt.Errorf("--quantifies-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.QuantifiesID == "" {
		err := fmt.Errorf("--quantifies-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Quantity != "" {
		attributes["quantity"] = opts.Quantity
	}
	if opts.ExplicitQuantity != "" {
		attributes["explicit-quantity"] = opts.ExplicitQuantity
	}

	relationships := map[string]any{
		"service-type-unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.ServiceTypeUnitOfMeasure,
			},
		},
		"quantifies": map[string]any{
			"data": map[string]any{
				"type": opts.QuantifiesType,
				"id":   opts.QuantifiesID,
			},
		},
	}

	if opts.MaterialType != "" {
		relationships["material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		}
	}
	if opts.TrailerClassification != "" {
		relationships["trailer-classification"] = map[string]any{
			"data": map[string]any{
				"type": "trailer-classifications",
				"id":   opts.TrailerClassification,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "service-type-unit-of-measure-quantities",
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

	body, _, err := client.Post(cmd.Context(), "/v1/service-type-unit-of-measure-quantities", jsonBody)
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

	row := serviceTypeUnitOfMeasureQuantityRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created service type unit of measure quantity %s\n", row.ID)
	return nil
}

func parseDoServiceTypeUnitOfMeasureQuantitiesCreateOptions(cmd *cobra.Command) (doServiceTypeUnitOfMeasureQuantitiesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantity, _ := cmd.Flags().GetString("quantity")
	explicitQuantity, _ := cmd.Flags().GetString("explicit-quantity")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	quantifiesType, _ := cmd.Flags().GetString("quantifies-type")
	quantifiesID, _ := cmd.Flags().GetString("quantifies-id")
	materialType, _ := cmd.Flags().GetString("material-type")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doServiceTypeUnitOfMeasureQuantitiesCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		Quantity:                 quantity,
		ExplicitQuantity:         explicitQuantity,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		QuantifiesType:           quantifiesType,
		QuantifiesID:             quantifiesID,
		MaterialType:             materialType,
		TrailerClassification:    trailerClassification,
	}, nil
}

func serviceTypeUnitOfMeasureQuantityRowFromSingle(resp jsonAPISingleResponse) serviceTypeUnitOfMeasureQuantityRow {
	attrs := resp.Data.Attributes
	row := serviceTypeUnitOfMeasureQuantityRow{
		ID:                 resp.Data.ID,
		Quantity:           stringAttr(attrs, "quantity"),
		ExplicitQuantity:   stringAttr(attrs, "explicit-quantity"),
		CalculatedQuantity: stringAttr(attrs, "calculated-quantity"),
	}
	if rel, ok := resp.Data.Relationships["service-type-unit-of-measure"]; ok && rel.Data != nil {
		row.ServiceTypeUnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["quantifies"]; ok && rel.Data != nil {
		row.QuantifiesType = rel.Data.Type
		row.QuantifiesID = rel.Data.ID
	}
	return row
}
