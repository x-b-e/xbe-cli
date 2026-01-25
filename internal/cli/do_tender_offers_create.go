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

type doTenderOffersCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	TenderID                    string
	TenderType                  string
	Comment                     string
	SkipCertificationValidation bool
}

func newDoTenderOffersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Offer a tender",
		Long: `Offer a tender.

Required flags:
  --tender-type   Tender type (customer-tenders, broker-tenders) (required)
  --tender        Tender ID (required)

Optional flags:
  --comment                       Offer comment
  --skip-certification-validation Skip certification requirement validation`,
		Example: `  # Offer a customer tender
  xbe do tender-offers create --tender-type customer-tenders --tender 12345

  # Offer with comment and skip certification validation
  xbe do tender-offers create --tender-type broker-tenders --tender 67890 \\
    --comment "Offering by dispatch" --skip-certification-validation

  # JSON output
  xbe do tender-offers create --tender-type customer-tenders --tender 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTenderOffersCreate,
	}
	initDoTenderOffersCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderOffersCmd.AddCommand(newDoTenderOffersCreateCmd())
}

func initDoTenderOffersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-type", "", "Tender type (customer-tenders, broker-tenders) (required)")
	cmd.Flags().String("tender", "", "Tender ID (required)")
	cmd.Flags().String("comment", "", "Offer comment")
	cmd.Flags().Bool("skip-certification-validation", false, "Skip certification requirement validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderOffersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderOffersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TenderType) == "" {
		err := fmt.Errorf("--tender-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TenderID) == "" {
		err := fmt.Errorf("--tender is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}
	if cmd.Flags().Changed("skip-certification-validation") {
		attributes["skip-certification-validation"] = opts.SkipCertificationValidation
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": opts.TenderType,
				"id":   opts.TenderID,
			},
		},
	}

	data := map[string]any{
		"type":          "tender-offers",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tender-offers", jsonBody)
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

	row := buildTenderOfferRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender offer %s\n", row.ID)
	return nil
}

func parseDoTenderOffersCreateOptions(cmd *cobra.Command) (doTenderOffersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderType, _ := cmd.Flags().GetString("tender-type")
	tenderID, _ := cmd.Flags().GetString("tender")
	comment, _ := cmd.Flags().GetString("comment")
	skipCertificationValidation, _ := cmd.Flags().GetBool("skip-certification-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderOffersCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		TenderID:                    tenderID,
		TenderType:                  tenderType,
		Comment:                     comment,
		SkipCertificationValidation: skipCertificationValidation,
	}, nil
}
