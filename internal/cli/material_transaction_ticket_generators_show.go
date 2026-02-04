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

type materialTransactionTicketGeneratorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newMaterialTransactionTicketGeneratorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction ticket generator details",
		Long: `Show the full details of a material transaction ticket generator.

Output Fields:
  ID
  Format Rule
  Organization Type
  Organization ID
  Broker ID

Arguments:
  <id>    The ticket generator ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a ticket generator
  xbe view material-transaction-ticket-generators show 123

  # Output as JSON
  xbe view material-transaction-ticket-generators show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionTicketGeneratorsShow,
	}
	initMaterialTransactionTicketGeneratorsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionTicketGeneratorsCmd.AddCommand(newMaterialTransactionTicketGeneratorsShowCmd())
}

func initMaterialTransactionTicketGeneratorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionTicketGeneratorsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialTransactionTicketGeneratorsShowOptions(cmd)
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
		return fmt.Errorf("material transaction ticket generator id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-ticket-generators]", "format-rule,organization,broker")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-ticket-generators/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildMaterialTransactionTicketGeneratorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionTicketGeneratorDetails(cmd, details)
}

func parseMaterialTransactionTicketGeneratorsShowOptions(cmd *cobra.Command) (materialTransactionTicketGeneratorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionTicketGeneratorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderMaterialTransactionTicketGeneratorDetails(cmd *cobra.Command, details materialTransactionTicketGeneratorRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FormatRule != "" {
		fmt.Fprintf(out, "Format Rule: %s\n", details.FormatRule)
	}
	if details.OrganizationType != "" {
		fmt.Fprintf(out, "Organization Type: %s\n", details.OrganizationType)
	}
	if details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization ID: %s\n", details.OrganizationID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}

	return nil
}
