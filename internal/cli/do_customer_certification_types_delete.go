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

type doCustomerCertificationTypesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoCustomerCertificationTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a customer certification type",
		Long: `Delete a customer certification type.

Provide the customer certification type ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a customer certification type
  xbe do customer-certification-types delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerCertificationTypesDelete,
	}
	initDoCustomerCertificationTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doCustomerCertificationTypesCmd.AddCommand(newDoCustomerCertificationTypesDeleteCmd())
}

func initDoCustomerCertificationTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerCertificationTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerCertificationTypesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a customer certification type")
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
		return fmt.Errorf("customer certification type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-certification-types]", "customer,certification-type")

	getBody, _, err := client.Get(cmd.Context(), "/v1/customer-certification-types/"+opts.ID, query)
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

	row := customerCertificationTypeRowFromSingle(getResp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/customer-certification-types/"+opts.ID)
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

	if row.CustomerID != "" && row.CertificationTypeID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted customer certification type %s (customer %s, certification type %s)\n", row.ID, row.CustomerID, row.CertificationTypeID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted customer certification type %s\n", opts.ID)
	return nil
}

func parseDoCustomerCertificationTypesDeleteOptions(cmd *cobra.Command, args []string) (doCustomerCertificationTypesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerCertificationTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
