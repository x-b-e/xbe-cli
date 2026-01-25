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

type crewRatesListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	ResourceType               string
	ResourceID                 string
	ResourceClassificationType string
	ResourceClassificationID   string
	StartOn                    string
	StartOnMin                 string
	StartOnMax                 string
	EndOn                      string
	EndOnMin                   string
	EndOnMax                   string
	IsActive                   string
	Broker                     string
	CraftClass                 string
	Search                     string
}

type crewRateRow struct {
	ID                         string `json:"id"`
	Description                string `json:"description,omitempty"`
	PricePerUnit               string `json:"price_per_unit,omitempty"`
	StartOn                    string `json:"start_on,omitempty"`
	EndOn                      string `json:"end_on,omitempty"`
	IsActive                   bool   `json:"is_active,omitempty"`
	BrokerID                   string `json:"broker_id,omitempty"`
	ResourceClassificationType string `json:"resource_classification_type,omitempty"`
	ResourceClassificationID   string `json:"resource_classification_id,omitempty"`
	ResourceType               string `json:"resource_type,omitempty"`
	ResourceID                 string `json:"resource_id,omitempty"`
	CraftClassID               string `json:"craft_class_id,omitempty"`
}

func newCrewRatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List crew rates",
		Long: `List crew rates.

Output Columns:
  ID           Crew rate identifier
  DESCRIPTION  Description
  PRICE        Price per unit
  START        Start date
  END          End date
  ACTIVE       Active status
  BROKER       Broker ID
  SCOPE        Resource classification/resource/craft class

Filters:
  --resource-type                 Filter by resource type (Laborer, Equipment)
  --resource-id                   Filter by resource ID
  --resource-classification-type  Filter by resource classification type (LaborClassification, EquipmentClassification)
  --resource-classification-id    Filter by resource classification ID
  --start-on                      Filter by start date
  --start-on-min                  Filter by minimum start date
  --start-on-max                  Filter by maximum start date
  --end-on                        Filter by end date
  --end-on-min                    Filter by minimum end date
  --end-on-max                    Filter by maximum end date
  --is-active                     Filter by active status (true/false)
  --broker                        Filter by broker ID
  --craft-class                   Filter by craft class ID
  --search                        Search crew rates by description`,
		Example: `  # List crew rates
  xbe view crew-rates list

  # Filter by broker and active status
  xbe view crew-rates list --broker 123 --is-active true

  # Filter by resource classification
  xbe view crew-rates list --resource-classification-type LaborClassification --resource-classification-id 456

  # Filter by resource
  xbe view crew-rates list --resource-type Equipment --resource-id 789

  # Filter by date range
  xbe view crew-rates list --start-on-min 2025-01-01 --end-on-max 2025-12-31

  # Search by description
  xbe view crew-rates list --search "night shift"

  # Output as JSON
  xbe view crew-rates list --json`,
		RunE: runCrewRatesList,
	}
	initCrewRatesListFlags(cmd)
	return cmd
}

func init() {
	crewRatesCmd.AddCommand(newCrewRatesListCmd())
}

func initCrewRatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("resource-type", "", "Filter by resource type (Laborer, Equipment)")
	cmd.Flags().String("resource-id", "", "Filter by resource ID")
	cmd.Flags().String("resource-classification-type", "", "Filter by resource classification type (LaborClassification, EquipmentClassification)")
	cmd.Flags().String("resource-classification-id", "", "Filter by resource classification ID")
	cmd.Flags().String("start-on", "", "Filter by start date")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date")
	cmd.Flags().String("end-on", "", "Filter by end date")
	cmd.Flags().String("end-on-min", "", "Filter by minimum end date")
	cmd.Flags().String("end-on-max", "", "Filter by maximum end date")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("craft-class", "", "Filter by craft class ID")
	cmd.Flags().String("search", "", "Search crew rates by description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCrewRatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCrewRatesListOptions(cmd)
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
	query.Set("fields[crew-rates]", "description,price-per-unit,start-on,end-on,is-active,broker,resource,resource-classification,craft-class")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	if opts.ResourceType != "" && opts.ResourceID != "" {
		query.Set("filter[resource]", opts.ResourceType+"|"+opts.ResourceID)
	}
	if opts.ResourceClassificationType != "" && opts.ResourceClassificationID != "" {
		query.Set("filter[resource_classification]", opts.ResourceClassificationType+"|"+opts.ResourceClassificationID)
	}

	setFilterIfPresent(query, "filter[start_on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start_on_min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start_on_max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[end_on]", opts.EndOn)
	setFilterIfPresent(query, "filter[end_on_min]", opts.EndOnMin)
	setFilterIfPresent(query, "filter[end_on_max]", opts.EndOnMax)
	setFilterIfPresent(query, "filter[is_active]", opts.IsActive)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[craft_class]", opts.CraftClass)
	setFilterIfPresent(query, "filter[q]", opts.Search)

	body, _, err := client.Get(cmd.Context(), "/v1/crew-rates", query)
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

	rows := buildCrewRateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCrewRatesTable(cmd, rows)
}

func parseCrewRatesListOptions(cmd *cobra.Command) (crewRatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	resourceClassificationType, _ := cmd.Flags().GetString("resource-classification-type")
	resourceClassificationID, _ := cmd.Flags().GetString("resource-classification-id")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	endOn, _ := cmd.Flags().GetString("end-on")
	endOnMin, _ := cmd.Flags().GetString("end-on-min")
	endOnMax, _ := cmd.Flags().GetString("end-on-max")
	isActive, _ := cmd.Flags().GetString("is-active")
	broker, _ := cmd.Flags().GetString("broker")
	craftClass, _ := cmd.Flags().GetString("craft-class")
	search, _ := cmd.Flags().GetString("search")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return crewRatesListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		ResourceType:               resourceType,
		ResourceID:                 resourceID,
		ResourceClassificationType: resourceClassificationType,
		ResourceClassificationID:   resourceClassificationID,
		StartOn:                    startOn,
		StartOnMin:                 startOnMin,
		StartOnMax:                 startOnMax,
		EndOn:                      endOn,
		EndOnMin:                   endOnMin,
		EndOnMax:                   endOnMax,
		IsActive:                   isActive,
		Broker:                     broker,
		CraftClass:                 craftClass,
		Search:                     search,
	}, nil
}

func buildCrewRateRows(resp jsonAPIResponse) []crewRateRow {
	rows := make([]crewRateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCrewRateRow(resource))
	}
	return rows
}

func buildCrewRateRow(resource jsonAPIResource) crewRateRow {
	row := crewRateRow{
		ID:           resource.ID,
		Description:  strings.TrimSpace(stringAttr(resource.Attributes, "description")),
		PricePerUnit: stringAttr(resource.Attributes, "price-per-unit"),
		StartOn:      formatDate(stringAttr(resource.Attributes, "start-on")),
		EndOn:        formatDate(stringAttr(resource.Attributes, "end-on")),
		IsActive:     boolAttr(resource.Attributes, "is-active"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource-classification"]; ok && rel.Data != nil {
		row.ResourceClassificationType = rel.Data.Type
		row.ResourceClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		row.ResourceType = rel.Data.Type
		row.ResourceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["craft-class"]; ok && rel.Data != nil {
		row.CraftClassID = rel.Data.ID
	}

	return row
}

func renderCrewRatesTable(cmd *cobra.Command, rows []crewRateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No crew rates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDESCRIPTION\tPRICE\tSTART\tEND\tACTIVE\tBROKER\tSCOPE")
	for _, row := range rows {
		scopeParts := make([]string, 0, 3)
		if row.ResourceClassificationType != "" && row.ResourceClassificationID != "" {
			scopeParts = append(scopeParts, row.ResourceClassificationType+"/"+row.ResourceClassificationID)
		}
		if row.ResourceType != "" && row.ResourceID != "" {
			scopeParts = append(scopeParts, row.ResourceType+"/"+row.ResourceID)
		}
		if row.CraftClassID != "" {
			scopeParts = append(scopeParts, "craft-classes/"+row.CraftClassID)
		}
		scope := strings.Join(scopeParts, ", ")

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
			row.ID,
			truncateString(row.Description, 32),
			row.PricePerUnit,
			row.StartOn,
			row.EndOn,
			row.IsActive,
			row.BrokerID,
			truncateString(scope, 40),
		)
	}
	return writer.Flush()
}
