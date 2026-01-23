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

type developerReferencesListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	DeveloperReferenceType string
	Developer              string
	SubjectType            string
	SubjectID              string
}

type developerReferenceRow struct {
	ID                       string `json:"id"`
	Value                    string `json:"value,omitempty"`
	DeveloperReferenceTypeID string `json:"developer_reference_type_id,omitempty"`
	DeveloperID              string `json:"developer_id,omitempty"`
	SubjectType              string `json:"subject_type,omitempty"`
	SubjectID                string `json:"subject_id,omitempty"`
}

func newDeveloperReferencesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developer references",
		Long: `List developer references.

Output Columns:
  ID              Developer reference identifier
  VALUE           Reference value
  TYPE            Developer reference type ID
  DEVELOPER       Developer ID
  SUBJECT         Subject type and ID

Filters:
  --developer-reference-type  Filter by developer reference type ID
  --developer                 Filter by developer ID
  --subject-type              Filter by subject type
  --subject-id                Filter by subject ID`,
		Example: `  # List all developer references
  xbe view developer-references list

  # Filter by developer
  xbe view developer-references list --developer 123

  # Filter by subject
  xbe view developer-references list --subject-type projects --subject-id 456

  # Output as JSON
  xbe view developer-references list --json`,
		RunE: runDeveloperReferencesList,
	}
	initDeveloperReferencesListFlags(cmd)
	return cmd
}

func init() {
	developerReferencesCmd.AddCommand(newDeveloperReferencesListCmd())
}

func initDeveloperReferencesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("developer-reference-type", "", "Filter by developer reference type ID")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("subject-type", "", "Filter by subject type")
	cmd.Flags().String("subject-id", "", "Filter by subject ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperReferencesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeveloperReferencesListOptions(cmd)
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
	query.Set("include", "developer-reference-type,developer,subject")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[developer_reference_type]", opts.DeveloperReferenceType)
	setFilterIfPresent(query, "filter[developer]", opts.Developer)

	// Handle polymorphic subject filter
	if opts.SubjectType != "" && opts.SubjectID != "" {
		query.Set("filter[subject]", opts.SubjectType+"|"+opts.SubjectID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/developer-references", query)
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

	rows := buildDeveloperReferenceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeveloperReferencesTable(cmd, rows)
}

func parseDeveloperReferencesListOptions(cmd *cobra.Command) (developerReferencesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	developerReferenceType, _ := cmd.Flags().GetString("developer-reference-type")
	developer, _ := cmd.Flags().GetString("developer")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerReferencesListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		DeveloperReferenceType: developerReferenceType,
		Developer:              developer,
		SubjectType:            subjectType,
		SubjectID:              subjectID,
	}, nil
}

func buildDeveloperReferenceRows(resp jsonAPIResponse) []developerReferenceRow {
	rows := make([]developerReferenceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := developerReferenceRow{
			ID:    resource.ID,
			Value: stringAttr(resource.Attributes, "value"),
		}

		if rel, ok := resource.Relationships["developer-reference-type"]; ok && rel.Data != nil {
			row.DeveloperReferenceTypeID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
			row.DeveloperID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderDeveloperReferencesTable(cmd *cobra.Command, rows []developerReferenceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developer references found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tVALUE\tTYPE\tDEVELOPER\tSUBJECT")
	for _, row := range rows {
		subject := ""
		if row.SubjectType != "" && row.SubjectID != "" {
			subject = row.SubjectType + "/" + row.SubjectID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Value, 30),
			row.DeveloperReferenceTypeID,
			row.DeveloperID,
			truncateString(subject, 30),
		)
	}
	return writer.Flush()
}
