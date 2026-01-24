package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doLehmanRobertsApexViewpointTicketExportsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	SaleDateMin   string
	SaleDateMax   string
	TemplateName  string
	LocationIDs   []string
	OmitHeaderRow bool
}

type lehmanRobertsApexViewpointTicketExportDetails struct {
	ID            string   `json:"id"`
	TemplateName  string   `json:"template_name,omitempty"`
	SaleDateMin   string   `json:"sale_date_min,omitempty"`
	SaleDateMax   string   `json:"sale_date_max,omitempty"`
	LocationIDs   []string `json:"location_ids,omitempty"`
	OmitHeaderRow bool     `json:"omit_header_row"`
	CSV           string   `json:"csv,omitempty"`
}

func newDoLehmanRobertsApexViewpointTicketExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Lehman Roberts Apex Viewpoint ticket export",
		Long: `Create a Lehman Roberts Apex Viewpoint ticket export.

Required:
  --template-name   Viewpoint template name (lrJWSCash, lrJWSCrdt, lrTicketsV)
  --sale-date-min   Earliest sale date (YYYY-MM-DD)
  --sale-date-max   Latest sale date (YYYY-MM-DD)

Optional:
  --location-ids    Location IDs to include (comma-separated or repeated)
  --omit-header-row Omit the header row in the CSV

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a ticket export
  xbe do lehman-roberts-apex-viewpoint-ticket-exports create \
    --template-name lrJWSCash \
    --sale-date-min 2025-01-01 \
    --sale-date-max 2025-01-31

  # Export without a header row
  xbe do lehman-roberts-apex-viewpoint-ticket-exports create \
    --template-name lrJWSCrdt \
    --sale-date-min 2025-01-01 \
    --sale-date-max 2025-01-31 \
    --omit-header-row`,
		Args: cobra.NoArgs,
		RunE: runDoLehmanRobertsApexViewpointTicketExportsCreate,
	}
	initDoLehmanRobertsApexViewpointTicketExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doLehmanRobertsApexViewpointTicketExportsCmd.AddCommand(newDoLehmanRobertsApexViewpointTicketExportsCreateCmd())
}

func initDoLehmanRobertsApexViewpointTicketExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("template-name", "", "Viewpoint template name (lrJWSCash, lrJWSCrdt, lrTicketsV)")
	cmd.Flags().String("sale-date-min", "", "Earliest sale date (YYYY-MM-DD)")
	cmd.Flags().String("sale-date-max", "", "Latest sale date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("location-ids", nil, "Location IDs to include (comma-separated or repeated)")
	cmd.Flags().Bool("omit-header-row", false, "Omit the header row in the CSV")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("template-name")
	_ = cmd.MarkFlagRequired("sale-date-min")
	_ = cmd.MarkFlagRequired("sale-date-max")
}

func runDoLehmanRobertsApexViewpointTicketExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLehmanRobertsApexViewpointTicketExportsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	saleDateMin := strings.TrimSpace(opts.SaleDateMin)
	saleDateMax := strings.TrimSpace(opts.SaleDateMax)
	templateName := strings.TrimSpace(opts.TemplateName)
	if templateName == "" {
		err := fmt.Errorf("--template-name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if saleDateMin == "" {
		err := fmt.Errorf("--sale-date-min is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if saleDateMax == "" {
		err := fmt.Errorf("--sale-date-max is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"template-name": templateName,
		"sale-date-min": saleDateMin,
		"sale-date-max": saleDateMax,
	}

	locationIDs := compactStringSlice(opts.LocationIDs)
	if cmd.Flags().Changed("location-ids") {
		attributes["location-ids"] = locationIDs
	}
	if cmd.Flags().Changed("omit-header-row") {
		attributes["omit-header-row"] = opts.OmitHeaderRow
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "lehman-roberts-apex-viewpoint-ticket-exports",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lehman-roberts-apex-viewpoint-ticket-exports", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildLehmanRobertsApexViewpointTicketExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLehmanRobertsApexViewpointTicketExportDetails(cmd, details)
}

func parseDoLehmanRobertsApexViewpointTicketExportsCreateOptions(cmd *cobra.Command) (doLehmanRobertsApexViewpointTicketExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	templateName, _ := cmd.Flags().GetString("template-name")
	saleDateMin, _ := cmd.Flags().GetString("sale-date-min")
	saleDateMax, _ := cmd.Flags().GetString("sale-date-max")
	locationIDs, _ := cmd.Flags().GetStringSlice("location-ids")
	omitHeaderRow, _ := cmd.Flags().GetBool("omit-header-row")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLehmanRobertsApexViewpointTicketExportsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		SaleDateMin:   saleDateMin,
		SaleDateMax:   saleDateMax,
		TemplateName:  templateName,
		LocationIDs:   locationIDs,
		OmitHeaderRow: omitHeaderRow,
	}, nil
}

func buildLehmanRobertsApexViewpointTicketExportDetails(resp jsonAPISingleResponse) lehmanRobertsApexViewpointTicketExportDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return lehmanRobertsApexViewpointTicketExportDetails{
		ID:            resource.ID,
		TemplateName:  stringAttr(attrs, "template-name"),
		SaleDateMin:   formatDate(stringAttr(attrs, "sale-date-min")),
		SaleDateMax:   formatDate(stringAttr(attrs, "sale-date-max")),
		LocationIDs:   stringSliceAttr(attrs, "location-ids"),
		OmitHeaderRow: boolAttr(attrs, "omit-header-row"),
		CSV:           stringAttr(attrs, "csv"),
	}
}

func renderLehmanRobertsApexViewpointTicketExportDetails(cmd *cobra.Command, details lehmanRobertsApexViewpointTicketExportDetails) error {
	out := cmd.OutOrStdout()

	if details.ID != "" {
		fmt.Fprintf(out, "ID: %s\n", details.ID)
	}
	if details.TemplateName != "" {
		fmt.Fprintf(out, "Template Name: %s\n", details.TemplateName)
	}
	if details.SaleDateMin != "" {
		fmt.Fprintf(out, "Sale Date Min: %s\n", details.SaleDateMin)
	}
	if details.SaleDateMax != "" {
		fmt.Fprintf(out, "Sale Date Max: %s\n", details.SaleDateMax)
	}
	if len(details.LocationIDs) > 0 {
		fmt.Fprintf(out, "Location IDs: %s\n", strings.Join(details.LocationIDs, ", "))
	}
	fmt.Fprintf(out, "Omit Header Row: %t\n", details.OmitHeaderRow)

	if details.CSV != "" {
		fmt.Fprintln(out, "\nCSV:")
		fmt.Fprintln(out, details.CSV)
		return nil
	}

	if details.ID != "" {
		fmt.Fprintf(out, "Created Lehman Roberts Apex Viewpoint ticket export %s\n", details.ID)
		return nil
	}

	fmt.Fprintln(out, "Created Lehman Roberts Apex Viewpoint ticket export")
	return nil
}
