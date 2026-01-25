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

type developerReferenceTypesListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Developer   string
	SubjectType string
	Broker      string
}

func newDeveloperReferenceTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developer reference types",
		Long: `List developer reference types with filtering and pagination.

Developer reference types define custom reference fields for developers.

Output Columns:
  ID            Reference type identifier
  NAME          Reference type name
  SUBJECT TYPES Entity types this reference applies to
  DEVELOPER     Developer ID

Filters:
  --developer    Filter by developer ID
  --subject-type Filter by subject type
  --broker       Filter by broker ID`,
		Example: `  # List all developer reference types
  xbe view developer-reference-types list

  # Filter by developer
  xbe view developer-reference-types list --developer 123

  # Filter by broker
  xbe view developer-reference-types list --broker 456

  # Output as JSON
  xbe view developer-reference-types list --json`,
		RunE: runDeveloperReferenceTypesList,
	}
	initDeveloperReferenceTypesListFlags(cmd)
	return cmd
}

func init() {
	developerReferenceTypesCmd.AddCommand(newDeveloperReferenceTypesListCmd())
}

func initDeveloperReferenceTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("subject-type", "", "Filter by subject type")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperReferenceTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeveloperReferenceTypesListOptions(cmd)
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
	query.Set("fields[developer-reference-types]", "name,subject-types,developer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[subject-type]", opts.SubjectType)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/developer-reference-types", query)
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

	rows := buildDeveloperReferenceTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeveloperReferenceTypesTable(cmd, rows)
}

func parseDeveloperReferenceTypesListOptions(cmd *cobra.Command) (developerReferenceTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	developer, _ := cmd.Flags().GetString("developer")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerReferenceTypesListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Developer:   developer,
		SubjectType: subjectType,
		Broker:      broker,
	}, nil
}

type developerReferenceTypeRow struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	SubjectTypes []string `json:"subject_types,omitempty"`
	DeveloperID  string   `json:"developer_id,omitempty"`
}

func buildDeveloperReferenceTypeRows(resp jsonAPIResponse) []developerReferenceTypeRow {
	rows := make([]developerReferenceTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := developerReferenceTypeRow{
			ID:   resource.ID,
			Name: stringAttr(resource.Attributes, "name"),
		}

		if st, ok := resource.Attributes["subject-types"].([]any); ok {
			for _, s := range st {
				if str, ok := s.(string); ok {
					row.SubjectTypes = append(row.SubjectTypes, str)
				}
			}
		}

		if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
			row.DeveloperID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderDeveloperReferenceTypesTable(cmd *cobra.Command, rows []developerReferenceTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developer reference types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSUBJECT TYPES\tDEVELOPER")
	for _, row := range rows {
		subjectTypes := strings.Join(row.SubjectTypes, ", ")
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(subjectTypes, 30),
			row.DeveloperID,
		)
	}
	return writer.Flush()
}
