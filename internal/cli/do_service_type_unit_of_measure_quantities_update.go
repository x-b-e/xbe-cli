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

type doServiceTypeUnitOfMeasureQuantitiesUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ID                       string
	Quantity                 string
	ExplicitQuantity         string
	ServiceTypeUnitOfMeasure string
	QuantifiesType           string
	QuantifiesID             string
	MaterialType             string
	TrailerClassification    string
}

func newDoServiceTypeUnitOfMeasureQuantitiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a service type unit of measure quantity",
		Long: `Update a service type unit of measure quantity.

All flags are optional. Only provided flags will be updated.

Optional flags:
  --quantity                      Quantity value
  --explicit-quantity             Explicit quantity value
  --service-type-unit-of-measure  Service type unit of measure ID
  --quantifies-type               Quantified resource type (requires --quantifies-id)
  --quantifies-id                 Quantified resource ID (requires --quantifies-type)
  --material-type                 Material type ID (set empty to clear)
  --trailer-classification        Trailer classification ID (set empty to clear)`,
		Example: `  # Update quantity
  xbe do service-type-unit-of-measure-quantities update 123 --quantity 8

  # Update explicit quantity
  xbe do service-type-unit-of-measure-quantities update 123 --explicit-quantity 10.5

  # Set material type
  xbe do service-type-unit-of-measure-quantities update 123 --material-type 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoServiceTypeUnitOfMeasureQuantitiesUpdate,
	}
	initDoServiceTypeUnitOfMeasureQuantitiesUpdateFlags(cmd)
	return cmd
}

func init() {
	doServiceTypeUnitOfMeasureQuantitiesCmd.AddCommand(newDoServiceTypeUnitOfMeasureQuantitiesUpdateCmd())
}

func initDoServiceTypeUnitOfMeasureQuantitiesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Quantity value")
	cmd.Flags().String("explicit-quantity", "", "Explicit quantity value")
	cmd.Flags().String("service-type-unit-of-measure", "", "Service type unit of measure ID")
	cmd.Flags().String("quantifies-type", "", "Quantified resource type")
	cmd.Flags().String("quantifies-id", "", "Quantified resource ID")
	cmd.Flags().String("material-type", "", "Material type ID (empty to clear)")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID (empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoServiceTypeUnitOfMeasureQuantitiesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoServiceTypeUnitOfMeasureQuantitiesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("explicit-quantity") {
		attributes["explicit-quantity"] = opts.ExplicitQuantity
	}

	if cmd.Flags().Changed("service-type-unit-of-measure") {
		if opts.ServiceTypeUnitOfMeasure == "" {
			return fmt.Errorf("--service-type-unit-of-measure cannot be empty")
		}
		relationships["service-type-unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.ServiceTypeUnitOfMeasure,
			},
		}
	}

	if cmd.Flags().Changed("quantifies-type") || cmd.Flags().Changed("quantifies-id") {
		if opts.QuantifiesType == "" || opts.QuantifiesID == "" {
			return fmt.Errorf("--quantifies-type and --quantifies-id are required together")
		}
		relationships["quantifies"] = map[string]any{
			"data": map[string]any{
				"type": opts.QuantifiesType,
				"id":   opts.QuantifiesID,
			},
		}
	}

	if cmd.Flags().Changed("material-type") {
		if opts.MaterialType == "" {
			relationships["material-type"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["material-type"] = map[string]any{
				"data": map[string]any{
					"type": "material-types",
					"id":   opts.MaterialType,
				},
			}
		}
	}

	if cmd.Flags().Changed("trailer-classification") {
		if opts.TrailerClassification == "" {
			relationships["trailer-classification"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["trailer-classification"] = map[string]any{
				"data": map[string]any{
					"type": "trailer-classifications",
					"id":   opts.TrailerClassification,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "service-type-unit-of-measure-quantities",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/service-type-unit-of-measure-quantities/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated service type unit of measure quantity %s\n", row.ID)
	return nil
}

func parseDoServiceTypeUnitOfMeasureQuantitiesUpdateOptions(cmd *cobra.Command, args []string) (doServiceTypeUnitOfMeasureQuantitiesUpdateOptions, error) {
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

	return doServiceTypeUnitOfMeasureQuantitiesUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		Quantity:                 quantity,
		ExplicitQuantity:         explicitQuantity,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		QuantifiesType:           quantifiesType,
		QuantifiesID:             quantifiesID,
		MaterialType:             materialType,
		TrailerClassification:    trailerClassification,
	}, nil
}
