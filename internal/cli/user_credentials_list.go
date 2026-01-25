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

type userCredentialsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	User                         string
	UserCredentialClassification string
	IssuedOnMin                  string
	IssuedOnMax                  string
	ExpiresOnMin                 string
	ExpiresOnMax                 string
	ActiveOn                     string
}

func newUserCredentialsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user credentials",
		Long: `List user credentials with filtering and pagination.

User credentials are documents or certifications assigned to users.

Output Columns:
  ID                     Credential identifier
  USER                   User ID
  CLASSIFICATION         Classification ID
  ISSUED ON              Issue date
  EXPIRES ON             Expiration date

Filters:
  --user                              Filter by user ID
  --user-credential-classification    Filter by classification ID
  --issued-on-min                     Filter by minimum issue date
  --issued-on-max                     Filter by maximum issue date
  --expires-on-min                    Filter by minimum expiration date
  --expires-on-max                    Filter by maximum expiration date
  --active-on                         Filter by active on date`,
		Example: `  # List all user credentials
  xbe view user-credentials list

  # Filter by user
  xbe view user-credentials list --user 123

  # Filter by classification
  xbe view user-credentials list --user-credential-classification 456

  # Filter by active on date
  xbe view user-credentials list --active-on 2024-01-15

  # Output as JSON
  xbe view user-credentials list --json`,
		RunE: runUserCredentialsList,
	}
	initUserCredentialsListFlags(cmd)
	return cmd
}

func init() {
	userCredentialsCmd.AddCommand(newUserCredentialsListCmd())
}

func initUserCredentialsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("user-credential-classification", "", "Filter by classification ID")
	cmd.Flags().String("issued-on-min", "", "Filter by minimum issue date (YYYY-MM-DD)")
	cmd.Flags().String("issued-on-max", "", "Filter by maximum issue date (YYYY-MM-DD)")
	cmd.Flags().String("expires-on-min", "", "Filter by minimum expiration date (YYYY-MM-DD)")
	cmd.Flags().String("expires-on-max", "", "Filter by maximum expiration date (YYYY-MM-DD)")
	cmd.Flags().String("active-on", "", "Filter by active on date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserCredentialsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserCredentialsListOptions(cmd)
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
	query.Set("fields[user-credentials]", "issued-on,expires-on,user,user-credential-classification")
	query.Set("include", "user,user-credential-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[user_credential_classification]", opts.UserCredentialClassification)
	setFilterIfPresent(query, "filter[issued_on][min]", opts.IssuedOnMin)
	setFilterIfPresent(query, "filter[issued_on][max]", opts.IssuedOnMax)
	setFilterIfPresent(query, "filter[expires_on][min]", opts.ExpiresOnMin)
	setFilterIfPresent(query, "filter[expires_on][max]", opts.ExpiresOnMax)
	setFilterIfPresent(query, "filter[active_on]", opts.ActiveOn)

	body, _, err := client.Get(cmd.Context(), "/v1/user-credentials", query)
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

	rows := buildUserCredentialRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserCredentialsTable(cmd, rows)
}

func parseUserCredentialsListOptions(cmd *cobra.Command) (userCredentialsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	user, _ := cmd.Flags().GetString("user")
	classification, _ := cmd.Flags().GetString("user-credential-classification")
	issuedOnMin, _ := cmd.Flags().GetString("issued-on-min")
	issuedOnMax, _ := cmd.Flags().GetString("issued-on-max")
	expiresOnMin, _ := cmd.Flags().GetString("expires-on-min")
	expiresOnMax, _ := cmd.Flags().GetString("expires-on-max")
	activeOn, _ := cmd.Flags().GetString("active-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userCredentialsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		User:                         user,
		UserCredentialClassification: classification,
		IssuedOnMin:                  issuedOnMin,
		IssuedOnMax:                  issuedOnMax,
		ExpiresOnMin:                 expiresOnMin,
		ExpiresOnMax:                 expiresOnMax,
		ActiveOn:                     activeOn,
	}, nil
}

type userCredentialRow struct {
	ID                             string `json:"id"`
	UserID                         string `json:"user_id,omitempty"`
	UserCredentialClassificationID string `json:"user_credential_classification_id,omitempty"`
	IssuedOn                       string `json:"issued_on,omitempty"`
	ExpiresOn                      string `json:"expires_on,omitempty"`
}

func buildUserCredentialRows(resp jsonAPIResponse) []userCredentialRow {
	rows := make([]userCredentialRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := userCredentialRow{
			ID:        resource.ID,
			IssuedOn:  stringAttr(resource.Attributes, "issued-on"),
			ExpiresOn: stringAttr(resource.Attributes, "expires-on"),
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["user-credential-classification"]; ok && rel.Data != nil {
			row.UserCredentialClassificationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderUserCredentialsTable(cmd *cobra.Command, rows []userCredentialRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user credentials found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tCLASSIFICATION\tISSUED ON\tEXPIRES ON")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.UserID,
			row.UserCredentialClassificationID,
			row.IssuedOn,
			row.ExpiresOn,
		)
	}
	return writer.Flush()
}
