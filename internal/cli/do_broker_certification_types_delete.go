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

type doBrokerCertificationTypesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoBrokerCertificationTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a broker certification type",
		Long: `Delete a broker certification type.

Provide the broker certification type ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a broker certification type
  xbe do broker-certification-types delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerCertificationTypesDelete,
	}
	initDoBrokerCertificationTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doBrokerCertificationTypesCmd.AddCommand(newDoBrokerCertificationTypesDeleteCmd())
}

func initDoBrokerCertificationTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerCertificationTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerCertificationTypesDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a broker certification type")
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
		return fmt.Errorf("broker certification type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-certification-types]", "broker,certification-type")

	getBody, _, err := client.Get(cmd.Context(), "/v1/broker-certification-types/"+opts.ID, query)
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

	row := brokerCertificationTypeRowFromSingle(getResp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/broker-certification-types/"+opts.ID)
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

	if row.BrokerID != "" && row.CertificationTypeID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted broker certification type %s (broker %s, certification type %s)\n", row.ID, row.BrokerID, row.CertificationTypeID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted broker certification type %s\n", opts.ID)
	return nil
}

func parseDoBrokerCertificationTypesDeleteOptions(cmd *cobra.Command, args []string) (doBrokerCertificationTypesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerCertificationTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
