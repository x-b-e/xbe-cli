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

type doCertificationTypesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoCertificationTypesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a certification type",
		Long: `Delete a certification type.

This permanently deletes the certification type if it has no associated certifications.

The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The certification type ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a certification type
  xbe do certification-types delete 456 --confirm

  # Get JSON output of deleted record
  xbe do certification-types delete 456 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCertificationTypesDelete,
	}
	initDoCertificationTypesDeleteFlags(cmd)
	return cmd
}

func init() {
	doCertificationTypesCmd.AddCommand(newDoCertificationTypesDeleteCmd())
}

func initDoCertificationTypesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCertificationTypesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCertificationTypesDeleteOptions(cmd)
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
		return fmt.Errorf("certification type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// First, fetch the record so we can show what was deleted
	query := url.Values{}
	query.Set("fields[certification-types]", "name,can-apply-to")

	getBody, _, err := client.Get(cmd.Context(), "/v1/certification-types/"+id, query)
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
	row := buildCertificationTypeRowFromSingle(getResp)

	// Delete the record
	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/certification-types/"+id)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted certification type %s (%s)\n", id, name)
	return nil
}

func parseDoCertificationTypesDeleteOptions(cmd *cobra.Command) (doCertificationTypesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCertificationTypesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
