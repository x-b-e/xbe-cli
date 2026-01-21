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

type doCertificationRequirementsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	PeriodStart string
	PeriodEnd   string
}

func newDoCertificationRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a certification requirement",
		Long: `Update a certification requirement.

Optional flags:
  --period-start    Period start date (YYYY-MM-DD)
  --period-end      Period end date (YYYY-MM-DD)`,
		Example: `  # Update period start
  xbe do certification-requirements update 123 --period-start 2024-01-15

  # Update period end
  xbe do certification-requirements update 123 --period-end 2025-06-30`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCertificationRequirementsUpdate,
	}
	initDoCertificationRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCertificationRequirementsCmd.AddCommand(newDoCertificationRequirementsUpdateCmd())
}

func initDoCertificationRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("period-start", "", "Period start date (YYYY-MM-DD)")
	cmd.Flags().String("period-end", "", "Period end date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCertificationRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCertificationRequirementsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}

	if cmd.Flags().Changed("period-start") {
		attributes["period-start"] = opts.PeriodStart
	}
	if cmd.Flags().Changed("period-end") {
		attributes["period-end"] = opts.PeriodEnd
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "certification-requirements",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/certification-requirements/"+opts.ID, jsonBody)
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

	row := buildCertificationRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated certification requirement %s\n", row.ID)
	return nil
}

func parseDoCertificationRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doCertificationRequirementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	periodStart, _ := cmd.Flags().GetString("period-start")
	periodEnd, _ := cmd.Flags().GetString("period-end")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCertificationRequirementsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}, nil
}
