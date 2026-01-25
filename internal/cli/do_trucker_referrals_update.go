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

type doTruckerReferralsUpdateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ID                           string
	Trucker                      string
	User                         string
	Notes                        string
	ReferredOn                   string
	TruckerFirstShiftBonusAmount string
	TruckFirstShiftBonusAmount   string
}

func newDoTruckerReferralsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker referral",
		Long: `Update an existing trucker referral.

Arguments:
  <id>    The trucker referral ID (required)

Optional flags:
  --trucker                          Trucker ID
  --user                             Referring user ID
  --notes                            Referral notes
  --referred-on                      Referral date (YYYY-MM-DD)
  --trucker-first-shift-bonus-amount Trucker first shift bonus amount
  --truck-first-shift-bonus-amount   Truck first shift bonus amount

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a trucker referral
  xbe do trucker-referrals update 123 \
    --notes "Updated notes" \
    --trucker-first-shift-bonus-amount 300.00

  # Update trucker and user
  xbe do trucker-referrals update 123 --trucker 456 --user 789

  # Get JSON output
  xbe do trucker-referrals update 123 --notes "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerReferralsUpdate,
	}
	initDoTruckerReferralsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerReferralsCmd.AddCommand(newDoTruckerReferralsUpdateCmd())
}

func initDoTruckerReferralsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("user", "", "Referring user ID")
	cmd.Flags().String("notes", "", "Referral notes")
	cmd.Flags().String("referred-on", "", "Referral date (YYYY-MM-DD)")
	cmd.Flags().String("trucker-first-shift-bonus-amount", "", "Trucker first shift bonus amount")
	cmd.Flags().String("truck-first-shift-bonus-amount", "", "Truck first shift bonus amount")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerReferralsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerReferralsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("referred-on") {
		attributes["referred-on"] = opts.ReferredOn
	}
	if cmd.Flags().Changed("trucker-first-shift-bonus-amount") {
		attributes["trucker-first-shift-bonus-amount"] = opts.TruckerFirstShiftBonusAmount
	}
	if cmd.Flags().Changed("truck-first-shift-bonus-amount") {
		attributes["truck-first-shift-bonus-amount"] = opts.TruckFirstShiftBonusAmount
	}

	if cmd.Flags().Changed("trucker") {
		if strings.TrimSpace(opts.Trucker) == "" {
			relationships["trucker"] = map[string]any{"data": nil}
		} else {
			relationships["trucker"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.Trucker,
				},
			}
		}
	}
	if cmd.Flags().Changed("user") {
		if strings.TrimSpace(opts.User) == "" {
			relationships["user"] = map[string]any{"data": nil}
		} else {
			relationships["user"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.User,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "trucker-referrals",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-referrals/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker referral %s\n", details.ID)
	return nil
}

func parseDoTruckerReferralsUpdateOptions(cmd *cobra.Command, args []string) (doTruckerReferralsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	user, _ := cmd.Flags().GetString("user")
	notes, _ := cmd.Flags().GetString("notes")
	referredOn, _ := cmd.Flags().GetString("referred-on")
	truckerFirstShiftBonusAmount, _ := cmd.Flags().GetString("trucker-first-shift-bonus-amount")
	truckFirstShiftBonusAmount, _ := cmd.Flags().GetString("truck-first-shift-bonus-amount")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerReferralsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
		Trucker:                      trucker,
		User:                         user,
		Notes:                        notes,
		ReferredOn:                   referredOn,
		TruckerFirstShiftBonusAmount: truckerFirstShiftBonusAmount,
		TruckFirstShiftBonusAmount:   truckFirstShiftBonusAmount,
	}, nil
}
