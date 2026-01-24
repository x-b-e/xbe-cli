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

type doTruckerMembershipsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoTruckerMembershipsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a trucker membership",
		Long: `Delete a trucker membership.

This action is irreversible. The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The trucker membership ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a trucker membership
  xbe do trucker-memberships delete 686 --confirm

  # Get JSON output of deleted record
  xbe do trucker-memberships delete 686 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerMembershipsDelete,
	}
	initDoTruckerMembershipsDeleteFlags(cmd)
	return cmd
}

func init() {
	doTruckerMembershipsCmd.AddCommand(newDoTruckerMembershipsDeleteCmd())
}

func initDoTruckerMembershipsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerMembershipsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerMembershipsDeleteOptions(cmd)
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
		return fmt.Errorf("trucker membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")

	getBody, _, err := client.Get(cmd.Context(), "/v1/trucker-memberships/"+id, query)
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

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/trucker-memberships/"+id)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted trucker membership %s (%s in %s)\n",
		details.ID,
		details.UserName,
		details.OrganizationName,
	)
	return nil
}

func parseDoTruckerMembershipsDeleteOptions(cmd *cobra.Command) (doTruckerMembershipsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerMembershipsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
