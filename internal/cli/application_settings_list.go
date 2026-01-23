package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type applicationSettingsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

type applicationSettingRow struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

func newApplicationSettingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List application settings",
		Long: `List application settings.

Application settings are global key/value pairs used to configure platform behavior.
Access is restricted to admin users.

Output Columns:
  ID           Setting identifier
  KEY          Setting key
  VALUE        Setting value
  DESCRIPTION  Setting description (if present)

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  None`,
		Example: `  # List application settings
  xbe view application-settings list

  # Paginate results
  xbe view application-settings list --limit 25 --offset 50

  # Output as JSON
  xbe view application-settings list --json`,
		RunE: runApplicationSettingsList,
	}
	initApplicationSettingsListFlags(cmd)
	return cmd
}

func init() {
	applicationSettingsCmd.AddCommand(newApplicationSettingsListCmd())
}

func initApplicationSettingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runApplicationSettingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseApplicationSettingsListOptions(cmd)
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
	query.Set("sort", "key")
	query.Set("fields[application-settings]", "key,value,description")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/application-settings", query)
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

	rows := buildApplicationSettingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderApplicationSettingsTable(cmd, rows)
}

func parseApplicationSettingsListOptions(cmd *cobra.Command) (applicationSettingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return applicationSettingsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func buildApplicationSettingRows(resp jsonAPIResponse) []applicationSettingRow {
	rows := make([]applicationSettingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, applicationSettingRow{
			ID:          resource.ID,
			Key:         stringAttr(resource.Attributes, "key"),
			Value:       stringAttr(resource.Attributes, "value"),
			Description: stringAttr(resource.Attributes, "description"),
		})
	}
	return rows
}

func buildApplicationSettingRowFromSingle(resp jsonAPISingleResponse) applicationSettingRow {
	attrs := resp.Data.Attributes
	return applicationSettingRow{
		ID:          resp.Data.ID,
		Key:         stringAttr(attrs, "key"),
		Value:       stringAttr(attrs, "value"),
		Description: stringAttr(attrs, "description"),
	}
}

func renderApplicationSettingsTable(cmd *cobra.Command, rows []applicationSettingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No application settings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tKEY\tVALUE\tDESCRIPTION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Key, 30),
			truncateString(row.Value, 40),
			truncateString(row.Description, 40),
		)
	}
	return writer.Flush()
}
