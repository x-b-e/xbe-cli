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

type userSearchesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

func newUserSearchesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user searches",
		Long: `List user searches with pagination.

User searches look up users by contact method and value. Results are returned
from the server when available.

Output Columns:
  ID                Search identifier
  CONTACT METHOD    Contact method used for the search
  CONTACT VALUE     Contact value used for the search
  ONLY ADMIN/MEMBER Restrict matches to admins or members
  MATCHING USER     Matched user name/email or ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List user searches
  xbe view user-searches list

  # Output as JSON
  xbe view user-searches list --json`,
		Args: cobra.NoArgs,
		RunE: runUserSearchesList,
	}
	initUserSearchesListFlags(cmd)
	return cmd
}

func init() {
	userSearchesCmd.AddCommand(newUserSearchesListCmd())
}

func initUserSearchesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserSearchesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserSearchesListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-searches]", "contact-method,contact-value,only-admin-or-member,matching-user")
	query.Set("include", "matching-user")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/user-searches", query)
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

	rows := buildUserSearchRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserSearchesTable(cmd, rows)
}

func parseUserSearchesListOptions(cmd *cobra.Command) (userSearchesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userSearchesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func renderUserSearchesTable(cmd *cobra.Command, rows []userSearchRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user searches found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCONTACT METHOD\tCONTACT VALUE\tONLY ADMIN/MEMBER\tMATCHING USER")
	for _, row := range rows {
		matchingLabel := userSearchMatchingLabel(row)
		contactValue := truncateString(userSearchContactValueLabel(row), 32)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ContactMethod,
			contactValue,
			userSearchOnlyAdminLabel(row),
			truncateString(matchingLabel, 30),
		)
	}
	return writer.Flush()
}
