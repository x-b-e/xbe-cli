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

type transportOrderStopsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type transportOrderStopDetails struct {
	ID                          string   `json:"id"`
	TransportOrderID            string   `json:"transport_order_id,omitempty"`
	LocationID                  string   `json:"location_id,omitempty"`
	Role                        string   `json:"role,omitempty"`
	Status                      string   `json:"status,omitempty"`
	Position                    int      `json:"position,omitempty"`
	AtMin                       string   `json:"at_min,omitempty"`
	AtMax                       string   `json:"at_max,omitempty"`
	ExternalTmsStopNumber       string   `json:"external_tms_stop_number,omitempty"`
	TransportOrderStopMaterials []string `json:"transport_order_stop_material_ids,omitempty"`
	ProjectTransportPlanStops   []string `json:"project_transport_plan_stop_ids,omitempty"`
	TransportReferenceIDs       []string `json:"transport_reference_ids,omitempty"`
}

func newTransportOrderStopsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show transport order stop details",
		Long: `Show the full details of a transport order stop.

Output Fields:
  ID            Stop identifier
  Transport Order Transport order ID
  Location      Transport location ID
  Role          Stop role (pickup, delivery)
  Status        Stop status
  Position      Stop position
  At Min        Earliest scheduled time
  At Max        Latest scheduled time
  External TMS Stop Number External TMS stop number (if set)
  Stop Materials Transport order stop material IDs
  Plan Stops    Project transport plan stop IDs
  References    Transport reference IDs

Arguments:
  <id>    The transport order stop ID (required). You can find IDs using the list command.`,
		Example: `  # Show a transport order stop
  xbe view transport-order-stops show 123

  # Get JSON output
  xbe view transport-order-stops show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTransportOrderStopsShow,
	}
	initTransportOrderStopsShowFlags(cmd)
	return cmd
}

func init() {
	transportOrderStopsCmd.AddCommand(newTransportOrderStopsShowCmd())
}

func initTransportOrderStopsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderStopsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTransportOrderStopsShowOptions(cmd)
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
		return fmt.Errorf("transport order stop id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[transport-order-stops]", "position,role,status,at-min,at-max,external-tms-stop-number,transport-order,location,transport-order-stop-materials,project-transport-plan-stops,transport-references")

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-stops/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTransportOrderStopDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTransportOrderStopDetails(cmd, details)
}

func parseTransportOrderStopsShowOptions(cmd *cobra.Command) (transportOrderStopsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportOrderStopsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTransportOrderStopDetails(resp jsonAPISingleResponse) transportOrderStopDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := transportOrderStopDetails{
		ID:                    resource.ID,
		Role:                  stringAttr(attrs, "role"),
		Status:                stringAttr(attrs, "status"),
		Position:              intAttr(attrs, "position"),
		AtMin:                 formatDateTime(stringAttr(attrs, "at-min")),
		AtMax:                 formatDateTime(stringAttr(attrs, "at-max")),
		ExternalTmsStopNumber: stringAttr(attrs, "external-tms-stop-number"),
	}

	if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
		details.TransportOrderID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["location"]; ok && rel.Data != nil {
		details.LocationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["transport-order-stop-materials"]; ok && rel.raw != nil {
		details.TransportOrderStopMaterials = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["project-transport-plan-stops"]; ok && rel.raw != nil {
		details.ProjectTransportPlanStops = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["transport-references"]; ok && rel.raw != nil {
		details.TransportReferenceIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderTransportOrderStopDetails(cmd *cobra.Command, details transportOrderStopDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TransportOrderID != "" {
		fmt.Fprintf(out, "Transport Order: %s\n", details.TransportOrderID)
	}
	if details.LocationID != "" {
		fmt.Fprintf(out, "Location: %s\n", details.LocationID)
	}
	if details.Role != "" {
		fmt.Fprintf(out, "Role: %s\n", details.Role)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Position != 0 {
		fmt.Fprintf(out, "Position: %d\n", details.Position)
	}
	if details.AtMin != "" {
		fmt.Fprintf(out, "At Min: %s\n", details.AtMin)
	}
	if details.AtMax != "" {
		fmt.Fprintf(out, "At Max: %s\n", details.AtMax)
	}
	if details.ExternalTmsStopNumber != "" {
		fmt.Fprintf(out, "External TMS Stop Number: %s\n", details.ExternalTmsStopNumber)
	}
	if len(details.TransportOrderStopMaterials) > 0 {
		fmt.Fprintf(out, "Stop Materials: %s\n", strings.Join(details.TransportOrderStopMaterials, ", "))
	}
	if len(details.ProjectTransportPlanStops) > 0 {
		fmt.Fprintf(out, "Plan Stops: %s\n", strings.Join(details.ProjectTransportPlanStops, ", "))
	}
	if len(details.TransportReferenceIDs) > 0 {
		fmt.Fprintf(out, "References: %s\n", strings.Join(details.TransportReferenceIDs, ", "))
	}

	return nil
}
