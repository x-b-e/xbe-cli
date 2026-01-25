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

type doPromptersDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoPromptersDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a prompter",
		Long: `Delete a prompter.

Arguments:
  <id>    The prompter ID (required)

Flags:
  --confirm    Required flag to confirm deletion

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Delete a prompter
  xbe do prompters delete 123 --confirm

  # Get JSON output of the record
  xbe do prompters delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPromptersDelete,
	}
	initDoPromptersDeleteFlags(cmd)
	return cmd
}

func init() {
	doPromptersCmd.AddCommand(newDoPromptersDeleteCmd())
}

func initDoPromptersDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPromptersDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPromptersDeleteOptions(cmd)
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
		return fmt.Errorf("prompter id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prompters]", "name,is-active")

	getBody, _, err := client.Get(cmd.Context(), "/v1/prompters/"+id, query)
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

	row := buildPrompterRowFromSingle(resp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/prompters/"+id)
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

	if row.Name != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted prompter %s (%s)\n", row.ID, row.Name)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted prompter %s\n", row.ID)
	return nil
}

func parseDoPromptersDeleteOptions(cmd *cobra.Command) (doPromptersDeleteOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doPromptersDeleteOptions{}, err
	}
	confirm, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return doPromptersDeleteOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doPromptersDeleteOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doPromptersDeleteOptions{}, err
	}

	return doPromptersDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
