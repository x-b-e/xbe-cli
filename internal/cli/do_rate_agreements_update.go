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

type doRateAgreementsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Name    string
	Status  string
	Seller  string
	Buyer   string
}

func newDoRateAgreementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing rate agreement",
		Long: `Update an existing rate agreement.

Provide the rate agreement ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name    Rate agreement name
  --status  Status (active/inactive)
  --seller  Seller in Type|ID format (Broker or Trucker)
  --buyer   Buyer in Type|ID format (Customer or Broker)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update name
  xbe do rate-agreements update 123 --name "Updated Standard"

  # Update status
  xbe do rate-agreements update 123 --status inactive

  # Update seller and buyer
  xbe do rate-agreements update 123 --seller "Broker|456" --buyer "Customer|789"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRateAgreementsUpdate,
	}
	initDoRateAgreementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRateAgreementsCmd.AddCommand(newDoRateAgreementsUpdateCmd())
}

func initDoRateAgreementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Rate agreement name")
	cmd.Flags().String("status", "", "Status (active/inactive)")
	cmd.Flags().String("seller", "", "Seller in Type|ID format (Broker or Trucker)")
	cmd.Flags().String("buyer", "", "Buyer in Type|ID format (Customer or Broker)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRateAgreementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRateAgreementsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("seller") {
		sellerType, sellerID, err := parseRateAgreementParty(opts.Seller, "seller")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["seller"] = map[string]any{
			"data": map[string]any{
				"type": sellerType,
				"id":   sellerID,
			},
		}
	}
	if cmd.Flags().Changed("buyer") {
		buyerType, buyerID, err := parseRateAgreementParty(opts.Buyer, "buyer")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["buyer"] = map[string]any{
			"data": map[string]any{
				"type": buyerType,
				"id":   buyerID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --status, --seller, --buyer")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "rate-agreements",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/rate-agreements/"+opts.ID, jsonBody)
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
		fmt.Fprintf(cmd.OutOrStdout(), "Updated rate agreement %s (%s)\n", row.ID, row.Name)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated rate agreement %s\n", row.ID)
	return nil
}

func parseDoRateAgreementsUpdateOptions(cmd *cobra.Command, args []string) (doRateAgreementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	status, _ := cmd.Flags().GetString("status")
	seller, _ := cmd.Flags().GetString("seller")
	buyer, _ := cmd.Flags().GetString("buyer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAgreementsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Name:    name,
		Status:  status,
		Seller:  seller,
		Buyer:   buyer,
	}, nil
}
