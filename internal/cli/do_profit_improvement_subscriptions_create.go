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

type doProfitImprovementSubscriptionsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	UserID            string
	ProfitImprovement string
	ContactMethod     string
}

func newDoProfitImprovementSubscriptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a profit improvement subscription",
		Long: `Create a profit improvement subscription.

Required flags:
  --user               User ID
  --profit-improvement Profit improvement ID

Optional flags:
  --contact-method     Contact method (email_address, mobile_number)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a subscription
  xbe do profit-improvement-subscriptions create --user 123 --profit-improvement 456 --contact-method email_address

  # Create using default contact method
  xbe do profit-improvement-subscriptions create --user 123 --profit-improvement 456`,
		RunE: runDoProfitImprovementSubscriptionsCreate,
	}
	initDoProfitImprovementSubscriptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doProfitImprovementSubscriptionsCmd.AddCommand(newDoProfitImprovementSubscriptionsCreateCmd())
}

func initDoProfitImprovementSubscriptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("profit-improvement", "", "Profit improvement ID (required)")
	cmd.Flags().String("contact-method", "", "Contact method (email_address, mobile_number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("profit-improvement")
}

func runDoProfitImprovementSubscriptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProfitImprovementSubscriptionsCreateOptions(cmd)
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
	if opts.ContactMethod != "" {
		attributes["contact-method"] = opts.ContactMethod
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
		"profit-improvement": map[string]any{
			"data": map[string]any{
				"type": "profit-improvements",
				"id":   opts.ProfitImprovement,
			},
		},
	}

	data := map[string]any{
		"type":          "profit-improvement-subscriptions",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/profit-improvement-subscriptions", jsonBody)
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

	row := profitImprovementSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created profit improvement subscription %s\n", row.ID)
	return nil
}

func parseDoProfitImprovementSubscriptionsCreateOptions(cmd *cobra.Command) (doProfitImprovementSubscriptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	profitImprovement, _ := cmd.Flags().GetString("profit-improvement")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProfitImprovementSubscriptionsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		UserID:            userID,
		ProfitImprovement: profitImprovement,
		ContactMethod:     contactMethod,
	}, nil
}

func profitImprovementSubscriptionRowFromSingle(resp jsonAPISingleResponse) profitImprovementSubscriptionRow {
	resource := resp.Data
	attrs := resource.Attributes

	row := profitImprovementSubscriptionRow{
		ID:                     resource.ID,
		ContactMethod:          stringAttr(attrs, "contact-method"),
		ContactMethodEffective: stringAttr(attrs, "contact-method-effective"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["profit-improvement"]; ok && rel.Data != nil {
		row.ProfitImprovementID = rel.Data.ID
	}

	return row
}
