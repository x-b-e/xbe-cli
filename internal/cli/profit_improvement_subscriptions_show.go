package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type profitImprovementSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type profitImprovementSubscriptionDetails struct {
	ID                      string `json:"id"`
	ContactMethod           string `json:"contact_method,omitempty"`
	ContactMethodEffective  string `json:"contact_method_effective,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	User                    string `json:"user,omitempty"`
	UserEmail               string `json:"user_email,omitempty"`
	UserMobile              string `json:"user_mobile,omitempty"`
	ProfitImprovementID     string `json:"profit_improvement_id,omitempty"`
	ProfitImprovementTitle  string `json:"profit_improvement_title,omitempty"`
	ProfitImprovementStatus string `json:"profit_improvement_status,omitempty"`
}

func newProfitImprovementSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show profit improvement subscription details",
		Long: `Show the full details of a profit improvement subscription.

Arguments:
  <id>  Profit improvement subscription ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a subscription
  xbe view profit-improvement-subscriptions show 123

  # JSON output
  xbe view profit-improvement-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProfitImprovementSubscriptionsShow,
	}
	initProfitImprovementSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	profitImprovementSubscriptionsCmd.AddCommand(newProfitImprovementSubscriptionsShowCmd())
}

func initProfitImprovementSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProfitImprovementSubscriptionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProfitImprovementSubscriptionsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("profit improvement subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[profit-improvement-subscriptions]", "contact-method,contact-method-effective,user,profit-improvement")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[profit-improvements]", "title,status")
	query.Set("include", "user,profit-improvement")

	body, _, err := client.Get(cmd.Context(), "/v1/profit-improvement-subscriptions/"+id, query)
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

	details := buildProfitImprovementSubscriptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProfitImprovementSubscriptionDetails(cmd, details)
}

func parseProfitImprovementSubscriptionsShowOptions(cmd *cobra.Command) (profitImprovementSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return profitImprovementSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProfitImprovementSubscriptionDetails(resp jsonAPISingleResponse) profitImprovementSubscriptionDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := profitImprovementSubscriptionDetails{
		ID:                     resource.ID,
		ContactMethod:          stringAttr(attrs, "contact-method"),
		ContactMethodEffective: stringAttr(attrs, "contact-method-effective"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.User = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
			details.UserMobile = stringAttr(user.Attributes, "mobile-number")
		}
	}

	if rel, ok := resource.Relationships["profit-improvement"]; ok && rel.Data != nil {
		details.ProfitImprovementID = rel.Data.ID
		if pi, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProfitImprovementTitle = stringAttr(pi.Attributes, "title")
			details.ProfitImprovementStatus = stringAttr(pi.Attributes, "status")
		}
	}

	return details
}

func renderProfitImprovementSubscriptionDetails(cmd *cobra.Command, details profitImprovementSubscriptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ContactMethod != "" {
		fmt.Fprintf(out, "Contact Method: %s\n", details.ContactMethod)
	}
	if details.ContactMethodEffective != "" {
		fmt.Fprintf(out, "Contact Method Effective: %s\n", details.ContactMethodEffective)
	}
	if details.User != "" {
		fmt.Fprintf(out, "User: %s\n", details.User)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.UserMobile != "" {
		fmt.Fprintf(out, "User Mobile: %s\n", details.UserMobile)
	}
	if details.ProfitImprovementTitle != "" {
		fmt.Fprintf(out, "Profit Improvement: %s\n", details.ProfitImprovementTitle)
	}
	if details.ProfitImprovementID != "" {
		fmt.Fprintf(out, "Profit Improvement ID: %s\n", details.ProfitImprovementID)
	}
	if details.ProfitImprovementStatus != "" {
		fmt.Fprintf(out, "Profit Improvement Status: %s\n", details.ProfitImprovementStatus)
	}

	return nil
}
