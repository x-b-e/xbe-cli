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

type truckScopesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

func newTruckScopesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List truck scopes",
		Long: `List truck scopes with pagination.

Truck scopes define geographic and equipment restrictions for trucking operations.

Output Columns:
  ID                        Scope identifier
  TRAILER CLASSIFICATIONS   Allowed trailer classification IDs
  AUTHORIZED STATES         Authorized state codes
  ADDRESS                   Address location
  PROXIMITY                 Address proximity in meters
  ORG TYPE                  Organization type
  ORG ID                    Organization ID`,
		Example: `  # List all truck scopes
  xbe view truck-scopes list

  # Output as JSON
  xbe view truck-scopes list --json`,
		RunE: runTruckScopesList,
	}
	initTruckScopesListFlags(cmd)
	return cmd
}

func init() {
	truckScopesCmd.AddCommand(newTruckScopesListCmd())
}

func initTruckScopesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckScopesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckScopesListOptions(cmd)
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
	query.Set("fields[truck-scopes]", "trailer-classification-ids,authorized-state-codes,address,address-city,address-state-code,address-proximity-meters,organization")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/truck-scopes", query)
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

	rows := buildTruckScopeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckScopesTable(cmd, rows)
}

func parseTruckScopesListOptions(cmd *cobra.Command) (truckScopesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckScopesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

type truckScopeRow struct {
	ID                       string   `json:"id"`
	TrailerClassificationIDs []string `json:"trailer_classification_ids,omitempty"`
	AuthorizedStateCodes     []string `json:"authorized_state_codes,omitempty"`
	Address                  string   `json:"address,omitempty"`
	AddressCity              string   `json:"address_city,omitempty"`
	AddressStateCode         string   `json:"address_state_code,omitempty"`
	AddressProximityMeters   any      `json:"address_proximity_meters,omitempty"`
	OrganizationType         string   `json:"organization_type,omitempty"`
	OrganizationID           string   `json:"organization_id,omitempty"`
}

func buildTruckScopeRows(resp jsonAPIResponse) []truckScopeRow {
	rows := make([]truckScopeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := truckScopeRow{
			ID:                     resource.ID,
			Address:                stringAttr(resource.Attributes, "address"),
			AddressCity:            stringAttr(resource.Attributes, "address-city"),
			AddressStateCode:       stringAttr(resource.Attributes, "address-state-code"),
			AddressProximityMeters: resource.Attributes["address-proximity-meters"],
		}

		if ids, ok := resource.Attributes["trailer-classification-ids"].([]any); ok {
			for _, id := range ids {
				if s, ok := id.(string); ok {
					row.TrailerClassificationIDs = append(row.TrailerClassificationIDs, s)
				} else if n, ok := id.(float64); ok {
					row.TrailerClassificationIDs = append(row.TrailerClassificationIDs, fmt.Sprintf("%.0f", n))
				}
			}
		}

		if codes, ok := resource.Attributes["authorized-state-codes"].([]any); ok {
			for _, code := range codes {
				if s, ok := code.(string); ok {
					row.AuthorizedStateCodes = append(row.AuthorizedStateCodes, s)
				}
			}
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTruckScopesTable(cmd *cobra.Command, rows []truckScopeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No truck scopes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRAILER CLASS IDS\tAUTHORIZED STATES\tADDRESS\tORG TYPE\tORG ID")
	for _, row := range rows {
		trailerIDs := strings.Join(row.TrailerClassificationIDs, ",")
		stateCodes := strings.Join(row.AuthorizedStateCodes, ",")
		address := ""
		if row.AddressCity != "" || row.AddressStateCode != "" {
			address = fmt.Sprintf("%s, %s", row.AddressCity, row.AddressStateCode)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(trailerIDs, 20),
			truncateString(stateCodes, 20),
			truncateString(address, 20),
			row.OrganizationType,
			row.OrganizationID,
		)
	}
	return writer.Flush()
}
