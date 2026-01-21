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

type doLaborClassificationsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoLaborClassificationsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a labor classification",
		Long: `Delete a labor classification.

This permanently deletes the labor classification.

The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The labor classification ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a labor classification
  xbe do labor-classifications delete 123 --confirm

  # Get JSON output of deleted record
  xbe do labor-classifications delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLaborClassificationsDelete,
	}
	initDoLaborClassificationsDeleteFlags(cmd)
	return cmd
}

func init() {
	doLaborClassificationsCmd.AddCommand(newDoLaborClassificationsDeleteCmd())
}

func initDoLaborClassificationsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLaborClassificationsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLaborClassificationsDeleteOptions(cmd)
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
		return fmt.Errorf("labor classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// First, fetch the record so we can show what was deleted
	query := url.Values{}
	query.Set("fields[labor-classifications]", "name,abbreviation")

	getBody, _, err := client.Get(cmd.Context(), "/v1/labor-classifications/"+id, query)
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
	row := buildLaborClassificationRow(getResp)

	// Delete the record
	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/labor-classifications/"+id)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted labor classification %s (%s)\n", id, name)
	return nil
}

func parseDoLaborClassificationsDeleteOptions(cmd *cobra.Command) (doLaborClassificationsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLaborClassificationsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
