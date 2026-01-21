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

type certificationTypesListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Broker             string
	Name               string
	CanApplyTo         string
	CanBeRequirementOf string
}

type certificationTypeRow struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	CanApplyTo         string `json:"can_apply_to,omitempty"`
	RequiresExpiration bool   `json:"requires_expiration"`
	CanBeRequirementOf string `json:"can_be_requirement_of,omitempty"`
	Broker             string `json:"broker,omitempty"`
	BrokerID           string `json:"broker_id,omitempty"`
}

func newCertificationTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certification types",
		Long: `List certification types with filtering and pagination.

Certification types define the types of certifications that can be tracked
for drivers, truckers, or equipment (e.g., CDL, HAZMAT, DOT medical).

Output Columns:
  ID           Certification type identifier
  NAME         Certification name
  APPLIES TO   What the certification applies to (driver, trucker, equipment)
  EXPIRES      Whether the certification requires an expiration date
  BROKER       Broker name

Filters:
  --broker               Filter by broker ID
  --name                 Filter by name (partial match, case-insensitive)
  --can-apply-to         Filter by what it applies to
  --can-be-requirement-of Filter by what it can be a requirement of`,
		Example: `  # List all certification types
  xbe view certification-types list

  # Filter by broker
  xbe view certification-types list --broker 123

  # Filter by what they apply to
  xbe view certification-types list --can-apply-to driver

  # Output as JSON
  xbe view certification-types list --json`,
		RunE: runCertificationTypesList,
	}
	initCertificationTypesListFlags(cmd)
	return cmd
}

func init() {
	certificationTypesCmd.AddCommand(newCertificationTypesListCmd())
}

func initCertificationTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("can-apply-to", "", "Filter by what it applies to")
	cmd.Flags().String("can-be-requirement-of", "", "Filter by what it can be a requirement of")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCertificationTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCertificationTypesListOptions(cmd)
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
	query.Set("fields[certification-types]", "name,can-apply-to,requires-expiration,can-be-requirement-of,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[can-apply-to]", opts.CanApplyTo)
	setFilterIfPresent(query, "filter[can-be-requirement-of]", opts.CanBeRequirementOf)

	body, _, err := client.Get(cmd.Context(), "/v1/certification-types", query)
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

	rows := buildCertificationTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCertificationTypesTable(cmd, rows)
}

func parseCertificationTypesListOptions(cmd *cobra.Command) (certificationTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	canApplyTo, _ := cmd.Flags().GetString("can-apply-to")
	canBeRequirementOf, _ := cmd.Flags().GetString("can-be-requirement-of")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return certificationTypesListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Broker:             broker,
		Name:               name,
		CanApplyTo:         canApplyTo,
		CanBeRequirementOf: canBeRequirementOf,
	}, nil
}

func buildCertificationTypeRows(resp jsonAPIResponse) []certificationTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]certificationTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := certificationTypeRow{
			ID:                 resource.ID,
			Name:               stringAttr(resource.Attributes, "name"),
			CanApplyTo:         stringAttr(resource.Attributes, "can-apply-to"),
			RequiresExpiration: boolAttr(resource.Attributes, "requires-expiration"),
			CanBeRequirementOf: strings.Join(stringSliceAttr(resource.Attributes, "can-be-requirement-of"), ", "),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCertificationTypesTable(cmd *cobra.Command, rows []certificationTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No certification types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tAPPLIES TO\tEXPIRES\tBROKER")
	for _, row := range rows {
		expires := "no"
		if row.RequiresExpiration {
			expires = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.CanApplyTo, 15),
			expires,
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
