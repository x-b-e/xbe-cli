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

type doGlossaryTermsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoGlossaryTermsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a glossary term",
		Long: `Delete a glossary term.

This action is irreversible. The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>    The glossary term ID (required)

Flags:
  --confirm    Required flag to confirm deletion`,
		Example: `  # Delete a glossary term
  xbe do glossary-terms delete 123 --confirm

  # Get JSON output of deleted record
  xbe do glossary-terms delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoGlossaryTermsDelete,
	}
	initDoGlossaryTermsDeleteFlags(cmd)
	return cmd
}

func init() {
	doGlossaryTermsCmd.AddCommand(newDoGlossaryTermsDeleteCmd())
}

func initDoGlossaryTermsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGlossaryTermsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGlossaryTermsDeleteOptions(cmd)
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
		return fmt.Errorf("glossary term id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// First, fetch the record so we can show what was deleted
	query := url.Values{}
	query.Set("fields[glossary-terms]", "term,definition,source")

	getBody, _, err := client.Get(cmd.Context(), "/v1/glossary-terms/"+id, query)
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

	details := buildGlossaryTermDetails(resp)

	// Now delete the record
	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/glossary-terms/"+id)
	if err != nil {
		if len(deleteBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(deleteBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Output the deleted record
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted glossary term %s (%s)\n", details.ID, details.Term)
	return nil
}

func parseDoGlossaryTermsDeleteOptions(cmd *cobra.Command) (doGlossaryTermsDeleteOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doGlossaryTermsDeleteOptions{}, err
	}
	confirm, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return doGlossaryTermsDeleteOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doGlossaryTermsDeleteOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doGlossaryTermsDeleteOptions{}, err
	}

	return doGlossaryTermsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
