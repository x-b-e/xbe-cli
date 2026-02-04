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

type equipmentMovementTripJobProductionPlansShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementTripJobProductionPlanDetails struct {
	ID                    string `json:"id"`
	EquipmentMovementTrip string `json:"equipment_movement_trip_id,omitempty"`
	JobProductionPlan     string `json:"job_production_plan_id,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newEquipmentMovementTripJobProductionPlansShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement trip job production plan details",
		Long: `Show the full details of an equipment movement trip job production plan link.

Output Fields:
  ID
  Equipment Movement Trip ID
  Job Production Plan ID
  Created At
  Updated At

Arguments:
  <id>    The link ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a link
  xbe view equipment-movement-trip-job-production-plans show 123

  # Output as JSON
  xbe view equipment-movement-trip-job-production-plans show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementTripJobProductionPlansShow,
	}
	initEquipmentMovementTripJobProductionPlansShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripJobProductionPlansCmd.AddCommand(newEquipmentMovementTripJobProductionPlansShowCmd())
}

func initEquipmentMovementTripJobProductionPlansShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripJobProductionPlansShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseEquipmentMovementTripJobProductionPlansShowOptions(cmd)
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
		return fmt.Errorf("equipment movement trip job production plan id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-movement-trip-job-production-plans]", "created-at,updated-at,equipment-movement-trip,job-production-plan")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trip-job-production-plans/"+id, query)
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

	details := buildEquipmentMovementTripJobProductionPlanDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementTripJobProductionPlanDetails(cmd, details)
}

func parseEquipmentMovementTripJobProductionPlansShowOptions(cmd *cobra.Command) (equipmentMovementTripJobProductionPlansShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripJobProductionPlansShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementTripJobProductionPlanDetails(resp jsonAPISingleResponse) equipmentMovementTripJobProductionPlanDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := equipmentMovementTripJobProductionPlanDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["equipment-movement-trip"]; ok && rel.Data != nil {
		details.EquipmentMovementTrip = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlan = rel.Data.ID
	}

	return details
}

func renderEquipmentMovementTripJobProductionPlanDetails(cmd *cobra.Command, details equipmentMovementTripJobProductionPlanDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EquipmentMovementTrip != "" {
		fmt.Fprintf(out, "Equipment Movement Trip ID: %s\n", details.EquipmentMovementTrip)
	}
	if details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlan)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
