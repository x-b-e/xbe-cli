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

type tractorCredentialsListOptions struct {
	BaseURL                                string
	Token                                  string
	JSON                                   bool
	NoAuth                                 bool
	Limit                                  int
	Offset                                 int
	Tractor                                string
	TractorTrailerCredentialClassification string
	IssuedOnMin                            string
	IssuedOnMax                            string
	ExpiresOnMin                           string
	ExpiresOnMax                           string
	ActiveOn                               string
}

func newTractorCredentialsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tractor credentials",
		Long: `List tractor credentials with filtering and pagination.

Tractor credentials are documents or certifications assigned to tractors.

Output Columns:
  ID                     Credential identifier
  TRACTOR                Tractor ID
  CLASSIFICATION         Classification ID
  ISSUED ON              Issue date
  EXPIRES ON             Expiration date

Filters:
  --tractor                                    Filter by tractor ID
  --tractor-trailer-credential-classification  Filter by classification ID
  --issued-on-min                              Filter by minimum issue date
  --issued-on-max                              Filter by maximum issue date
  --expires-on-min                             Filter by minimum expiration date
  --expires-on-max                             Filter by maximum expiration date
  --active-on                                  Filter by active on date`,
		Example: `  # List all tractor credentials
  xbe view tractor-credentials list

  # Filter by tractor
  xbe view tractor-credentials list --tractor 123

  # Filter by classification
  xbe view tractor-credentials list --tractor-trailer-credential-classification 456

  # Filter by active on date
  xbe view tractor-credentials list --active-on 2024-01-15

  # Output as JSON
  xbe view tractor-credentials list --json`,
		RunE: runTractorCredentialsList,
	}
	initTractorCredentialsListFlags(cmd)
	return cmd
}

func init() {
	tractorCredentialsCmd.AddCommand(newTractorCredentialsListCmd())
}

func initTractorCredentialsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("tractor-trailer-credential-classification", "", "Filter by classification ID")
	cmd.Flags().String("issued-on-min", "", "Filter by minimum issue date (YYYY-MM-DD)")
	cmd.Flags().String("issued-on-max", "", "Filter by maximum issue date (YYYY-MM-DD)")
	cmd.Flags().String("expires-on-min", "", "Filter by minimum expiration date (YYYY-MM-DD)")
	cmd.Flags().String("expires-on-max", "", "Filter by maximum expiration date (YYYY-MM-DD)")
	cmd.Flags().String("active-on", "", "Filter by active on date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTractorCredentialsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTractorCredentialsListOptions(cmd)
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
	query.Set("fields[tractor-credentials]", "issued-on,expires-on,tractor,tractor-trailer-credential-classification")
	query.Set("include", "tractor,tractor-trailer-credential-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[tractor_trailer_credential_classification]", opts.TractorTrailerCredentialClassification)
	setFilterIfPresent(query, "filter[issued_on][min]", opts.IssuedOnMin)
	setFilterIfPresent(query, "filter[issued_on][max]", opts.IssuedOnMax)
	setFilterIfPresent(query, "filter[expires_on][min]", opts.ExpiresOnMin)
	setFilterIfPresent(query, "filter[expires_on][max]", opts.ExpiresOnMax)
	setFilterIfPresent(query, "filter[active_on]", opts.ActiveOn)

	body, _, err := client.Get(cmd.Context(), "/v1/tractor-credentials", query)
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

	rows := buildTractorCredentialRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTractorCredentialsTable(cmd, rows)
}

func parseTractorCredentialsListOptions(cmd *cobra.Command) (tractorCredentialsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	tractor, _ := cmd.Flags().GetString("tractor")
	classification, _ := cmd.Flags().GetString("tractor-trailer-credential-classification")
	issuedOnMin, _ := cmd.Flags().GetString("issued-on-min")
	issuedOnMax, _ := cmd.Flags().GetString("issued-on-max")
	expiresOnMin, _ := cmd.Flags().GetString("expires-on-min")
	expiresOnMax, _ := cmd.Flags().GetString("expires-on-max")
	activeOn, _ := cmd.Flags().GetString("active-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tractorCredentialsListOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		NoAuth:                                 noAuth,
		Limit:                                  limit,
		Offset:                                 offset,
		Tractor:                                tractor,
		TractorTrailerCredentialClassification: classification,
		IssuedOnMin:                            issuedOnMin,
		IssuedOnMax:                            issuedOnMax,
		ExpiresOnMin:                           expiresOnMin,
		ExpiresOnMax:                           expiresOnMax,
		ActiveOn:                               activeOn,
	}, nil
}

type tractorCredentialRow struct {
	ID                                       string `json:"id"`
	TractorID                                string `json:"tractor_id,omitempty"`
	TractorTrailerCredentialClassificationID string `json:"tractor_trailer_credential_classification_id,omitempty"`
	IssuedOn                                 string `json:"issued_on,omitempty"`
	ExpiresOn                                string `json:"expires_on,omitempty"`
}

func buildTractorCredentialRows(resp jsonAPIResponse) []tractorCredentialRow {
	rows := make([]tractorCredentialRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tractorCredentialRow{
			ID:        resource.ID,
			IssuedOn:  stringAttr(resource.Attributes, "issued-on"),
			ExpiresOn: stringAttr(resource.Attributes, "expires-on"),
		}

		if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
			row.TractorID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["tractor-trailer-credential-classification"]; ok && rel.Data != nil {
			row.TractorTrailerCredentialClassificationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTractorCredentialsTable(cmd *cobra.Command, rows []tractorCredentialRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tractor credentials found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRACTOR\tCLASSIFICATION\tISSUED ON\tEXPIRES ON")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TractorID,
			row.TractorTrailerCredentialClassificationID,
			row.IssuedOn,
			row.ExpiresOn,
		)
	}
	return writer.Flush()
}
