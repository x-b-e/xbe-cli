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

type truckerReferralsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type truckerReferralDetails struct {
	ID                           string `json:"id"`
	TruckerID                    string `json:"trucker_id,omitempty"`
	UserID                       string `json:"user_id,omitempty"`
	Notes                        string `json:"notes,omitempty"`
	ReferredOn                   string `json:"referred_on,omitempty"`
	TruckerFirstShiftBonusAmount string `json:"trucker_first_shift_bonus_amount,omitempty"`
	TruckFirstShiftBonusAmount   string `json:"truck_first_shift_bonus_amount,omitempty"`
	CreatedAt                    string `json:"created_at,omitempty"`
	UpdatedAt                    string `json:"updated_at,omitempty"`
}

func newTruckerReferralsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trucker referral details",
		Long: `Show the full details of a trucker referral.

Output Fields:
  ID
  Trucker ID
  User ID
  Notes
  Referred On
  Trucker First Shift Bonus Amount
  Truck First Shift Bonus Amount
  Created At
  Updated At

Arguments:
  <id>    The trucker referral ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a trucker referral
  xbe view trucker-referrals show 123

  # Get JSON output
  xbe view trucker-referrals show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTruckerReferralsShow,
	}
	initTruckerReferralsShowFlags(cmd)
	return cmd
}

func init() {
	truckerReferralsCmd.AddCommand(newTruckerReferralsShowCmd())
}

func initTruckerReferralsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerReferralsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTruckerReferralsShowOptions(cmd)
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
		return fmt.Errorf("trucker referral id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-referrals/"+id, nil)
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

	details := buildTruckerReferralDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTruckerReferralDetails(cmd, details)
}

func parseTruckerReferralsShowOptions(cmd *cobra.Command) (truckerReferralsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerReferralsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTruckerReferralDetails(resp jsonAPISingleResponse) truckerReferralDetails {
	attrs := resp.Data.Attributes
	details := truckerReferralDetails{
		ID:                           resp.Data.ID,
		Notes:                        stringAttr(attrs, "notes"),
		ReferredOn:                   formatDate(stringAttr(attrs, "referred-on")),
		TruckerFirstShiftBonusAmount: stringAttr(attrs, "trucker-first-shift-bonus-amount"),
		TruckFirstShiftBonusAmount:   stringAttr(attrs, "truck-first-shift-bonus-amount"),
		CreatedAt:                    formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                    formatDateTime(stringAttr(attrs, "updated-at")),
	}

	details.TruckerID = relationshipIDFromMap(resp.Data.Relationships, "trucker")
	details.UserID = relationshipIDFromMap(resp.Data.Relationships, "user")

	return details
}

func renderTruckerReferralDetails(cmd *cobra.Command, details truckerReferralDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.ReferredOn != "" {
		fmt.Fprintf(out, "Referred On: %s\n", details.ReferredOn)
	}
	if details.TruckerFirstShiftBonusAmount != "" {
		fmt.Fprintf(out, "Trucker First Shift Bonus Amount: %s\n", details.TruckerFirstShiftBonusAmount)
	}
	if details.TruckFirstShiftBonusAmount != "" {
		fmt.Fprintf(out, "Truck First Shift Bonus Amount: %s\n", details.TruckFirstShiftBonusAmount)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
