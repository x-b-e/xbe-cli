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

type doProfitImprovementSubscriptionsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	ContactMethod string
}

func newDoProfitImprovementSubscriptionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a profit improvement subscription",
		Long: `Update a profit improvement subscription.

Optional flags:
  --contact-method  Contact method (email_address, mobile_number)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update contact method
  xbe do profit-improvement-subscriptions update 123 --contact-method mobile_number`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProfitImprovementSubscriptionsUpdate,
	}
	initDoProfitImprovementSubscriptionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProfitImprovementSubscriptionsCmd.AddCommand(newDoProfitImprovementSubscriptionsUpdateCmd())
}

func initDoProfitImprovementSubscriptionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("contact-method", "", "Contact method (email_address, mobile_number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProfitImprovementSubscriptionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProfitImprovementSubscriptionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("contact-method") {
		attributes["contact-method"] = opts.ContactMethod
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "profit-improvement-subscriptions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/profit-improvement-subscriptions/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated profit improvement subscription %s\n", row.ID)
	return nil
}

func parseDoProfitImprovementSubscriptionsUpdateOptions(cmd *cobra.Command, args []string) (doProfitImprovementSubscriptionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProfitImprovementSubscriptionsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		ContactMethod: contactMethod,
	}, nil
}
