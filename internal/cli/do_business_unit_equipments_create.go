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

type doBusinessUnitEquipmentsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	BusinessUnit string
	Equipment    string
}

func newDoBusinessUnitEquipmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a business unit equipment link",
		Long: `Create a business unit equipment link.

Required flags:
  --business-unit  Business unit ID (required)
  --equipment      Equipment ID (required)

Note: Equipment can only be linked once per business unit.`,
		Example: `  # Link equipment to a business unit
  xbe do business-unit-equipments create --business-unit 123 --equipment 456

  # Output as JSON
  xbe do business-unit-equipments create --business-unit 123 --equipment 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoBusinessUnitEquipmentsCreate,
	}
	initDoBusinessUnitEquipmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doBusinessUnitEquipmentsCmd.AddCommand(newDoBusinessUnitEquipmentsCreateCmd())
}

func initDoBusinessUnitEquipmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("business-unit", "", "Business unit ID (required)")
	cmd.Flags().String("equipment", "", "Equipment ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBusinessUnitEquipmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBusinessUnitEquipmentsCreateOptions(cmd)
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

	if opts.BusinessUnit == "" {
		err := fmt.Errorf("--business-unit is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Equipment == "" {
		err := fmt.Errorf("--equipment is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"business-unit": map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   opts.BusinessUnit,
			},
		},
		"equipment": map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "business-unit-equipments",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/business-unit-equipments", jsonBody)
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

	row := businessUnitEquipmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created business unit equipment %s\n", row.ID)
	return nil
}

func parseDoBusinessUnitEquipmentsCreateOptions(cmd *cobra.Command) (doBusinessUnitEquipmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	equipment, _ := cmd.Flags().GetString("equipment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBusinessUnitEquipmentsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		BusinessUnit: businessUnit,
		Equipment:    equipment,
	}, nil
}
