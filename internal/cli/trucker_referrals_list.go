package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type truckerReferralsListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	Trucker       string
	Broker        string
	User          string
	ReferredOn    string
	ReferredOnMin string
	ReferredOnMax string
}

type truckerReferralRow struct {
	ID                           string `json:"id"`
	TruckerID                    string `json:"trucker_id,omitempty"`
	UserID                       string `json:"user_id,omitempty"`
	ReferredOn                   string `json:"referred_on,omitempty"`
	TruckerFirstShiftBonusAmount string `json:"trucker_first_shift_bonus_amount,omitempty"`
	TruckFirstShiftBonusAmount   string `json:"truck_first_shift_bonus_amount,omitempty"`
}

func newTruckerReferralsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker referrals",
		Long: `List trucker referrals.

Output Columns:
  ID             Trucker referral identifier
  TRUCKER        Trucker ID
  USER           Referring user ID
  REFERRED_ON    Referral date
  TRUCKER_BONUS  Trucker first shift bonus amount
  TRUCK_BONUS    Truck first shift bonus amount

Filters:
  --trucker           Filter by trucker ID
  --broker            Filter by broker ID
  --user              Filter by referring user ID
  --referred-on       Filter by referral date (YYYY-MM-DD)
  --referred-on-min   Filter by referral date on/after (YYYY-MM-DD)
  --referred-on-max   Filter by referral date on/before (YYYY-MM-DD)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trucker referrals
  xbe view trucker-referrals list

  # Filter by trucker
  xbe view trucker-referrals list --trucker 123

  # Filter by broker
  xbe view trucker-referrals list --broker 456

  # Filter by referral date range
  xbe view trucker-referrals list --referred-on-min 2025-01-01 --referred-on-max 2025-01-31

  # Output as JSON
  xbe view trucker-referrals list --json`,
		Args: cobra.NoArgs,
		RunE: runTruckerReferralsList,
	}
	initTruckerReferralsListFlags(cmd)
	return cmd
}

func init() {
	truckerReferralsCmd.AddCommand(newTruckerReferralsListCmd())
}

func initTruckerReferralsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("user", "", "Filter by referring user ID")
	cmd.Flags().String("referred-on", "", "Filter by referral date (YYYY-MM-DD)")
	cmd.Flags().String("referred-on-min", "", "Filter by referral date on/after (YYYY-MM-DD)")
	cmd.Flags().String("referred-on-max", "", "Filter by referral date on/before (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerReferralsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerReferralsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[trucker-referrals]", "referred-on,trucker-first-shift-bonus-amount,truck-first-shift-bonus-amount,trucker,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[referred-on]", opts.ReferredOn)
	setFilterIfPresent(query, "filter[referred-on-min]", opts.ReferredOnMin)
	setFilterIfPresent(query, "filter[referred-on-max]", opts.ReferredOnMax)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-referrals", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildTruckerReferralRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerReferralsTable(cmd, rows)
}

func parseTruckerReferralsListOptions(cmd *cobra.Command) (truckerReferralsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trucker, _ := cmd.Flags().GetString("trucker")
	broker, _ := cmd.Flags().GetString("broker")
	user, _ := cmd.Flags().GetString("user")
	referredOn, _ := cmd.Flags().GetString("referred-on")
	referredOnMin, _ := cmd.Flags().GetString("referred-on-min")
	referredOnMax, _ := cmd.Flags().GetString("referred-on-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerReferralsListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		Trucker:       trucker,
		Broker:        broker,
		User:          user,
		ReferredOn:    referredOn,
		ReferredOnMin: referredOnMin,
		ReferredOnMax: referredOnMax,
	}, nil
}

func buildTruckerReferralRows(resp jsonAPIResponse) []truckerReferralRow {
	rows := make([]truckerReferralRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := truckerReferralRow{
			ID:                           resource.ID,
			TruckerID:                    relationshipIDFromMap(resource.Relationships, "trucker"),
			UserID:                       relationshipIDFromMap(resource.Relationships, "user"),
			ReferredOn:                   formatDate(stringAttr(attrs, "referred-on")),
			TruckerFirstShiftBonusAmount: stringAttr(attrs, "trucker-first-shift-bonus-amount"),
			TruckFirstShiftBonusAmount:   stringAttr(attrs, "truck-first-shift-bonus-amount"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTruckerReferralsTable(cmd *cobra.Command, rows []truckerReferralRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker referrals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRUCKER\tUSER\tREFERRED_ON\tTRUCKER_BONUS\tTRUCK_BONUS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TruckerID,
			row.UserID,
			row.ReferredOn,
			row.TruckerFirstShiftBonusAmount,
			row.TruckFirstShiftBonusAmount,
		)
	}
	return writer.Flush()
}
