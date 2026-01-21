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

type doExternalIdentificationTypesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoExternalIdentificationTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an external identification type",
		Long: `Delete an external identification type.

This permanently deletes the identification type.

The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The identification type ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete an external identification type
  xbe do external-identification-types delete 123 --confirm

  # Get JSON output of deleted record
  xbe do external-identification-types delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoExternalIdentificationTypesDelete,
	}
	initDoExternalIdentificationTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doExternalIdentificationTypesCmd.AddCommand(newDoExternalIdentificationTypesDeleteCmd())
}

func initDoExternalIdentificationTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExternalIdentificationTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExternalIdentificationTypesDeleteOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require --confirm flag
	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required for deletion")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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
		return fmt.Errorf("external identification type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// First, fetch the record so we can show what was deleted
	query := url.Values{}
	query.Set("fields[external-identification-types]", "name")

	getBody, _, err := client.Get(cmd.Context(), "/v1/external-identification-types/"+id, query)
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

	// Store name for confirmation message
	name := stringAttr(getResp.Data.Attributes, "name")
	row := buildExternalIdentificationTypeRow(getResp)

	// Delete the record
	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/external-identification-types/"+id)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted external identification type %s (%s)\n", id, name)
	return nil
}

func parseDoExternalIdentificationTypesDeleteOptions(cmd *cobra.Command) (doExternalIdentificationTypesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExternalIdentificationTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
