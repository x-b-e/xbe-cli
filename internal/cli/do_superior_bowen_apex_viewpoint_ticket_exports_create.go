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

type doSuperiorBowenApexViewpointTicketExportsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	SaleDateMin string
	SaleDateMax string
	LocationIDs []string
}

type superiorBowenApexViewpointTicketExportDetails struct {
	ID          string   `json:"id"`
	SaleDateMin string   `json:"sale_date_min,omitempty"`
	SaleDateMax string   `json:"sale_date_max,omitempty"`
	LocationIDs []string `json:"location_ids,omitempty"`
	CSV         string   `json:"csv,omitempty"`
}

func newDoSuperiorBowenApexViewpointTicketExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Superior Bowen Apex Viewpoint ticket export",
		Long: `Create a Superior Bowen Apex Viewpoint ticket export.

Required:
  --sale-date-min   Earliest sale date (YYYY-MM-DD)
  --sale-date-max   Latest sale date (YYYY-MM-DD)

Optional:
  --location-ids    Location IDs to include (comma-separated or repeated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a ticket export
  xbe do superior-bowen-apex-viewpoint-ticket-exports create \
    --sale-date-min 2025-01-01 \
    --sale-date-max 2025-01-31

  # Create an export for specific locations
  xbe do superior-bowen-apex-viewpoint-ticket-exports create \
    --sale-date-min 2025-01-01 \
    --sale-date-max 2025-01-31 \
    --location-ids 001,004`,
		Args: cobra.NoArgs,
		RunE: runDoSuperiorBowenApexViewpointTicketExportsCreate,
	}
	initDoSuperiorBowenApexViewpointTicketExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doSuperiorBowenApexViewpointTicketExportsCmd.AddCommand(newDoSuperiorBowenApexViewpointTicketExportsCreateCmd())
}

func initDoSuperiorBowenApexViewpointTicketExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("sale-date-min", "", "Earliest sale date (YYYY-MM-DD)")
	cmd.Flags().String("sale-date-max", "", "Latest sale date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("location-ids", nil, "Location IDs to include (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("sale-date-min")
	_ = cmd.MarkFlagRequired("sale-date-max")
}

func runDoSuperiorBowenApexViewpointTicketExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoSuperiorBowenApexViewpointTicketExportsCreateOptions(cmd)
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
		"sale-date-min": saleDateMin,
		"sale-date-max": saleDateMax,
	}

	locationIDs := compactStringSlice(opts.LocationIDs)
	if cmd.Flags().Changed("location-ids") {
		attributes["location-ids"] = locationIDs
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "superior-bowen-apex-viewpoint-ticket-exports",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/superior-bowen-apex-viewpoint-ticket-exports", jsonBody)
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

	details := buildSuperiorBowenApexViewpointTicketExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderSuperiorBowenApexViewpointTicketExportDetails(cmd, details)
}

func parseDoSuperiorBowenApexViewpointTicketExportsCreateOptions(cmd *cobra.Command) (doSuperiorBowenApexViewpointTicketExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	saleDateMin, _ := cmd.Flags().GetString("sale-date-min")
	saleDateMax, _ := cmd.Flags().GetString("sale-date-max")
	locationIDs, _ := cmd.Flags().GetStringSlice("location-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSuperiorBowenApexViewpointTicketExportsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		SaleDateMin: saleDateMin,
		SaleDateMax: saleDateMax,
		LocationIDs: locationIDs,
	}, nil
}

func buildSuperiorBowenApexViewpointTicketExportDetails(resp jsonAPISingleResponse) superiorBowenApexViewpointTicketExportDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return superiorBowenApexViewpointTicketExportDetails{
		ID:          resource.ID,
		SaleDateMin: formatDate(stringAttr(attrs, "sale-date-min")),
		SaleDateMax: formatDate(stringAttr(attrs, "sale-date-max")),
		LocationIDs: stringSliceAttr(attrs, "location-ids"),
		CSV:         stringAttr(attrs, "csv"),
	}
}

func renderSuperiorBowenApexViewpointTicketExportDetails(cmd *cobra.Command, details superiorBowenApexViewpointTicketExportDetails) error {
	out := cmd.OutOrStdout()

	if details.ID != "" {
		fmt.Fprintf(out, "ID: %s\n", details.ID)
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

	if details.CSV != "" {
		fmt.Fprintln(out, "\nCSV:")
		fmt.Fprintln(out, details.CSV)
		return nil
	}

	if details.ID != "" {
		fmt.Fprintf(out, "Created Superior Bowen Apex Viewpoint ticket export %s\n", details.ID)
		return nil
	}

	fmt.Fprintln(out, "Created Superior Bowen Apex Viewpoint ticket export")
	return nil
}
