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

type doMaterialPurchaseOrderReleaseRedemptionsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	TicketNumber        string
	MaterialTransaction string
}

func newDoMaterialPurchaseOrderReleaseRedemptionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material purchase order release redemption",
		Long: `Update a material purchase order release redemption.

All flags are optional. Only provided flags will update the redemption.

Optional flags:
  --ticket-number          Ticket number

Relationships:
  --material-transaction   Material transaction ID (set empty to clear)`,
		Example: `  # Update the ticket number
  xbe do material-purchase-order-release-redemptions update 123 --ticket-number T-200

  # Attach a material transaction
  xbe do material-purchase-order-release-redemptions update 123 --material-transaction 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialPurchaseOrderReleaseRedemptionsUpdate,
	}
	initDoMaterialPurchaseOrderReleaseRedemptionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialPurchaseOrderReleaseRedemptionsCmd.AddCommand(newDoMaterialPurchaseOrderReleaseRedemptionsUpdateCmd())
}

func initDoMaterialPurchaseOrderReleaseRedemptionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("ticket-number", "", "Ticket number")
	cmd.Flags().String("material-transaction", "", "Material transaction ID (empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialPurchaseOrderReleaseRedemptionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialPurchaseOrderReleaseRedemptionsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("ticket-number") {
		attributes["ticket-number"] = opts.TicketNumber
	}

	if cmd.Flags().Changed("material-transaction") {
		if opts.MaterialTransaction == "" {
			relationships["material-transaction"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["material-transaction"] = map[string]any{
				"data": map[string]any{
					"type": "material-transactions",
					"id":   opts.MaterialTransaction,
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
		"type": "material-purchase-order-release-redemptions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-purchase-order-release-redemptions/"+opts.ID, jsonBody)
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

	row := materialPurchaseOrderReleaseRedemptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material purchase order release redemption %s\n", row.ID)
	return nil
}

func parseDoMaterialPurchaseOrderReleaseRedemptionsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialPurchaseOrderReleaseRedemptionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	ticketNumber, _ := cmd.Flags().GetString("ticket-number")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialPurchaseOrderReleaseRedemptionsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		TicketNumber:        ticketNumber,
		MaterialTransaction: materialTransaction,
	}, nil
}
