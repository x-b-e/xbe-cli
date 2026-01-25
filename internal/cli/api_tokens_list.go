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

type apiTokensListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	User    string
}

func newApiTokensListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API tokens",
		Long: `List API tokens with filtering and pagination.

API tokens are scoped to a user and may be revoked or expire.

Output Columns:
  ID            Token identifier
  NAME          Token label (if set)
  USER          User ID
  EXPIRES AT    Expiration timestamp (if set)
  REVOKED AT    Revocation timestamp (if revoked)
  LAST USED AT  Last usage timestamp (if any)

Filters:
  --user    Filter by user ID`,
		Example: `  # List API tokens
  xbe view api-tokens list

  # Filter by user
  xbe view api-tokens list --user 123

  # Paginate results
  xbe view api-tokens list --limit 20 --offset 40

  # Output as JSON
  xbe view api-tokens list --json`,
		RunE: runApiTokensList,
	}
	initApiTokensListFlags(cmd)
	return cmd
}

func init() {
	apiTokensCmd.AddCommand(newApiTokensListCmd())
}

func initApiTokensListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runApiTokensList(cmd *cobra.Command, _ []string) error {
	opts, err := parseApiTokensListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[api-tokens]", "name,expires-at,revoked-at,last-used-at,user")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[user]", opts.User)

	body, _, err := client.Get(cmd.Context(), "/v1/api-tokens", query)
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

	rows := buildApiTokenRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderApiTokensTable(cmd, rows)
}

func parseApiTokensListOptions(cmd *cobra.Command) (apiTokensListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return apiTokensListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		User:    user,
	}, nil
}

type apiTokenRow struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	CreatedBy  string `json:"created_by_id,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	RevokedAt  string `json:"revoked_at,omitempty"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	Token      string `json:"token,omitempty"`
}

func buildApiTokenRows(resp jsonAPIResponse) []apiTokenRow {
	rows := make([]apiTokenRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := apiTokenRow{
			ID:         resource.ID,
			Name:       strings.TrimSpace(stringAttr(resource.Attributes, "name")),
			ExpiresAt:  formatDateTime(stringAttr(resource.Attributes, "expires-at")),
			RevokedAt:  formatDateTime(stringAttr(resource.Attributes, "revoked-at")),
			LastUsedAt: formatDateTime(stringAttr(resource.Attributes, "last-used-at")),
			Token:      stringAttr(resource.Attributes, "token"),
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedBy = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderApiTokensTable(cmd *cobra.Command, rows []apiTokenRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No API tokens found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tUSER\tEXPIRES AT\tREVOKED AT\tLAST USED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Name,
			row.UserID,
			row.ExpiresAt,
			row.RevokedAt,
			row.LastUsedAt,
		)
	}
	return writer.Flush()
}
