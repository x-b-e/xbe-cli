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

type certificationRequirementsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	CertificationType string
	RequiredBy        string
}

func newCertificationRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certification requirements",
		Long: `List certification requirements with filtering and pagination.

Certification requirements define which certifications are needed by entities.

Output Columns:
  ID                  Requirement identifier
  REQUIRED BY TYPE    Type of requiring entity
  REQUIRED BY ID      ID of requiring entity
  CERT TYPE           Certification type ID
  PERIOD START        Period start date
  PERIOD END          Period end date

Filters:
  --certification-type    Filter by certification type ID
  --required-by           Filter by requiring entity (format: Type|ID, e.g., Project|123)`,
		Example: `  # List all certification requirements
  xbe view certification-requirements list

  # Filter by certification type
  xbe view certification-requirements list --certification-type 123

  # Filter by requiring entity
  xbe view certification-requirements list --required-by "Project|456"

  # Output as JSON
  xbe view certification-requirements list --json`,
		RunE: runCertificationRequirementsList,
	}
	initCertificationRequirementsListFlags(cmd)
	return cmd
}

func init() {
	certificationRequirementsCmd.AddCommand(newCertificationRequirementsListCmd())
}

func initCertificationRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("certification-type", "", "Filter by certification type ID")
	cmd.Flags().String("required-by", "", "Filter by requiring entity (Type|ID, e.g., Project|123)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCertificationRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCertificationRequirementsListOptions(cmd)
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
	query.Set("fields[certification-requirements]", "period-start,period-end,certification-type,required-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[certification_type]", opts.CertificationType)
	setFilterIfPresent(query, "filter[required_by]", opts.RequiredBy)

	body, _, err := client.Get(cmd.Context(), "/v1/certification-requirements", query)
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

	rows := buildCertificationRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCertificationRequirementsTable(cmd, rows)
}

func parseCertificationRequirementsListOptions(cmd *cobra.Command) (certificationRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	certificationType, _ := cmd.Flags().GetString("certification-type")
	requiredBy, _ := cmd.Flags().GetString("required-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return certificationRequirementsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		CertificationType: certificationType,
		RequiredBy:        requiredBy,
	}, nil
}

type certificationRequirementRow struct {
	ID                  string `json:"id"`
	RequiredByType      string `json:"required_by_type,omitempty"`
	RequiredByID        string `json:"required_by_id,omitempty"`
	CertificationTypeID string `json:"certification_type_id,omitempty"`
	PeriodStart         string `json:"period_start,omitempty"`
	PeriodEnd           string `json:"period_end,omitempty"`
}

func buildCertificationRequirementRows(resp jsonAPIResponse) []certificationRequirementRow {
	rows := make([]certificationRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := certificationRequirementRow{
			ID:          resource.ID,
			PeriodStart: stringAttr(resource.Attributes, "period-start"),
			PeriodEnd:   stringAttr(resource.Attributes, "period-end"),
		}

		if rel, ok := resource.Relationships["required-by"]; ok && rel.Data != nil {
			row.RequiredByType = rel.Data.Type
			row.RequiredByID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["certification-type"]; ok && rel.Data != nil {
			row.CertificationTypeID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCertificationRequirementsTable(cmd *cobra.Command, rows []certificationRequirementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No certification requirements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREQUIRED BY TYPE\tREQUIRED BY ID\tCERT TYPE\tPERIOD START\tPERIOD END")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.RequiredByType,
			row.RequiredByID,
			row.CertificationTypeID,
			row.PeriodStart,
			row.PeriodEnd,
		)
	}
	return writer.Flush()
}
