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

type doBaseSummaryTemplatesDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Confirm bool
}

func newDoBaseSummaryTemplatesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a base summary template",
		Long: `Delete a base summary template.

This action is irreversible. The --confirm flag is required to prevent accidental deletion.

Arguments:
  <id>  The base summary template ID (required)

Flags:
  --confirm  Required flag to confirm deletion

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Delete a base summary template
  xbe do base-summary-templates delete 123 --confirm

  # JSON output of deleted record
  xbe do base-summary-templates delete 123 --confirm --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBaseSummaryTemplatesDelete,
	}
	initDoBaseSummaryTemplatesDeleteFlags(cmd)
	return cmd
}

func init() {
	doBaseSummaryTemplatesCmd.AddCommand(newDoBaseSummaryTemplatesDeleteCmd())
}

func initDoBaseSummaryTemplatesDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBaseSummaryTemplatesDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBaseSummaryTemplatesDeleteOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("base summary template id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[base-summary-templates]", "label,group-bys,filters,explicit-metrics,start-date,end-date,broker,created-by,created-at,updated-at")

	getBody, _, err := client.Get(cmd.Context(), "/v1/base-summary-templates/"+id, query)
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

	details := buildBaseSummaryTemplateDetails(resp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/base-summary-templates/"+id)
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

	if details.Label != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted base summary template %s (%s)\n", details.ID, details.Label)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted base summary template %s\n", details.ID)
	return nil
}

func parseDoBaseSummaryTemplatesDeleteOptions(cmd *cobra.Command) (doBaseSummaryTemplatesDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBaseSummaryTemplatesDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Confirm: confirm,
	}, nil
}
