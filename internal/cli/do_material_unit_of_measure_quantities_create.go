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

type doMaterialUnitOfMeasureQuantitiesCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	MaterialTransaction string
	UnitOfMeasure       string
	Quantity            string
}

func newDoMaterialUnitOfMeasureQuantitiesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material unit of measure quantity",
		Long: `Create a material unit of measure quantity.

Required flags:
  --material-transaction  Material transaction ID
  --unit-of-measure       Unit of measure ID
  --quantity              Quantity (0 to 200000)

Notes:
  The material transaction must be editable.`,
		Example: `  # Create a quantity record
  xbe do material-unit-of-measure-quantities create \
    --material-transaction 123 \
    --unit-of-measure 45 \
    --quantity 18.5`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialUnitOfMeasureQuantitiesCreate,
	}
	initDoMaterialUnitOfMeasureQuantitiesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialUnitOfMeasureQuantitiesCmd.AddCommand(newDoMaterialUnitOfMeasureQuantitiesCreateCmd())
}

func initDoMaterialUnitOfMeasureQuantitiesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("quantity", "", "Quantity (0 to 200000)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-transaction")
	_ = cmd.MarkFlagRequired("unit-of-measure")
	_ = cmd.MarkFlagRequired("quantity")
}

func runDoMaterialUnitOfMeasureQuantitiesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialUnitOfMeasureQuantitiesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.MaterialTransaction) == "" {
		err := fmt.Errorf("--material-transaction is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.UnitOfMeasure) == "" {
		err := fmt.Errorf("--unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Quantity) == "" {
		err := fmt.Errorf("--quantity is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"quantity": opts.Quantity,
	}
	relationships := map[string]any{
		"material-transaction": map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   opts.MaterialTransaction,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-unit-of-measure-quantities",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-unit-of-measure-quantities", jsonBody)
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

	row := buildMaterialUnitOfMeasureQuantityRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material unit of measure quantity %s\n", row.ID)
	return nil
}

func parseDoMaterialUnitOfMeasureQuantitiesCreateOptions(cmd *cobra.Command) (doMaterialUnitOfMeasureQuantitiesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantity, _ := cmd.Flags().GetString("quantity")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialUnitOfMeasureQuantitiesCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		MaterialTransaction: materialTransaction,
		UnitOfMeasure:       unitOfMeasure,
		Quantity:            quantity,
	}, nil
}
