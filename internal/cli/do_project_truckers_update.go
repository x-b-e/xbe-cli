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

type doProjectTruckersUpdateOptions struct {
	BaseURL                                                string
	Token                                                  string
	JSON                                                   bool
	ID                                                     string
	IsExcludedFromTimeCardPayrollCertificationRequirements bool
}

func newDoProjectTruckersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project trucker",
		Long: `Update a project trucker.

Provide the project trucker ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --is-excluded-from-time-card-payroll-certification-requirements  Exclusion flag (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update the exclusion flag
  xbe do project-truckers update 123 \
    --is-excluded-from-time-card-payroll-certification-requirements=true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTruckersUpdate,
	}
	initDoProjectTruckersUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTruckersCmd.AddCommand(newDoProjectTruckersUpdateCmd())
}

func initDoProjectTruckersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-excluded-from-time-card-payroll-certification-requirements", false, "Exclude from payroll certification requirements (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTruckersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTruckersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-excluded-from-time-card-payroll-certification-requirements") {
		attributes["is-excluded-from-time-card-payroll-certification-requirements"] = opts.IsExcludedFromTimeCardPayrollCertificationRequirements
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --is-excluded-from-time-card-payroll-certification-requirements")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-truckers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-truckers/"+opts.ID, jsonBody)
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

	row := buildProjectTruckerRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
	if opts.JSON {
		if len(row) > 0 {
			return writeJSON(cmd.OutOrStdout(), row[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project trucker %s\n", opts.ID)
	return nil
}

func parseDoProjectTruckersUpdateOptions(cmd *cobra.Command, args []string) (doProjectTruckersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isExcluded, _ := cmd.Flags().GetBool("is-excluded-from-time-card-payroll-certification-requirements")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTruckersUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		IsExcludedFromTimeCardPayrollCertificationRequirements: isExcluded,
	}, nil
}
