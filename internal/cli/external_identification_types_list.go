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

type externalIdentificationTypesListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Name       string
	CanApplyTo string
}

func newExternalIdentificationTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List external identification types",
		Long: `List external identification types with filtering and pagination.

External identification types define the kinds of external IDs that can be
associated with entities (e.g., license numbers for truckers, tax IDs for brokers).

Output Columns:
  ID           External identification type identifier
  NAME         Type name
  CAN APPLY TO Entity types this can apply to
  REGEX        Format validation regex (if defined)
  UNIQUE       Whether values must be globally unique

Filters:
  --name          Filter by name (partial match, case-insensitive)
  --can-apply-to  Filter by entity type (e.g., Trucker, Broker)`,
		Example: `  # List all external identification types
  xbe view external-identification-types list

  # Filter by name
  xbe view external-identification-types list --name "license"

  # Filter by what they can apply to
  xbe view external-identification-types list --can-apply-to Trucker

  # Output as JSON
  xbe view external-identification-types list --json`,
		RunE: runExternalIdentificationTypesList,
	}
	initExternalIdentificationTypesListFlags(cmd)
	return cmd
}

func init() {
	externalIdentificationTypesCmd.AddCommand(newExternalIdentificationTypesListCmd())
}

func initExternalIdentificationTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("can-apply-to", "", "Filter by entity type (e.g., Trucker)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runExternalIdentificationTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseExternalIdentificationTypesListOptions(cmd)
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
	query.Set("fields[external-identification-types]", "name,can-apply-to,format-validation-regex,value-should-be-globally-unique")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[can-apply-to]", opts.CanApplyTo)

	body, _, err := client.Get(cmd.Context(), "/v1/external-identification-types", query)
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

	rows := buildExternalIdentificationTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderExternalIdentificationTypesTable(cmd, rows)
}

func parseExternalIdentificationTypesListOptions(cmd *cobra.Command) (externalIdentificationTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	canApplyTo, _ := cmd.Flags().GetString("can-apply-to")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return externalIdentificationTypesListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Name:       name,
		CanApplyTo: canApplyTo,
	}, nil
}

func buildExternalIdentificationTypeRows(resp jsonAPIResponse) []externalIdentificationTypeRow {
	rows := make([]externalIdentificationTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := externalIdentificationTypeRow{
			ID:                          resource.ID,
			Name:                        stringAttr(resource.Attributes, "name"),
			CanApplyTo:                  stringSliceAttr(resource.Attributes, "can-apply-to"),
			FormatValidationRegex:       stringAttr(resource.Attributes, "format-validation-regex"),
			ValueShouldBeGloballyUnique: boolAttr(resource.Attributes, "value-should-be-globally-unique"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderExternalIdentificationTypesTable(cmd *cobra.Command, rows []externalIdentificationTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No external identification types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tCAN APPLY TO\tREGEX\tUNIQUE")
	for _, row := range rows {
		canApplyTo := strings.Join(row.CanApplyTo, ", ")
		unique := "no"
		if row.ValueShouldBeGloballyUnique {
			unique = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(canApplyTo, 25),
			truncateString(row.FormatValidationRegex, 20),
			unique,
		)
	}
	return writer.Flush()
}
