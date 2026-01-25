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

type doJobProductionPlanSubscriptionsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	ContactMethod string
}

func newDoJobProductionPlanSubscriptionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan subscription",
		Long: `Update an existing job production plan subscription.

Optional flags:
  --contact-method  Contact method (email_address, mobile_number)`,
		Example: `  # Update contact method
  xbe do job-production-plan-subscriptions update 123 --contact-method mobile_number`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanSubscriptionsUpdate,
	}
	initDoJobProductionPlanSubscriptionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanSubscriptionsCmd.AddCommand(newDoJobProductionPlanSubscriptionsUpdateCmd())
}

func initDoJobProductionPlanSubscriptionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("contact-method", "", "Contact method (email_address, mobile_number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanSubscriptionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanSubscriptionsUpdateOptions(cmd, args)
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
			"type":       "job-production-plan-subscriptions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-subscriptions/"+opts.ID, jsonBody)
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

	row := jobProductionPlanSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan subscription %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanSubscriptionsUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanSubscriptionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanSubscriptionsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		ContactMethod: contactMethod,
	}, nil
}
