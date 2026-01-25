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

type doTransportOrderMaterialsUpdateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ID               string
	MaterialType     string
	UnitOfMeasure    string
	QuantityExplicit string
}

func newDoTransportOrderMaterialsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a transport order material",
		Long: `Update a transport order material.

All flags are optional. Only provided flags will update the material.

Optional flags:
  --quantity-explicit  Explicit quantity

Relationships:
  --material-type      Material type ID
  --unit-of-measure    Unit of measure ID`,
		Example: `  # Update explicit quantity
  xbe do transport-order-materials update 123 --quantity-explicit 12

  # Update material type
  xbe do transport-order-materials update 123 --material-type 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTransportOrderMaterialsUpdate,
	}
	initDoTransportOrderMaterialsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderMaterialsCmd.AddCommand(newDoTransportOrderMaterialsUpdateCmd())
}

func initDoTransportOrderMaterialsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity-explicit", "", "Explicit quantity")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportOrderMaterialsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportOrderMaterialsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("quantity-explicit") {
		attributes["quantity-explicit"] = opts.QuantityExplicit
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

	if cmd.Flags().Changed("unit-of-measure") {
		if opts.UnitOfMeasure == "" {
			relationships["unit-of-measure"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["unit-of-measure"] = map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.UnitOfMeasure,
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
		"type": "transport-order-materials",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/transport-order-materials/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated transport order material %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportOrderMaterialsUpdateOptions(cmd *cobra.Command, args []string) (doTransportOrderMaterialsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantityExplicit, _ := cmd.Flags().GetString("quantity-explicit")
	materialType, _ := cmd.Flags().GetString("material-type")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderMaterialsUpdateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ID:               args[0],
		MaterialType:     materialType,
		UnitOfMeasure:    unitOfMeasure,
		QuantityExplicit: quantityExplicit,
	}, nil
}
