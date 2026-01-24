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

type doRateAgreementsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Name    string
	Status  string
	Seller  string
	Buyer   string
}

func newDoRateAgreementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new rate agreement",
		Long: `Create a new rate agreement.

Required flags:
  --status  Status (active/inactive) (required)
  --seller  Seller in Type|ID format (Broker or Trucker) (required)
  --buyer   Buyer in Type|ID format (Customer or Broker) (required)

Optional flags:
  --name    Rate agreement name

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a rate agreement for a broker/customer
  xbe do rate-agreements create --name "Standard" --status active --seller "Broker|123" --buyer "Customer|456"

  # Create a rate agreement for a trucker/broker
  xbe do rate-agreements create --status active --seller "Trucker|123" --buyer "Broker|456"

  # JSON output
  xbe do rate-agreements create --status active --seller "Broker|123" --buyer "Customer|456" --json`,
		Args: cobra.NoArgs,
		RunE: runDoRateAgreementsCreate,
	}
	initDoRateAgreementsCreateFlags(cmd)
	return cmd
}

func init() {
	doRateAgreementsCmd.AddCommand(newDoRateAgreementsCreateCmd())
}

func initDoRateAgreementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Rate agreement name")
	cmd.Flags().String("status", "", "Status (active/inactive) (required)")
	cmd.Flags().String("seller", "", "Seller in Type|ID format (Broker or Trucker) (required)")
	cmd.Flags().String("buyer", "", "Buyer in Type|ID format (Customer or Broker) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRateAgreementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRateAgreementsCreateOptions(cmd)
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

	if opts.Status == "" {
		err := fmt.Errorf("--status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	sellerType, sellerID, err := parseRateAgreementParty(opts.Seller, "seller")
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	buyerType, buyerID, err := parseRateAgreementParty(opts.Buyer, "buyer")
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"status": opts.Status,
	}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}

	relationships := map[string]any{
		"seller": map[string]any{
			"data": map[string]any{
				"type": sellerType,
				"id":   sellerID,
			},
		},
		"buyer": map[string]any{
			"data": map[string]any{
				"type": buyerType,
				"id":   buyerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "rate-agreements",
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

	body, _, err := client.Post(cmd.Context(), "/v1/rate-agreements", jsonBody)
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

	row := buildRateAgreementRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Name != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created rate agreement %s (%s)\n", row.ID, row.Name)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created rate agreement %s\n", row.ID)
	return nil
}

func parseDoRateAgreementsCreateOptions(cmd *cobra.Command) (doRateAgreementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	status, _ := cmd.Flags().GetString("status")
	seller, _ := cmd.Flags().GetString("seller")
	buyer, _ := cmd.Flags().GetString("buyer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAgreementsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Name:    name,
		Status:  status,
		Seller:  seller,
		Buyer:   buyer,
	}, nil
}
