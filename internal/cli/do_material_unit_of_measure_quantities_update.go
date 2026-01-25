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

type doMaterialUnitOfMeasureQuantitiesUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	MaterialTransaction string
	UnitOfMeasure       string
	Quantity            string
}

func newDoMaterialUnitOfMeasureQuantitiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material unit of measure quantity",
		Long: `Update a material unit of measure quantity.

Optional flags:
  --quantity              Quantity (0 to 200000)
  --material-transaction  Material transaction ID
  --unit-of-measure       Unit of measure ID`,
		Example: `  # Update quantity
  xbe do material-unit-of-measure-quantities update 123 --quantity 20.5

  # Update unit of measure
  xbe do material-unit-of-measure-quantities update 123 --unit-of-measure 45`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialUnitOfMeasureQuantitiesUpdate,
	}
	initDoMaterialUnitOfMeasureQuantitiesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialUnitOfMeasureQuantitiesCmd.AddCommand(newDoMaterialUnitOfMeasureQuantitiesUpdateCmd())
}

func initDoMaterialUnitOfMeasureQuantitiesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Quantity (0 to 200000)")
	cmd.Flags().String("material-transaction", "", "Material transaction ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialUnitOfMeasureQuantitiesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialUnitOfMeasureQuantitiesUpdateOptions(cmd, args)
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
	hasChanges := false

	if cmd.Flags().Changed("quantity") {
		if strings.TrimSpace(opts.Quantity) == "" {
			err := fmt.Errorf("--quantity cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["quantity"] = opts.Quantity
		hasChanges = true
	}
	if cmd.Flags().Changed("material-transaction") {
		if strings.TrimSpace(opts.MaterialTransaction) == "" {
			err := fmt.Errorf("--material-transaction cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["material-transaction"] = map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   opts.MaterialTransaction,
			},
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("unit-of-measure") {
		if strings.TrimSpace(opts.UnitOfMeasure) == "" {
			err := fmt.Errorf("--unit-of-measure cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "material-unit-of-measure-quantities",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-unit-of-measure-quantities/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material unit of measure quantity %s\n", row.ID)
	return nil
}

func parseDoMaterialUnitOfMeasureQuantitiesUpdateOptions(cmd *cobra.Command, args []string) (doMaterialUnitOfMeasureQuantitiesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantity, _ := cmd.Flags().GetString("quantity")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialUnitOfMeasureQuantitiesUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		MaterialTransaction: materialTransaction,
		UnitOfMeasure:       unitOfMeasure,
		Quantity:            quantity,
	}, nil
}
