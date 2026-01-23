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

type doMaterialTransactionInspectionRejectionsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	UnitOfMeasure string
	Quantity      string
	Note          string
}

func newDoMaterialTransactionInspectionRejectionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material transaction inspection rejection",
		Long: `Update a material transaction inspection rejection.

Optional flags:
  --quantity        Rejected quantity (must be > 0)
  --unit-of-measure Unit of measure ID
  --note            Rejection note`,
		Example: `  # Update quantity
  xbe do material-transaction-inspection-rejections update 123 --quantity 12

  # Update note
  xbe do material-transaction-inspection-rejections update 123 --note "Updated note"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTransactionInspectionRejectionsUpdate,
	}
	initDoMaterialTransactionInspectionRejectionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionInspectionRejectionsCmd.AddCommand(newDoMaterialTransactionInspectionRejectionsUpdateCmd())
}

func initDoMaterialTransactionInspectionRejectionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Rejected quantity (must be > 0)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("note", "", "Rejection note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionInspectionRejectionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTransactionInspectionRejectionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
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
		"type": "material-transaction-inspection-rejections",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-transaction-inspection-rejections/"+opts.ID, jsonBody)
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

	row := buildMaterialTransactionInspectionRejectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material transaction inspection rejection %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionInspectionRejectionsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTransactionInspectionRejectionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantity, _ := cmd.Flags().GetString("quantity")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionInspectionRejectionsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		UnitOfMeasure: unitOfMeasure,
		Quantity:      quantity,
		Note:          note,
	}, nil
}
