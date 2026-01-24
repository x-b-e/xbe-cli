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

type doMissingRatesCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	JobID                    string
	ServiceTypeUnitOfMeasure string
	CurrencyCode             string
	CustomerPricePerUnit     string
	TruckerPricePerUnit      string
}

func newDoMissingRatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a missing rate",
		Long: `Create a missing rate.

Missing rates add service type unit of measure pricing to a job and create
rates for related customer and broker tenders.

Required:
  --job                         Job ID
  --service-type-unit-of-measure Service type unit of measure ID
  --currency-code               Currency code (e.g., USD)
  --customer-price-per-unit     Customer price per unit
  --trucker-price-per-unit      Trucker price per unit`,
		Example: `  # Create a missing rate
  xbe do missing-rates create \
    --job 123 \
    --service-type-unit-of-measure 456 \
    --currency-code USD \
    --customer-price-per-unit 100.00 \
    --trucker-price-per-unit 85.00`,
		RunE: runDoMissingRatesCreate,
	}
	initDoMissingRatesCreateFlags(cmd)
	return cmd
}

func init() {
	doMissingRatesCmd.AddCommand(newDoMissingRatesCreateCmd())
}

func initDoMissingRatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job", "", "Job ID")
	cmd.Flags().String("service-type-unit-of-measure", "", "Service type unit of measure ID")
	cmd.Flags().String("currency-code", "", "Currency code (e.g., USD)")
	cmd.Flags().String("customer-price-per-unit", "", "Customer price per unit")
	cmd.Flags().String("trucker-price-per-unit", "", "Trucker price per unit")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job")
	_ = cmd.MarkFlagRequired("service-type-unit-of-measure")
	_ = cmd.MarkFlagRequired("currency-code")
	_ = cmd.MarkFlagRequired("customer-price-per-unit")
	_ = cmd.MarkFlagRequired("trucker-price-per-unit")
}

func runDoMissingRatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMissingRatesCreateOptions(cmd)
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

	attributes := map[string]any{
		"currency-code":           opts.CurrencyCode,
		"customer-price-per-unit": opts.CustomerPricePerUnit,
		"trucker-price-per-unit":  opts.TruckerPricePerUnit,
	}

	relationships := map[string]any{
		"job": map[string]any{
			"data": map[string]any{
				"type": "jobs",
				"id":   opts.JobID,
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
			"type":          "missing-rates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/missing-rates", jsonBody)
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

	row := missingRateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created missing rate %s\n", row.ID)
	return nil
}

func parseDoMissingRatesCreateOptions(cmd *cobra.Command) (doMissingRatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobID, _ := cmd.Flags().GetString("job")
	serviceTypeUnitOfMeasure, _ := cmd.Flags().GetString("service-type-unit-of-measure")
	currencyCode, _ := cmd.Flags().GetString("currency-code")
	customerPricePerUnit, _ := cmd.Flags().GetString("customer-price-per-unit")
	truckerPricePerUnit, _ := cmd.Flags().GetString("trucker-price-per-unit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMissingRatesCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		JobID:                    jobID,
		ServiceTypeUnitOfMeasure: serviceTypeUnitOfMeasure,
		CurrencyCode:             currencyCode,
		CustomerPricePerUnit:     customerPricePerUnit,
		TruckerPricePerUnit:      truckerPricePerUnit,
	}, nil
}
