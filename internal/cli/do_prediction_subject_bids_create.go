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

type doPredictionSubjectBidsCreateOptions struct {
	BaseURL                                string
	Token                                  string
	JSON                                   bool
	Amount                                 float64
	Bidder                                 string
	LowestLosingBidPredictionSubjectDetail string
}

func newDoPredictionSubjectBidsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction subject bid",
		Long: `Create a prediction subject bid.

Required:
  --bidder                                   Bidder ID
  --lowest-losing-bid-prediction-subject-detail  Lowest losing bid detail ID

Optional:
  --amount  Bid amount

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a prediction subject bid
  xbe do prediction-subject-bids create \
    --bidder 123 \
    --lowest-losing-bid-prediction-subject-detail 456 \
    --amount 120000

  # Create a bid without specifying amount
  xbe do prediction-subject-bids create \
    --bidder 123 \
    --lowest-losing-bid-prediction-subject-detail 456`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionSubjectBidsCreate,
	}
	initDoPredictionSubjectBidsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectBidsCmd.AddCommand(newDoPredictionSubjectBidsCreateCmd())
}

func initDoPredictionSubjectBidsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Float64("amount", 0, "Bid amount")
	cmd.Flags().String("bidder", "", "Bidder ID (required)")
	cmd.Flags().String("lowest-losing-bid-prediction-subject-detail", "", "Lowest losing bid detail ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectBidsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionSubjectBidsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Bidder) == "" {
		err := fmt.Errorf("--bidder is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.LowestLosingBidPredictionSubjectDetail) == "" {
		err := fmt.Errorf("--lowest-losing-bid-prediction-subject-detail is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("amount") {
		attributes["amount"] = opts.Amount
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prediction-subject-bids",
			"attributes": attributes,
			"relationships": map[string]any{
				"bidder": map[string]any{
					"data": map[string]any{
						"type": "bidders",
						"id":   opts.Bidder,
					},
				},
				"lowest-losing-bid-prediction-subject-detail": map[string]any{
					"data": map[string]any{
						"type": "lowest-losing-bid-prediction-subject-details",
						"id":   opts.LowestLosingBidPredictionSubjectDetail,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-subject-bids", jsonBody)
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

	row := predictionSubjectBidRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	message := fmt.Sprintf("Created prediction subject bid %s", row.ID)
	if row.Amount != nil {
		message = fmt.Sprintf("%s (amount %s)", message, formatPredictionSubjectBidAmount(row.Amount))
	}
	fmt.Fprintln(cmd.OutOrStdout(), message)
	return nil
}

func parseDoPredictionSubjectBidsCreateOptions(cmd *cobra.Command) (doPredictionSubjectBidsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	amount, _ := cmd.Flags().GetFloat64("amount")
	bidder, _ := cmd.Flags().GetString("bidder")
	lowestDetail, _ := cmd.Flags().GetString("lowest-losing-bid-prediction-subject-detail")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectBidsCreateOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		Amount:                                 amount,
		Bidder:                                 bidder,
		LowestLosingBidPredictionSubjectDetail: lowestDetail,
	}, nil
}
