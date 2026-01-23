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

type doRatesUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	PricePerUnit         string
	CurrencyCode         string
	Status               string
	Name                 string
	Importance           string
	MaximumTravelMinutes string
}

func newDoRatesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a rate",
		Long: `Update a rate.

Optional:
  --price-per-unit           Price per unit
  --currency-code            Currency code (e.g., USD)
  --status                   Status
  --name                     Rate name
  --importance               Importance level
  --maximum-travel-minutes   Maximum travel minutes`,
		Example: `  # Update price
  xbe do rates update 123 --price-per-unit "55.00"

  # Update status
  xbe do rates update 123 --status inactive

  # Update name and importance
  xbe do rates update 123 --name "Premium Hourly Rate" --importance 1`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRatesUpdate,
	}
	initDoRatesUpdateFlags(cmd)
	return cmd
}

func init() {
	doRatesCmd.AddCommand(newDoRatesUpdateCmd())
}

func initDoRatesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("price-per-unit", "", "Price per unit")
	cmd.Flags().String("currency-code", "", "Currency code (e.g., USD)")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().String("name", "", "Rate name")
	cmd.Flags().String("importance", "", "Importance level")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRatesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRatesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("price-per-unit") {
		attributes["price-per-unit"] = opts.PricePerUnit
	}
	if cmd.Flags().Changed("currency-code") {
		attributes["currency-code"] = opts.CurrencyCode
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("importance") {
		attributes["importance"] = opts.Importance
	}
	if cmd.Flags().Changed("maximum-travel-minutes") {
		attributes["maximum-travel-minutes"] = opts.MaximumTravelMinutes
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "rates",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/rates/"+opts.ID, jsonBody)
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
		row := rateRow{
			ID:           resp.Data.ID,
			Name:         stringAttr(resp.Data.Attributes, "name"),
			PricePerUnit: stringAttr(resp.Data.Attributes, "price-per-unit"),
			CurrencyCode: stringAttr(resp.Data.Attributes, "currency-code"),
			Status:       stringAttr(resp.Data.Attributes, "status"),
		}
		if rel, ok := resp.Data.Relationships["rated"]; ok && rel.Data != nil {
			row.RatedType = rel.Data.Type
			row.RatedID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated rate %s\n", resp.Data.ID)
	return nil
}

func parseDoRatesUpdateOptions(cmd *cobra.Command, args []string) (doRatesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	pricePerUnit, _ := cmd.Flags().GetString("price-per-unit")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	status, _ := cmd.Flags().GetString("status")
	name, _ := cmd.Flags().GetString("name")
	importance, _ := cmd.Flags().GetString("importance")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRatesUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		PricePerUnit:         pricePerUnit,
		CurrencyCode:         currencyCode,
		Status:               status,
		Name:                 name,
		Importance:           importance,
		MaximumTravelMinutes: maximumTravelMinutes,
	}, nil
}
