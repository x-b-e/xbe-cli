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

type doTransportOrderStopMaterialsUpdateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ID               string
	QuantityExplicit string
}

func newDoTransportOrderStopMaterialsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a transport order stop material",
		Long: `Update a transport order stop material.

Optional flags:
  --quantity-explicit  Explicit quantity for the stop`,
		Example: `  # Update quantity for a transport order stop material
  xbe do transport-order-stop-materials update 123 --quantity-explicit 12.5`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTransportOrderStopMaterialsUpdate,
	}
	initDoTransportOrderStopMaterialsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderStopMaterialsCmd.AddCommand(newDoTransportOrderStopMaterialsUpdateCmd())
}

func initDoTransportOrderStopMaterialsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity-explicit", "", "Explicit quantity for the stop")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportOrderStopMaterialsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportOrderStopMaterialsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("quantity-explicit") {
		attributes["quantity-explicit"] = opts.QuantityExplicit
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --quantity-explicit")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "transport-order-stop-materials",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/transport-order-stop-materials/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated transport order stop material %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportOrderStopMaterialsUpdateOptions(cmd *cobra.Command, args []string) (doTransportOrderStopMaterialsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantityExplicit, _ := cmd.Flags().GetString("quantity-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderStopMaterialsUpdateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ID:               args[0],
		QuantityExplicit: quantityExplicit,
	}, nil
}
