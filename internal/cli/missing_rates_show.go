package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type missingRatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newMissingRatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show missing rate details",
		Long: `Show the full details of a missing rate.

Output Fields:
  ID             Missing rate identifier
  JOB ID         Job ID
  STUOM ID       Service type unit of measure ID
  CUSTOMER PPU   Customer price per unit
  TRUCKER PPU    Trucker price per unit
  CURRENCY       Currency code

Arguments:
  <id>  The missing rate ID (required). Use the list command to find IDs.`,
		Example: `  # Show a missing rate
  xbe view missing-rates show 123

  # Show as JSON
  xbe view missing-rates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMissingRatesShow,
	}
	initMissingRatesShowFlags(cmd)
	return cmd
}

func init() {
	missingRatesCmd.AddCommand(newMissingRatesShowCmd())
}

func initMissingRatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMissingRatesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMissingRatesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("missing rate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[missing-rates]", "job,service-type-unit-of-measure,currency-code,customer-price-per-unit,trucker-price-per-unit")

	body, _, err := client.Get(cmd.Context(), "/v1/missing-rates/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := missingRateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMissingRateDetails(cmd, details)
}

func parseMissingRatesShowOptions(cmd *cobra.Command) (missingRatesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return missingRatesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return missingRatesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return missingRatesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return missingRatesShowOptions{}, err
	}

	return missingRatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderMissingRateDetails(cmd *cobra.Command, details missingRateRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobID != "" {
		fmt.Fprintf(out, "Job ID: %s\n", details.JobID)
	}
	if details.ServiceTypeUnitOfMeasureID != "" {
		fmt.Fprintf(out, "Service Type Unit Of Measure ID: %s\n", details.ServiceTypeUnitOfMeasureID)
	}
	if details.CurrencyCode != "" {
		fmt.Fprintf(out, "Currency Code: %s\n", details.CurrencyCode)
	}
	if details.CustomerPricePerUnit != "" {
		fmt.Fprintf(out, "Customer Price Per Unit: %s\n", details.CustomerPricePerUnit)
	}
	if details.TruckerPricePerUnit != "" {
		fmt.Fprintf(out, "Trucker Price Per Unit: %s\n", details.TruckerPricePerUnit)
	}

	return nil
}
