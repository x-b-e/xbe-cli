package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type maintenancePartsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type partDetails struct {
	ID            string  `json:"id"`
	PartNumber    string  `json:"part_number,omitempty"`
	Name          string  `json:"name,omitempty"`
	Description   string  `json:"description,omitempty"`
	Manufacturer  string  `json:"manufacturer,omitempty"`
	UnitCost      float64 `json:"unit_cost,omitempty"`
	UnitOfMeasure string  `json:"unit_of_measure,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

func newMaintenancePartsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance part details",
		Long: `Show the full details of a maintenance requirement part.

Retrieves and displays comprehensive information about a part including
manufacturer, cost, and specifications.

Output Sections (table format):
  Core Info       ID, part number, name
  Manufacturer    Manufacturer information
  Cost            Unit cost and unit of measure
  Description     Full description

Arguments:
  <id>          The part ID (required). Find IDs using the list command.`,
		Example: `  # View a part by ID
  xbe view maintenance parts show 123

  # Get part as JSON
  xbe view maintenance parts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenancePartsShow,
	}
	initMaintenancePartsShowFlags(cmd)
	return cmd
}

func init() {
	maintenancePartsCmd.AddCommand(newMaintenancePartsShowCmd())
}

func initMaintenancePartsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenancePartsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenancePartsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("part id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-parts/"+id, query)
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

	details := buildPartDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPartDetails(cmd, details)
}

func parseMaintenancePartsShowOptions(cmd *cobra.Command) (maintenancePartsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenancePartsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPartDetails(resp jsonAPISingleResponse) partDetails {
	attrs := resp.Data.Attributes

	return partDetails{
		ID:            resp.Data.ID,
		PartNumber:    stringAttr(attrs, "part-number"),
		Name:          strings.TrimSpace(stringAttr(attrs, "name")),
		Description:   strings.TrimSpace(stringAttr(attrs, "description")),
		Manufacturer:  stringAttr(attrs, "manufacturer"),
		UnitCost:      float64Attr(attrs, "unit-cost"),
		UnitOfMeasure: stringAttr(attrs, "unit-of-measure"),
		CreatedAt:     formatDate(stringAttr(attrs, "created-at")),
		UpdatedAt:     formatDate(stringAttr(attrs, "updated-at")),
	}
}

func renderPartDetails(cmd *cobra.Command, d partDetails) error {
	out := cmd.OutOrStdout()

	// Core info
	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.PartNumber != "" {
		fmt.Fprintf(out, "Part Number: %s\n", d.PartNumber)
	}
	if d.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", d.Name)
	}
	if d.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", d.CreatedAt)
	}
	if d.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", d.UpdatedAt)
	}

	// Manufacturer
	if d.Manufacturer != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Manufacturer:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintf(out, "  %s\n", d.Manufacturer)
	}

	// Cost
	if d.UnitCost > 0 || d.UnitOfMeasure != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Pricing:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.UnitCost > 0 {
			fmt.Fprintf(out, "  Unit Cost: $%.2f\n", d.UnitCost)
		}
		if d.UnitOfMeasure != "" {
			fmt.Fprintf(out, "  Unit of Measure: %s\n", d.UnitOfMeasure)
		}
	}

	// Description
	if d.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, d.Description)
	}

	return nil
}
