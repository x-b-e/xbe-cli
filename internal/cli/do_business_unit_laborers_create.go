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

type doBusinessUnitLaborersCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	BusinessUnit string
	Laborer      string
}

func newDoBusinessUnitLaborersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a business unit laborer link",
		Long: `Create a business unit laborer link.

Required flags:
  --business-unit  Business unit ID (required)
  --laborer        Laborer ID (required)`,
		Example: `  # Link a business unit to a laborer
  xbe do business-unit-laborers create \\
    --business-unit 123 \\
    --laborer 456

  # JSON output
  xbe do business-unit-laborers create \\
    --business-unit 123 \\
    --laborer 456 \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoBusinessUnitLaborersCreate,
	}
	initDoBusinessUnitLaborersCreateFlags(cmd)
	return cmd
}

func init() {
	doBusinessUnitLaborersCmd.AddCommand(newDoBusinessUnitLaborersCreateCmd())
}

func initDoBusinessUnitLaborersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("business-unit", "", "Business unit ID (required)")
	cmd.Flags().String("laborer", "", "Laborer ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBusinessUnitLaborersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBusinessUnitLaborersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.BusinessUnit) == "" {
		err := fmt.Errorf("--business-unit is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Laborer) == "" {
		err := fmt.Errorf("--laborer is required")
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
		"laborer": map[string]any{
			"data": map[string]any{
				"type": "laborers",
				"id":   opts.Laborer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "business-unit-laborers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/business-unit-laborers", jsonBody)
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

	row := buildBusinessUnitLaborerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created business unit laborer %s\n", row.ID)
	return nil
}

func parseDoBusinessUnitLaborersCreateOptions(cmd *cobra.Command) (doBusinessUnitLaborersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	laborer, _ := cmd.Flags().GetString("laborer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBusinessUnitLaborersCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		BusinessUnit: businessUnit,
		Laborer:      laborer,
	}, nil
}

func buildBusinessUnitLaborerRowFromSingle(resp jsonAPISingleResponse) businessUnitLaborerRow {
	row := businessUnitLaborerRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["business-unit"]; ok && rel.Data != nil {
		row.BusinessUnitID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["laborer"]; ok && rel.Data != nil {
		row.LaborerID = rel.Data.ID
	}

	return row
}
