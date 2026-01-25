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

type doCustomerMembershipsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoCustomerMembershipsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a customer membership",
		Long: `Delete a customer membership.

This action is irreversible. The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The customer membership ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a customer membership
  xbe do customer-memberships delete 686 --confirm

  # Get JSON output of deleted record
  xbe do customer-memberships delete 686 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerMembershipsDelete,
	}
	initDoCustomerMembershipsDeleteFlags(cmd)
	return cmd
}

func init() {
	doCustomerMembershipsCmd.AddCommand(newDoCustomerMembershipsDeleteCmd())
}

func initDoCustomerMembershipsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerMembershipsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerMembershipsDeleteOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("customer membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")

	getBody, _, err := client.Get(cmd.Context(), "/v1/customer-memberships/"+id, query)
	if err != nil {
		if len(getBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(getBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(getBody, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildMembershipDetails(resp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/customer-memberships/"+id)
	if err != nil {
		if len(deleteBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(deleteBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted customer membership %s (%s in %s)\n",
		details.ID,
		details.UserName,
		details.OrganizationName,
	)
	return nil
}

func parseDoCustomerMembershipsDeleteOptions(cmd *cobra.Command) (doCustomerMembershipsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerMembershipsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
