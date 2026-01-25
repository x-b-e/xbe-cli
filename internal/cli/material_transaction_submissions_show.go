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

type materialTransactionSubmissionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newMaterialTransactionSubmissionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction submission details",
		Long: `Show the full details of a material transaction submission.

Output Fields:
  ID
  Material Transaction ID
  Comment

Arguments:
  <id>    The submission ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a submission
  xbe view material-transaction-submissions show 123

  # Output as JSON
  xbe view material-transaction-submissions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionSubmissionsShow,
	}
	initMaterialTransactionSubmissionsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionSubmissionsCmd.AddCommand(newMaterialTransactionSubmissionsShowCmd())
}

func initMaterialTransactionSubmissionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionSubmissionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionSubmissionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("material transaction submission id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-submissions]", "comment,material-transaction")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-submissions/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildMaterialTransactionSubmissionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionSubmissionDetails(cmd, details)
}

func parseMaterialTransactionSubmissionsShowOptions(cmd *cobra.Command) (materialTransactionSubmissionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionSubmissionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderMaterialTransactionSubmissionDetails(cmd *cobra.Command, details materialTransactionSubmissionRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialTransactionID != "" {
		fmt.Fprintf(out, "Material Transaction ID: %s\n", details.MaterialTransactionID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
