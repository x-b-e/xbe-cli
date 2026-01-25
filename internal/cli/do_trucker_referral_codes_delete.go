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

type doTruckerReferralCodesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoTruckerReferralCodesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a trucker referral code",
		Long: `Delete a trucker referral code.

Provide the trucker referral code ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a trucker referral code
  xbe do trucker-referral-codes delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerReferralCodesDelete,
	}
	initDoTruckerReferralCodesDeleteFlags(cmd)
	return cmd
}

func init() {
	doTruckerReferralCodesCmd.AddCommand(newDoTruckerReferralCodesDeleteCmd())
}

func initDoTruckerReferralCodesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerReferralCodesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerReferralCodesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a trucker referral code")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.ID) == "" {
		return fmt.Errorf("trucker referral code id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[trucker-referral-codes]", "code,value,broker")

	getBody, _, err := client.Get(cmd.Context(), "/v1/trucker-referral-codes/"+opts.ID, query)
	if err != nil {
		if len(getBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(getBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var getResp jsonAPISingleResponse
	if err := json.Unmarshal(getBody, &getResp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := truckerReferralCodeRowFromSingle(getResp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/trucker-referral-codes/"+opts.ID)
	if err != nil {
		if len(deleteBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(deleteBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Code != "" && row.BrokerID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted trucker referral code %s (code %s, broker %s)\n", row.ID, row.Code, row.BrokerID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted trucker referral code %s\n", opts.ID)
	return nil
}

func parseDoTruckerReferralCodesDeleteOptions(cmd *cobra.Command, args []string) (doTruckerReferralCodesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerReferralCodesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
