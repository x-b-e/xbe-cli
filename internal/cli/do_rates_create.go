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

type doRatesCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	RatedType                string
	RatedID                  string
	ServiceTypeUnitOfMeasure string
	PricePerUnit             string
	CurrencyCode             string
	Status                   string
	Name                     string
	Importance               string
	MaximumTravelMinutes     string
}

func newDoRatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a rate",
		Long: `Create a rate.

Required:
  --rated-type                   Rated type (e.g., broker-tenders, customer-tenders, rate-agreements)
  --rated-id                     Rated ID
  --service-type-unit-of-measure Service type unit of measure ID

Optional:
  --price-per-unit               Price per unit
  --currency-code                Currency code (e.g., USD)
  --status                       Status
  --name                         Rate name
  --importance                   Importance level
  --maximum-travel-minutes       Maximum travel minutes`,
		Example: `  # Create a rate for a tender
  xbe do rates create --rated-type broker-tenders --rated-id 123 --service-type-unit-of-measure 456

  # Create with price
  xbe do rates create --rated-type broker-tenders --rated-id 123 --service-type-unit-of-measure 456 \
    --price-per-unit "50.00" --currency-code USD

  # Create with name and status
  xbe do rates create --rated-type rate-agreements --rated-id 789 --service-type-unit-of-measure 456 \
    --name "Standard Hourly Rate" --status active`,
		RunE: runDoRatesCreate,
	}
	initDoRatesCreateFlags(cmd)
	return cmd
}

func init() {
	doRatesCmd.AddCommand(newDoRatesCreateCmd())
}

func initDoRatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rated-type", "", "Rated type (e.g., broker-tenders, customer-tenders, rate-agreements)")
	cmd.Flags().String("rated-id", "", "Rated ID")
	cmd.Flags().String("service-type-unit-of-measure", "", "Service type unit of measure ID")
	cmd.Flags().String("price-per-unit", "", "Price per unit")
	cmd.Flags().String("currency-code", "", "Currency code (e.g., USD)")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().String("name", "", "Rate name")
	cmd.Flags().String("importance", "", "Importance level")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("rated-type")
	_ = cmd.MarkFlagRequired("rated-id")
	_ = cmd.MarkFlagRequired("service-type-unit-of-measure")
}

func runDoRatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRatesCreateOptions(cmd)
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

	if opts.PricePerUnit != "" {
		attributes["price-per-unit"] = opts.PricePerUnit
	}
	if opts.CurrencyCode != "" {
		attributes["currency-code"] = opts.CurrencyCode
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Importance != "" {
		attributes["importance"] = opts.Importance
	}
	if opts.MaximumTravelMinutes != "" {
		attributes["maximum-travel-minutes"] = opts.MaximumTravelMinutes
	}

	relationships := map[string]any{
		"rated": map[string]any{
			"data": map[string]any{
				"type": opts.RatedType,
				"id":   opts.RatedID,
			},
		},
		"service-type-unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.ServiceTypeUnitOfMeasure,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "rates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/rates", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created rate %s\n", resp.Data.ID)
	return nil
}

func parseDoRatesCreateOptions(cmd *cobra.Command) (doRatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	ratedType, _ := cmd.Flags().GetString("rated-type")
	ratedID, _ := cmd.Flags().GetString("rated-id")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	pricePerUnit, _ := cmd.Flags().GetString("price-per-unit")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	status, _ := cmd.Flags().GetString("status")
	name, _ := cmd.Flags().GetString("name")
	importance, _ := cmd.Flags().GetString("importance")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRatesCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		RatedType:                ratedType,
		RatedID:                  ratedID,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		PricePerUnit:             pricePerUnit,
		CurrencyCode:             currencyCode,
		Status:                   status,
		Name:                     name,
		Importance:               importance,
		MaximumTravelMinutes:     maximumTravelMinutes,
	}, nil
}
