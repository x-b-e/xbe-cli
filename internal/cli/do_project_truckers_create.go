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

type doProjectTruckersCreateOptions struct {
	BaseURL                                                string
	Token                                                  string
	JSON                                                   bool
	Project                                                string
	Trucker                                                string
	IsExcludedFromTimeCardPayrollCertificationRequirements bool
}

func newDoProjectTruckersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project trucker",
		Long: `Create a project trucker.

Required flags:
  --project  Project ID
  --trucker  Trucker ID

Optional flags:
  --is-excluded-from-time-card-payroll-certification-requirements  Exclude from payroll certification requirements (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project trucker
  xbe do project-truckers create --project 123 --trucker 456

  # Create with exclusion flag
  xbe do project-truckers create \
    --project 123 \
    --trucker 456 \
    --is-excluded-from-time-card-payroll-certification-requirements=true`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTruckersCreate,
	}
	initDoProjectTruckersCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTruckersCmd.AddCommand(newDoProjectTruckersCreateCmd())
}

func initDoProjectTruckersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().Bool("is-excluded-from-time-card-payroll-certification-requirements", false, "Exclude from payroll certification requirements (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTruckersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTruckersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Trucker) == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("is-excluded-from-time-card-payroll-certification-requirements") {
		attributes["is-excluded-from-time-card-payroll-certification-requirements"] = opts.IsExcludedFromTimeCardPayrollCertificationRequirements
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	data := map[string]any{
		"type":          "project-truckers",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-truckers", jsonBody)
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

	if opts.JSON {
		row := buildProjectTruckerRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
		if len(row) > 0 {
			return writeJSON(cmd.OutOrStdout(), row[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project trucker %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectTruckersCreateOptions(cmd *cobra.Command) (doProjectTruckersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	trucker, _ := cmd.Flags().GetString("trucker")
	isExcluded, _ := cmd.Flags().GetBool("is-excluded-from-time-card-payroll-certification-requirements")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTruckersCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Project: project,
		Trucker: trucker,
		IsExcludedFromTimeCardPayrollCertificationRequirements: isExcluded,
	}, nil
}
