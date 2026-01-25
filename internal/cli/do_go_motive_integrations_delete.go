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

type doGoMotiveIntegrationsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoGoMotiveIntegrationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a GoMotive integration",
		Long: `Delete a GoMotive integration.

Required flags:
  --confirm  Confirm deletion`,
		Example: `  # Delete a GoMotive integration
  xbe do go-motive-integrations delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoGoMotiveIntegrationsDelete,
	}
	initDoGoMotiveIntegrationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doGoMotiveIntegrationsCmd.AddCommand(newDoGoMotiveIntegrationsDeleteCmd())
}

func initDoGoMotiveIntegrationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGoMotiveIntegrationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGoMotiveIntegrationsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a GoMotive integration")
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[go-motive-integrations]", "integration-identifier,friendly-name,broker,integration-config")
	query.Set("include", "broker,integration-config")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[integration-configs]", "friendly-name")

	getBody, _, err := client.Get(cmd.Context(), "/v1/go-motive-integrations/"+opts.ID, query)
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

	row := goMotiveIntegrationRowFromSingle(getResp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/go-motive-integrations/"+opts.ID)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted GoMotive integration %s\n", opts.ID)
	return nil
}

func parseDoGoMotiveIntegrationsDeleteOptions(cmd *cobra.Command, args []string) (doGoMotiveIntegrationsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doGoMotiveIntegrationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
