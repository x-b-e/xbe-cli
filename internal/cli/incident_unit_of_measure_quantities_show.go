package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type incidentUnitOfMeasureQuantitiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentUnitOfMeasureQuantityDetails struct {
	ID                 string `json:"id"`
	Quantity           string `json:"quantity,omitempty"`
	IsSetAutomatically bool   `json:"is_set_automatically"`
	UnitOfMeasureID    string `json:"unit_of_measure_id,omitempty"`
	IncidentType       string `json:"incident_type,omitempty"`
	IncidentID         string `json:"incident_id,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

func newIncidentUnitOfMeasureQuantitiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident unit of measure quantity details",
		Long: `Show full details of an incident unit of measure quantity.

Output Fields:
  ID               Quantity identifier
  Quantity         Quantity value
  Auto             Whether the quantity was set automatically
  Unit of Measure  Unit of measure ID
  Incident         Incident type and ID
  Created At       Creation timestamp
  Updated At       Last update timestamp

Arguments:
  <id>    Incident unit of measure quantity ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an incident unit of measure quantity
  xbe view incident-unit-of-measure-quantities show 123

  # JSON output
  xbe view incident-unit-of-measure-quantities show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentUnitOfMeasureQuantitiesShow,
	}
	initIncidentUnitOfMeasureQuantitiesShowFlags(cmd)
	return cmd
}

func init() {
	incidentUnitOfMeasureQuantitiesCmd.AddCommand(newIncidentUnitOfMeasureQuantitiesShowCmd())
}

func initIncidentUnitOfMeasureQuantitiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentUnitOfMeasureQuantitiesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseIncidentUnitOfMeasureQuantitiesShowOptions(cmd)
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
		return fmt.Errorf("incident unit of measure quantity id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-unit-of-measure-quantities]", "quantity,is-set-automatically,created-at,updated-at,unit-of-measure,incident")

	body, status, err := client.Get(cmd.Context(), "/v1/incident-unit-of-measure-quantities/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderIncidentUnitOfMeasureQuantitiesShowUnavailable(cmd, opts.JSON)
		}
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildIncidentUnitOfMeasureQuantityDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIncidentUnitOfMeasureQuantityDetails(cmd, details)
}

func parseIncidentUnitOfMeasureQuantitiesShowOptions(cmd *cobra.Command) (incidentUnitOfMeasureQuantitiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentUnitOfMeasureQuantitiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentUnitOfMeasureQuantityDetails(resp jsonAPISingleResponse) incidentUnitOfMeasureQuantityDetails {
	attrs := resp.Data.Attributes
	details := incidentUnitOfMeasureQuantityDetails{
		ID:                 resp.Data.ID,
		Quantity:           stringAttr(attrs, "quantity"),
		IsSetAutomatically: boolAttr(attrs, "is-set-automatically"),
		CreatedAt:          formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:          formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["incident"]; ok && rel.Data != nil {
		details.IncidentType = rel.Data.Type
		details.IncidentID = rel.Data.ID
	}

	return details
}

func renderIncidentUnitOfMeasureQuantitiesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), incidentUnitOfMeasureQuantityDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Incident unit of measure quantity not found.")
	return nil
}

func renderIncidentUnitOfMeasureQuantityDetails(cmd *cobra.Command, details incidentUnitOfMeasureQuantityDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Quantity: %s\n", formatOptional(details.Quantity))
	if details.IsSetAutomatically {
		fmt.Fprintln(out, "Auto: yes")
	} else {
		fmt.Fprintln(out, "Auto: no")
	}
	fmt.Fprintf(out, "Unit of Measure: %s\n", formatOptional(details.UnitOfMeasureID))
	fmt.Fprintf(out, "Incident: %s\n", formatOptional(formatIncidentReference(details.IncidentType, details.IncidentID)))
	fmt.Fprintf(out, "Created At: %s\n", formatOptional(details.CreatedAt))
	fmt.Fprintf(out, "Updated At: %s\n", formatOptional(details.UpdatedAt))
	return nil
}
