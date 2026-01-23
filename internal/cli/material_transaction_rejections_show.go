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

type materialTransactionRejectionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newMaterialTransactionRejectionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction rejection details",
		Long: `Show the full details of a material transaction rejection.

Output Fields:
  ID
  Material Transaction ID
  Comment

Arguments:
  <id>    The rejection ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a rejection
  xbe view material-transaction-rejections show 123

  # Output as JSON
  xbe view material-transaction-rejections show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionRejectionsShow,
	}
	initMaterialTransactionRejectionsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionRejectionsCmd.AddCommand(newMaterialTransactionRejectionsShowCmd())
}

func initMaterialTransactionRejectionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionRejectionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionRejectionsShowOptions(cmd)
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
		return fmt.Errorf("material transaction rejection id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-rejections]", "comment,material-transaction")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-rejections/"+id, query)
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

	details := buildMaterialTransactionRejectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionRejectionDetails(cmd, details)
}

func parseMaterialTransactionRejectionsShowOptions(cmd *cobra.Command) (materialTransactionRejectionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionRejectionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderMaterialTransactionRejectionDetails(cmd *cobra.Command, details materialTransactionRejectionRow) error {
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
