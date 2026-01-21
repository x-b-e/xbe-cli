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

type certificationsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	CertificationType string
	ByCertifies       string
	Status            string
	ExpiresWithinDays string
	ExpiresBefore     string
	Broker            string
}

func newCertificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certifications",
		Long: `List certifications with filtering and pagination.

Certifications are assigned to entities (users, truckers, etc.) based on certification types.

Output Columns:
  ID                  Certification identifier
  CERTIFIES TYPE      Type of entity certified
  CERTIFIES ID        ID of entity certified
  CERT TYPE           Certification type ID
  STATUS              Certification status
  EFFECTIVE           Effective date
  EXPIRES             Expiration date

Filters:
  --certification-type    Filter by certification type ID
  --by-certifies          Filter by certified entity (format: Type|ID, e.g., Trucker|123)
  --status                Filter by status
  --expires-within-days   Filter by expiration within N days
  --expires-before        Filter by expiration before date
  --broker                Filter by broker ID`,
		Example: `  # List all certifications
  xbe view certifications list

  # Filter by certification type
  xbe view certifications list --certification-type 123

  # Filter by certified entity
  xbe view certifications list --by-certifies "Trucker|456"

  # Filter by status
  xbe view certifications list --status active

  # Filter by expiration
  xbe view certifications list --expires-within-days 30

  # Output as JSON
  xbe view certifications list --json`,
		RunE: runCertificationsList,
	}
	initCertificationsListFlags(cmd)
	return cmd
}

func init() {
	certificationsCmd.AddCommand(newCertificationsListCmd())
}

func initCertificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("certification-type", "", "Filter by certification type ID")
	cmd.Flags().String("by-certifies", "", "Filter by certified entity (Type|ID, e.g., Trucker|123)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("expires-within-days", "", "Filter by expiration within N days")
	cmd.Flags().String("expires-before", "", "Filter by expiration before date (YYYY-MM-DD)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCertificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCertificationsListOptions(cmd)
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
	query.Set("fields[certifications]", "expires-at,effective-at,status,certification-type,certifies")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[certification_type]", opts.CertificationType)
	setFilterIfPresent(query, "filter[by_certifies]", opts.ByCertifies)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[expires_within_days]", opts.ExpiresWithinDays)
	setFilterIfPresent(query, "filter[expires_before]", opts.ExpiresBefore)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/certifications", query)
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

	rows := buildCertificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCertificationsTable(cmd, rows)
}

func parseCertificationsListOptions(cmd *cobra.Command) (certificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	certificationType, _ := cmd.Flags().GetString("certification-type")
	byCertifies, _ := cmd.Flags().GetString("by-certifies")
	status, _ := cmd.Flags().GetString("status")
	expiresWithinDays, _ := cmd.Flags().GetString("expires-within-days")
	expiresBefore, _ := cmd.Flags().GetString("expires-before")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return certificationsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		CertificationType: certificationType,
		ByCertifies:       byCertifies,
		Status:            status,
		ExpiresWithinDays: expiresWithinDays,
		ExpiresBefore:     expiresBefore,
		Broker:            broker,
	}, nil
}

type certificationRow struct {
	ID                  string `json:"id"`
	CertifiesType       string `json:"certifies_type,omitempty"`
	CertifiesID         string `json:"certifies_id,omitempty"`
	CertificationTypeID string `json:"certification_type_id,omitempty"`
	Status              string `json:"status,omitempty"`
	EffectiveAt         string `json:"effective_at,omitempty"`
	ExpiresAt           string `json:"expires_at,omitempty"`
}

func buildCertificationRows(resp jsonAPIResponse) []certificationRow {
	rows := make([]certificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := certificationRow{
			ID:          resource.ID,
			Status:      stringAttr(resource.Attributes, "status"),
			EffectiveAt: stringAttr(resource.Attributes, "effective-at"),
			ExpiresAt:   stringAttr(resource.Attributes, "expires-at"),
		}

		if rel, ok := resource.Relationships["certifies"]; ok && rel.Data != nil {
			row.CertifiesType = rel.Data.Type
			row.CertifiesID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["certification-type"]; ok && rel.Data != nil {
			row.CertificationTypeID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCertificationsTable(cmd *cobra.Command, rows []certificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No certifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCERTIFIES TYPE\tCERTIFIES ID\tCERT TYPE\tSTATUS\tEFFECTIVE\tEXPIRES")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.CertifiesType,
			row.CertifiesID,
			row.CertificationTypeID,
			row.Status,
			truncateString(row.EffectiveAt, 10),
			truncateString(row.ExpiresAt, 10),
		)
	}
	return writer.Flush()
}
