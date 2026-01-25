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

type doTransportOrderStopMaterialsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	QuantityExplicit       string
	TransportOrderMaterial string
	TransportOrderStop     string
}

func newDoTransportOrderStopMaterialsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a transport order stop material",
		Long: `Create a transport order stop material.

Required flags:
  --transport-order-material  Transport order material ID
  --transport-order-stop      Transport order stop ID

Optional flags:
  --quantity-explicit         Explicit quantity for the stop`,
		Example: `  # Create a transport order stop material
  xbe do transport-order-stop-materials create \
    --transport-order-material 123 \
    --transport-order-stop 456 \
    --quantity-explicit 10.5`,
		RunE: runDoTransportOrderStopMaterialsCreate,
	}
	initDoTransportOrderStopMaterialsCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderStopMaterialsCmd.AddCommand(newDoTransportOrderStopMaterialsCreateCmd())
}

func initDoTransportOrderStopMaterialsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity-explicit", "", "Explicit quantity for the stop")
	cmd.Flags().String("transport-order-material", "", "Transport order material ID (required)")
	cmd.Flags().String("transport-order-stop", "", "Transport order stop ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("transport-order-material")
	cmd.MarkFlagRequired("transport-order-stop")
}

func runDoTransportOrderStopMaterialsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTransportOrderStopMaterialsCreateOptions(cmd)
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
		"transport-order-material": map[string]any{
			"data": map[string]any{
				"type": "transport-order-materials",
				"id":   opts.TransportOrderMaterial,
			},
		},
		"transport-order-stop": map[string]any{
			"data": map[string]any{
				"type": "transport-order-stops",
				"id":   opts.TransportOrderStop,
			},
		},
	}

	data := map[string]any{
		"type":          "transport-order-stop-materials",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/transport-order-stop-materials", jsonBody)
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
			"id":                resp.Data.ID,
			"quantity_explicit": stringAttr(resp.Data.Attributes, "quantity-explicit"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created transport order stop material %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportOrderStopMaterialsCreateOptions(cmd *cobra.Command) (doTransportOrderStopMaterialsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantityExplicit, _ := cmd.Flags().GetString("quantity-explicit")
	transportOrderMaterial, _ := cmd.Flags().GetString("transport-order-material")
	transportOrderStop, _ := cmd.Flags().GetString("transport-order-stop")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderStopMaterialsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		QuantityExplicit:       quantityExplicit,
		TransportOrderMaterial: transportOrderMaterial,
		TransportOrderStop:     transportOrderStop,
	}, nil
}
