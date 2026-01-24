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

type doLowestLosingBidPredictionSubjectDetailsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	PredictionSubjectID            string
	LowestBidAmount                string
	BidAmount                      string
	WalkAwayBidAmount              string
	EngineerEstimateAmount         string
	InternalEngineerEstimateAmount string
	BidDetails                     string
}

func newDoLowestLosingBidPredictionSubjectDetailsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lowest losing bid prediction subject detail",
		Long: `Create a lowest losing bid prediction subject detail.

Required flags:
  --prediction-subject              Prediction subject ID (required)

Optional flags:
  --lowest-bid-amount               Lowest bid amount
  --bid-amount                      Bid amount
  --walk-away-bid-amount            Walk away bid amount
  --engineer-estimate-amount        Engineer estimate amount
  --internal-engineer-estimate-amount Internal engineer estimate amount
  --bid-details                     Bid details JSON array (bidder_name, amount)

Notes:
  bid-details should be a JSON array, e.g.:
  [{"bidder_name":"Acme","amount":125000.50}]`,
		Example: `  # Create a detail record
  xbe do lowest-losing-bid-prediction-subject-details create \
    --prediction-subject 123 \
    --bid-amount 120000 \
    --walk-away-bid-amount 140000

  # Create with bid details
  xbe do lowest-losing-bid-prediction-subject-details create \
    --prediction-subject 123 \
    --bid-details '[{"bidder_name":"Acme","amount":125000.50}]'`,
		Args: cobra.NoArgs,
		RunE: runDoLowestLosingBidPredictionSubjectDetailsCreate,
	}
	initDoLowestLosingBidPredictionSubjectDetailsCreateFlags(cmd)
	return cmd
}

func init() {
	doLowestLosingBidPredictionSubjectDetailsCmd.AddCommand(newDoLowestLosingBidPredictionSubjectDetailsCreateCmd())
}

func initDoLowestLosingBidPredictionSubjectDetailsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID (required)")
	cmd.Flags().String("lowest-bid-amount", "", "Lowest bid amount")
	cmd.Flags().String("bid-amount", "", "Bid amount")
	cmd.Flags().String("walk-away-bid-amount", "", "Walk away bid amount")
	cmd.Flags().String("engineer-estimate-amount", "", "Engineer estimate amount")
	cmd.Flags().String("internal-engineer-estimate-amount", "", "Internal engineer estimate amount")
	cmd.Flags().String("bid-details", "", "Bid details JSON array")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLowestLosingBidPredictionSubjectDetailsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLowestLosingBidPredictionSubjectDetailsCreateOptions(cmd)
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

	if opts.PredictionSubjectID == "" {
		err := fmt.Errorf("--prediction-subject is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.LowestBidAmount != "" {
		attributes["lowest-bid-amount"] = opts.LowestBidAmount
	}
	if opts.BidAmount != "" {
		attributes["bid-amount"] = opts.BidAmount
	}
	if opts.WalkAwayBidAmount != "" {
		attributes["walk-away-bid-amount"] = opts.WalkAwayBidAmount
	}
	if opts.EngineerEstimateAmount != "" {
		attributes["engineer-estimate-amount"] = opts.EngineerEstimateAmount
	}
	if opts.InternalEngineerEstimateAmount != "" {
		attributes["internal-engineer-estimate-amount"] = opts.InternalEngineerEstimateAmount
	}
	if opts.BidDetails != "" {
		var details any
		if err := json.Unmarshal([]byte(opts.BidDetails), &details); err != nil {
			return fmt.Errorf("invalid bid details JSON: %w", err)
		}
		attributes["bid-details"] = details
	}

	relationships := map[string]any{
		"prediction-subject": map[string]any{
			"data": map[string]any{
				"type": "prediction-subjects",
				"id":   opts.PredictionSubjectID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lowest-losing-bid-prediction-subject-details",
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

	body, _, err := client.Post(cmd.Context(), "/v1/lowest-losing-bid-prediction-subject-details", jsonBody)
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

	row := lowestLosingBidPredictionSubjectDetailRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lowest losing bid prediction subject detail %s\n", row.ID)
	return nil
}

func parseDoLowestLosingBidPredictionSubjectDetailsCreateOptions(cmd *cobra.Command) (doLowestLosingBidPredictionSubjectDetailsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	lowestBidAmount, _ := cmd.Flags().GetString("lowest-bid-amount")
	bidAmount, _ := cmd.Flags().GetString("bid-amount")
	walkAwayBidAmount, _ := cmd.Flags().GetString("walk-away-bid-amount")
	engineerEstimateAmount, _ := cmd.Flags().GetString("engineer-estimate-amount")
	internalEngineerEstimateAmount, _ := cmd.Flags().GetString("internal-engineer-estimate-amount")
	bidDetails, _ := cmd.Flags().GetString("bid-details")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLowestLosingBidPredictionSubjectDetailsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		PredictionSubjectID:            predictionSubject,
		LowestBidAmount:                lowestBidAmount,
		BidAmount:                      bidAmount,
		WalkAwayBidAmount:              walkAwayBidAmount,
		EngineerEstimateAmount:         engineerEstimateAmount,
		InternalEngineerEstimateAmount: internalEngineerEstimateAmount,
		BidDetails:                     bidDetails,
	}, nil
}
