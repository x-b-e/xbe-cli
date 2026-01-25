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

type equipmentMovementStopRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementStopRequirementDetails struct {
	ID            string `json:"id"`
	StopID        string `json:"stop_id,omitempty"`
	RequirementID string `json:"requirement_id,omitempty"`
	Kind          string `json:"kind,omitempty"`
	RequirementAt string `json:"requirement_at,omitempty"`
}

func newEquipmentMovementStopRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement stop requirement details",
		Long: `Show the full details of an equipment movement stop requirement.

Output Fields:
  ID
  Stop ID
  Requirement ID
  Kind
  Requirement At

Arguments:
  <id>    The stop requirement ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a stop requirement
  xbe view equipment-movement-stop-requirements show 123

  # Get JSON output
  xbe view equipment-movement-stop-requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementStopRequirementsShow,
	}
	initEquipmentMovementStopRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementStopRequirementsCmd.AddCommand(newEquipmentMovementStopRequirementsShowCmd())
}

func initEquipmentMovementStopRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementStopRequirementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentMovementStopRequirementsShowOptions(cmd)
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
		return fmt.Errorf("equipment movement stop requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-movement-stop-requirements]", "kind,requirement-at,stop,requirement")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-stop-requirements/"+id, query)
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

	details := buildEquipmentMovementStopRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementStopRequirementDetails(cmd, details)
}

func parseEquipmentMovementStopRequirementsShowOptions(cmd *cobra.Command) (equipmentMovementStopRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementStopRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementStopRequirementDetails(resp jsonAPISingleResponse) equipmentMovementStopRequirementDetails {
	attrs := resp.Data.Attributes
	details := equipmentMovementStopRequirementDetails{
		ID:            resp.Data.ID,
		Kind:          stringAttr(attrs, "kind"),
		RequirementAt: formatDateTime(stringAttr(attrs, "requirement-at")),
	}

	if rel, ok := resp.Data.Relationships["stop"]; ok && rel.Data != nil {
		details.StopID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["requirement"]; ok && rel.Data != nil {
		details.RequirementID = rel.Data.ID
	}

	return details
}

func renderEquipmentMovementStopRequirementDetails(cmd *cobra.Command, details equipmentMovementStopRequirementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.StopID != "" {
		fmt.Fprintf(out, "Stop ID: %s\n", details.StopID)
	}
	if details.RequirementID != "" {
		fmt.Fprintf(out, "Requirement ID: %s\n", details.RequirementID)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.RequirementAt != "" {
		fmt.Fprintf(out, "Requirement At: %s\n", details.RequirementAt)
	}

	return nil
}
