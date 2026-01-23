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

type doTransportOrderMaterialsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	TransportOrder   string
	MaterialType     string
	UnitOfMeasure    string
	QuantityExplicit string
}

func newDoTransportOrderMaterialsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a transport order material",
		Long: `Create a transport order material.

Required flags:
  --transport-order  Transport order ID
  --material-type    Material type ID
  --unit-of-measure  Unit of measure ID

Optional flags:
  --quantity-explicit  Explicit quantity`,
		Example: `  # Create a transport order material
  xbe do transport-order-materials create --transport-order 123 --material-type 456 --unit-of-measure 789 --quantity-explicit 10`,
		RunE: runDoTransportOrderMaterialsCreate,
	}
	initDoTransportOrderMaterialsCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderMaterialsCmd.AddCommand(newDoTransportOrderMaterialsCreateCmd())
}

func initDoTransportOrderMaterialsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("transport-order", "", "Transport order ID (required)")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (required)")
	cmd.Flags().String("quantity-explicit", "", "Explicit quantity")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("transport-order")
	_ = cmd.MarkFlagRequired("material-type")
	_ = cmd.MarkFlagRequired("unit-of-measure")
}

func runDoTransportOrderMaterialsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTransportOrderMaterialsCreateOptions(cmd)
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
	if opts.QuantityExplicit != "" {
		attributes["quantity-explicit"] = opts.QuantityExplicit
	}

	relationships := map[string]any{
		"transport-order": map[string]any{
			"data": map[string]any{
				"type": "transport-orders",
				"id":   opts.TransportOrder,
			},
		},
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		},
	}

	data := map[string]any{
		"type": "transport-order-materials",
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

	body, _, err := client.Post(cmd.Context(), "/v1/transport-order-materials", jsonBody)
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

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id": resp.Data.ID,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created transport order material %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportOrderMaterialsCreateOptions(cmd *cobra.Command) (doTransportOrderMaterialsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	materialType, _ := cmd.Flags().GetString("material-type")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantityExplicit, _ := cmd.Flags().GetString("quantity-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderMaterialsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		TransportOrder:   transportOrder,
		MaterialType:     materialType,
		UnitOfMeasure:    unitOfMeasure,
		QuantityExplicit: quantityExplicit,
	}, nil
}
