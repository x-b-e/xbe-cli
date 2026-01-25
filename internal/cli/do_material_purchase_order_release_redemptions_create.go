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

type doMaterialPurchaseOrderReleaseRedemptionsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Release             string
	TicketNumber        string
	MaterialTransaction string
}

func newDoMaterialPurchaseOrderReleaseRedemptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material purchase order release redemption",
		Long: `Create a material purchase order release redemption.

Required flags:
  --release                Release ID (required)
  --ticket-number          Ticket number (required unless --material-transaction is provided)

Relationships:
  --material-transaction   Material transaction ID (optional; required if --ticket-number omitted)`,
		Example: `  # Create a redemption with a ticket number
  xbe do material-purchase-order-release-redemptions create --release 123 --ticket-number T-100

  # Create a redemption with a material transaction
  xbe do material-purchase-order-release-redemptions create --release 123 --material-transaction 456`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialPurchaseOrderReleaseRedemptionsCreate,
	}
	initDoMaterialPurchaseOrderReleaseRedemptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialPurchaseOrderReleaseRedemptionsCmd.AddCommand(newDoMaterialPurchaseOrderReleaseRedemptionsCreateCmd())
}

func initDoMaterialPurchaseOrderReleaseRedemptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("release", "", "Release ID (required)")
	cmd.Flags().String("ticket-number", "", "Ticket number")
	cmd.Flags().String("material-transaction", "", "Material transaction ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialPurchaseOrderReleaseRedemptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialPurchaseOrderReleaseRedemptionsCreateOptions(cmd)
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

	if opts.Release == "" {
		err := fmt.Errorf("--release is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TicketNumber == "" && opts.MaterialTransaction == "" {
		err := fmt.Errorf("either --ticket-number or --material-transaction is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.TicketNumber != "" {
		attributes["ticket-number"] = opts.TicketNumber
	}

	relationships := map[string]any{
		"release": map[string]any{
			"data": map[string]any{
				"type": "material-purchase-order-releases",
				"id":   opts.Release,
			},
		},
	}

	if opts.MaterialTransaction != "" {
		relationships["material-transaction"] = map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   opts.MaterialTransaction,
			},
		}
	}

	data := map[string]any{
		"type":          "material-purchase-order-release-redemptions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-purchase-order-release-redemptions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material purchase order release redemption %s\n", row.ID)
	return nil
}

func parseDoMaterialPurchaseOrderReleaseRedemptionsCreateOptions(cmd *cobra.Command) (doMaterialPurchaseOrderReleaseRedemptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	release, _ := cmd.Flags().GetString("release")
	ticketNumber, _ := cmd.Flags().GetString("ticket-number")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialPurchaseOrderReleaseRedemptionsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Release:             release,
		TicketNumber:        ticketNumber,
		MaterialTransaction: materialTransaction,
	}, nil
}
