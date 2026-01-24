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

type doPredictionSubjectBidsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Amount  float64
}

func newDoPredictionSubjectBidsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction subject bid",
		Long: `Update a prediction subject bid.

Provide the bid ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --amount  Bid amount

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a bid amount
  xbe do prediction-subject-bids update 123 --amount 125000

  # Get JSON output
  xbe do prediction-subject-bids update 123 --amount 125000 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectBidsUpdate,
	}
	initDoPredictionSubjectBidsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectBidsCmd.AddCommand(newDoPredictionSubjectBidsUpdateCmd())
}

func initDoPredictionSubjectBidsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Float64("amount", 0, "Bid amount")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectBidsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectBidsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("amount") {
		attributes["amount"] = opts.Amount
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --amount")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prediction-subject-bids",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/prediction-subject-bids/"+opts.ID, jsonBody)
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

	message := fmt.Sprintf("Updated prediction subject bid %s", row.ID)
	if row.Amount != nil {
		message = fmt.Sprintf("%s (amount %s)", message, formatPredictionSubjectBidAmount(row.Amount))
	}
	fmt.Fprintln(cmd.OutOrStdout(), message)
	return nil
}

func parseDoPredictionSubjectBidsUpdateOptions(cmd *cobra.Command, args []string) (doPredictionSubjectBidsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	amount, _ := cmd.Flags().GetFloat64("amount")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectBidsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Amount:  amount,
	}, nil
}
