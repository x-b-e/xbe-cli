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

type doEquipmentMovementStopRequirementsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Stop        string
	Requirement string
	Kind        string
}

func newDoEquipmentMovementStopRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement stop requirement",
		Long: `Create an equipment movement stop requirement.

Required flags:
  --stop          Equipment movement stop ID (required)
  --requirement   Equipment movement requirement ID (required)

Optional flags:
  --kind          Requirement kind (origin or destination). If omitted, the API
                  attempts to derive it from the stop location and requirement.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a stop requirement
  xbe do equipment-movement-stop-requirements create \
    --stop 123 \
    --requirement 456

  # Create with explicit kind
  xbe do equipment-movement-stop-requirements create \
    --stop 123 \
    --requirement 456 \
    --kind origin`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementStopRequirementsCreate,
	}
	initDoEquipmentMovementStopRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementStopRequirementsCmd.AddCommand(newDoEquipmentMovementStopRequirementsCreateCmd())
}

func initDoEquipmentMovementStopRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("stop", "", "Equipment movement stop ID (required)")
	cmd.Flags().String("requirement", "", "Equipment movement requirement ID (required)")
	cmd.Flags().String("kind", "", "Requirement kind (origin/destination) (optional)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementStopRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementStopRequirementsCreateOptions(cmd)
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

	if opts.Stop == "" {
		err := fmt.Errorf("--stop is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Requirement == "" {
		err := fmt.Errorf("--requirement is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Kind != "" {
		attributes["kind"] = opts.Kind
	}

	relationships := map[string]any{
		"stop": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-stops",
				"id":   opts.Stop,
			},
		},
		"requirement": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-requirements",
				"id":   opts.Requirement,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-stop-requirements",
			"relationships": relationships,
		},
	}

	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-stop-requirements", jsonBody)
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

	row := buildEquipmentMovementStopRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement stop requirement %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementStopRequirementsCreateOptions(cmd *cobra.Command) (doEquipmentMovementStopRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	stop, _ := cmd.Flags().GetString("stop")
	requirement, _ := cmd.Flags().GetString("requirement")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementStopRequirementsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Stop:        stop,
		Requirement: requirement,
		Kind:        kind,
	}, nil
}

func buildEquipmentMovementStopRequirementRowFromSingle(resp jsonAPISingleResponse) equipmentMovementStopRequirementRow {
	attrs := resp.Data.Attributes
	row := equipmentMovementStopRequirementRow{
		ID:            resp.Data.ID,
		Kind:          stringAttr(attrs, "kind"),
		RequirementAt: formatDateTime(stringAttr(attrs, "requirement-at")),
	}

	if rel, ok := resp.Data.Relationships["stop"]; ok && rel.Data != nil {
		row.StopID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["requirement"]; ok && rel.Data != nil {
		row.RequirementID = rel.Data.ID
	}

	return row
}
