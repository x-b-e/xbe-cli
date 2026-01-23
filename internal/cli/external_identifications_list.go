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

type externalIdentificationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

func newExternalIdentificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List external identifications",
		Long: `List external identifications with pagination.

External identifications link external ID values (e.g., license numbers, tax IDs)
to entities like truckers, brokers, and material sites.

Output Columns:
  ID           External identification identifier
  VALUE        The external identification value
  TYPE ID      External identification type ID
  IDENTIFIES   Entity type and ID being identified`,
		Example: `  # List all external identifications
  xbe view external-identifications list

  # Output as JSON
  xbe view external-identifications list --json

  # List with pagination
  xbe view external-identifications list --limit 10 --offset 20`,
		RunE: runExternalIdentificationsList,
	}
	initExternalIdentificationsListFlags(cmd)
	return cmd
}

func init() {
	externalIdentificationsCmd.AddCommand(newExternalIdentificationsListCmd())
}

func initExternalIdentificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runExternalIdentificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseExternalIdentificationsListOptions(cmd)
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
	query.Set("fields[external-identifications]", "value,skip-value-validation,external-identification-type,identifies")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/external-identifications", query)
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

	rows := buildExternalIdentificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderExternalIdentificationsTable(cmd, rows)
}

func parseExternalIdentificationsListOptions(cmd *cobra.Command) (externalIdentificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return externalIdentificationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

type externalIdentificationRow struct {
	ID                           string `json:"id"`
	Value                        string `json:"value,omitempty"`
	SkipValueValidation          bool   `json:"skip_value_validation,omitempty"`
	ExternalIdentificationTypeID string `json:"external_identification_type_id,omitempty"`
	IdentifiesType               string `json:"identifies_type,omitempty"`
	IdentifiesID                 string `json:"identifies_id,omitempty"`
}

func buildExternalIdentificationRows(resp jsonAPIResponse) []externalIdentificationRow {
	rows := make([]externalIdentificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := externalIdentificationRow{
			ID:                  resource.ID,
			Value:               stringAttr(resource.Attributes, "value"),
			SkipValueValidation: boolAttr(resource.Attributes, "skip-value-validation"),
		}

		if rel, ok := resource.Relationships["external-identification-type"]; ok && rel.Data != nil {
			row.ExternalIdentificationTypeID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["identifies"]; ok && rel.Data != nil {
			row.IdentifiesType = rel.Data.Type
			row.IdentifiesID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderExternalIdentificationsTable(cmd *cobra.Command, rows []externalIdentificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No external identifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tVALUE\tTYPE ID\tIDENTIFIES TYPE\tIDENTIFIES ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Value, 30),
			row.ExternalIdentificationTypeID,
			row.IdentifiesType,
			row.IdentifiesID,
		)
	}
	return writer.Flush()
}
