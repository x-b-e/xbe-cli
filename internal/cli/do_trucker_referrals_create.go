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

type doTruckerReferralsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	Trucker                      string
	User                         string
	Notes                        string
	ReferredOn                   string
	TruckerFirstShiftBonusAmount string
	TruckFirstShiftBonusAmount   string
}

func newDoTruckerReferralsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trucker referral",
		Long: `Create a trucker referral.

Required flags:
  --trucker    Trucker ID
  --user       Referring user ID

Optional flags:
  --notes                            Referral notes
  --referred-on                      Referral date (YYYY-MM-DD)
  --trucker-first-shift-bonus-amount Trucker first shift bonus amount
  --truck-first-shift-bonus-amount   Truck first shift bonus amount

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a trucker referral
  xbe do trucker-referrals create \
    --trucker 123 \
    --user 456 \
    --notes "Referral from job fair" \
    --referred-on 2025-01-15 \
    --trucker-first-shift-bonus-amount 250.00 \
    --truck-first-shift-bonus-amount 100.00

  # Get JSON output
  xbe do trucker-referrals create --trucker 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTruckerReferralsCreate,
	}
	initDoTruckerReferralsCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerReferralsCmd.AddCommand(newDoTruckerReferralsCreateCmd())
}

func initDoTruckerReferralsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().String("user", "", "Referring user ID (required)")
	cmd.Flags().String("notes", "", "Referral notes")
	cmd.Flags().String("referred-on", "", "Referral date (YYYY-MM-DD)")
	cmd.Flags().String("trucker-first-shift-bonus-amount", "", "Trucker first shift bonus amount")
	cmd.Flags().String("truck-first-shift-bonus-amount", "", "Truck first shift bonus amount")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerReferralsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerReferralsCreateOptions(cmd)
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

	truckerID := strings.TrimSpace(opts.Trucker)
	if truckerID == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	userID := strings.TrimSpace(opts.User)
	if userID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Notes) != "" {
		attributes["notes"] = opts.Notes
	}
	if strings.TrimSpace(opts.ReferredOn) != "" {
		attributes["referred-on"] = opts.ReferredOn
	}
	if strings.TrimSpace(opts.TruckerFirstShiftBonusAmount) != "" {
		attributes["trucker-first-shift-bonus-amount"] = opts.TruckerFirstShiftBonusAmount
	}
	if strings.TrimSpace(opts.TruckFirstShiftBonusAmount) != "" {
		attributes["truck-first-shift-bonus-amount"] = opts.TruckFirstShiftBonusAmount
	}

	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   truckerID,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   userID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trucker-referrals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-referrals", jsonBody)
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

	details := buildTruckerReferralDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker referral %s\n", details.ID)
	return nil
}

func parseDoTruckerReferralsCreateOptions(cmd *cobra.Command) (doTruckerReferralsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	user, _ := cmd.Flags().GetString("user")
	notes, _ := cmd.Flags().GetString("notes")
	referredOn, _ := cmd.Flags().GetString("referred-on")
	truckerFirstShiftBonusAmount, _ := cmd.Flags().GetString("trucker-first-shift-bonus-amount")
	truckFirstShiftBonusAmount, _ := cmd.Flags().GetString("truck-first-shift-bonus-amount")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerReferralsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		Trucker:                      trucker,
		User:                         user,
		Notes:                        notes,
		ReferredOn:                   referredOn,
		TruckerFirstShiftBonusAmount: truckerFirstShiftBonusAmount,
		TruckFirstShiftBonusAmount:   truckFirstShiftBonusAmount,
	}, nil
}
