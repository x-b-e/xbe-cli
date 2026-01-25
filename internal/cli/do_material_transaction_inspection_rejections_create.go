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

type doMaterialTransactionInspectionRejectionsCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	MaterialTransactionInspection string
	UnitOfMeasure                 string
	Quantity                      string
	Note                          string
}

func newDoMaterialTransactionInspectionRejectionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction inspection rejection",
		Long: `Create a material transaction inspection rejection.

Required flags:
  --material-transaction-inspection  Material transaction inspection ID
  --unit-of-measure                  Unit of measure ID
  --quantity                         Rejected quantity (must be > 0)

Optional flags:
  --note  Rejection note

Notes:
  The material transaction inspection must be open.`,
		Example: `  # Create a rejection
  xbe do material-transaction-inspection-rejections create \\
    --material-transaction-inspection 123 \\
    --unit-of-measure 45 \\
    --quantity 10 \\
    --note "Excess moisture"`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionInspectionRejectionsCreate,
	}
	initDoMaterialTransactionInspectionRejectionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionInspectionRejectionsCmd.AddCommand(newDoMaterialTransactionInspectionRejectionsCreateCmd())
}

func initDoMaterialTransactionInspectionRejectionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction-inspection", "", "Material transaction inspection ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("quantity", "", "Rejected quantity (must be > 0)")
	cmd.Flags().String("note", "", "Rejection note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-transaction-inspection")
	_ = cmd.MarkFlagRequired("unit-of-measure")
	_ = cmd.MarkFlagRequired("quantity")
}

func runDoMaterialTransactionInspectionRejectionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionInspectionRejectionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.MaterialTransactionInspection) == "" {
		err := fmt.Errorf("--material-transaction-inspection is required")
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
	if strings.TrimSpace(opts.Note) != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"material-transaction-inspection": map[string]any{
			"data": map[string]any{
				"type": "material-transaction-inspections",
				"id":   opts.MaterialTransactionInspection,
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
			"type":          "material-transaction-inspection-rejections",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-inspection-rejections", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction inspection rejection %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionInspectionRejectionsCreateOptions(cmd *cobra.Command) (doMaterialTransactionInspectionRejectionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransactionInspection, _ := cmd.Flags().GetString("material-transaction-inspection")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantity, _ := cmd.Flags().GetString("quantity")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionInspectionRejectionsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		MaterialTransactionInspection: materialTransactionInspection,
		UnitOfMeasure:                 unitOfMeasure,
		Quantity:                      quantity,
		Note:                          note,
	}, nil
}
