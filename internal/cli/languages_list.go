package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type languagesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Code    string
}

func newLanguagesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List languages",
		Long: `List available languages.

Languages represent available language options that can be associated
with users and content.

Output Columns:
  ID    Language identifier
  NAME  Language name (e.g., English, Spanish)
  CODE  ISO language code (e.g., en, es)

Filters:
  --code  Filter by language code`,
		Example: `  # List all languages
  xbe view languages list

  # Filter by code
  xbe view languages list --code en

  # Output as JSON
  xbe view languages list --json`,
		RunE: runLanguagesList,
	}
	initLanguagesListFlags(cmd)
	return cmd
}

func init() {
	languagesCmd.AddCommand(newLanguagesListCmd())
}

func initLanguagesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("code", "", "Filter by language code")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLanguagesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLanguagesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
	query.Set("sort", "name")
	query.Set("fields[languages]", "name,code")

	setFilterIfPresent(query, "filter[code]", opts.Code)

	body, _, err := client.Get(cmd.Context(), "/v1/languages", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildLanguageRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLanguagesTable(cmd, rows)
}

func parseLanguagesListOptions(cmd *cobra.Command) (languagesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	code, _ := cmd.Flags().GetString("code")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return languagesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Code:    code,
	}, nil
}

type languageRow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func buildLanguageRows(resp jsonAPIResponse) []languageRow {
	rows := make([]languageRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := languageRow{
			ID:   resource.ID,
			Name: stringAttr(resource.Attributes, "name"),
			Code: stringAttr(resource.Attributes, "code"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderLanguagesTable(cmd *cobra.Command, rows []languageRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No languages found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tCODE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.Name,
			row.Code,
		)
	}
	return writer.Flush()
}
