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

type doDeveloperMembershipsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoDeveloperMembershipsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a developer membership",
		Long: `Delete a developer membership.

This action is irreversible. The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The developer membership ID (required)

Flags:
  --confirm    Required flag to confirm deletion

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Delete a developer membership
  xbe do developer-memberships delete 686 --confirm

  # Get JSON output of deleted record
  xbe do developer-memberships delete 686 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperMembershipsDelete,
	}
	initDoDeveloperMembershipsDeleteFlags(cmd)
	return cmd
}

func init() {
	doDeveloperMembershipsCmd.AddCommand(newDoDeveloperMembershipsDeleteCmd())
}

func initDoDeveloperMembershipsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperMembershipsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperMembershipsDeleteOptions(cmd)
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
		return fmt.Errorf("developer membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "user,organization,broker")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[developers]", "name")
	query.Set("fields[brokers]", "company-name")

	getBody, _, err := client.Get(cmd.Context(), "/v1/developer-memberships/"+id, query)
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

	details := buildDeveloperMembershipDetails(resp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/developer-memberships/"+id)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted developer membership %s (%s in %s)\n",
		details.ID,
		details.UserName,
		details.DeveloperName,
	)
	return nil
}

func parseDoDeveloperMembershipsDeleteOptions(cmd *cobra.Command) (doDeveloperMembershipsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperMembershipsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
