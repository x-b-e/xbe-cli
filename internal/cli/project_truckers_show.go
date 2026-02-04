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

type projectTruckersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTruckerDetails struct {
	ID                                                     string `json:"id"`
	Project                                                string `json:"project_id,omitempty"`
	Trucker                                                string `json:"trucker_id,omitempty"`
	IsExcludedFromTimeCardPayrollCertificationRequirements bool   `json:"is_excluded_from_time_card_payroll_certification_requirements"`
	CreatedAt                                              string `json:"created_at,omitempty"`
	UpdatedAt                                              string `json:"updated_at,omitempty"`
}

func newProjectTruckersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project trucker details",
		Long: `Show the full details of a project trucker.

Output Fields:
  ID
  Project ID
  Trucker ID
  Excluded From Time Card Payroll Certification Requirements
  Created At
  Updated At

Arguments:
  <id>    The project trucker ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project trucker
  xbe view project-truckers show 123

  # Output as JSON
  xbe view project-truckers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTruckersShow,
	}
	initProjectTruckersShowFlags(cmd)
	return cmd
}

func init() {
	projectTruckersCmd.AddCommand(newProjectTruckersShowCmd())
}

func initProjectTruckersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTruckersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTruckersShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project trucker id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-truckers]", "created-at,updated-at,project,trucker,is-excluded-from-time-card-payroll-certification-requirements")

	body, _, err := client.Get(cmd.Context(), "/v1/project-truckers/"+id, query)
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

	details := buildProjectTruckerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTruckerDetails(cmd, details)
}

func parseProjectTruckersShowOptions(cmd *cobra.Command) (projectTruckersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTruckersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTruckerDetails(resp jsonAPISingleResponse) projectTruckerDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectTruckerDetails{
		ID: resource.ID,
		IsExcludedFromTimeCardPayrollCertificationRequirements: boolAttr(attrs, "is-excluded-from-time-card-payroll-certification-requirements"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.Project = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.Trucker = rel.Data.ID
	}

	return details
}

func renderProjectTruckerDetails(cmd *cobra.Command, details projectTruckerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Project != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.Project)
	}
	if details.Trucker != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.Trucker)
	}
	fmt.Fprintf(out, "Excluded From Time Card Payroll Certification Requirements: %t\n", details.IsExcludedFromTimeCardPayrollCertificationRequirements)
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
