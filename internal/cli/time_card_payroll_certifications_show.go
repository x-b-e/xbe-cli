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

type timeCardPayrollCertificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardPayrollCertificationDetails struct {
	ID          string `json:"id"`
	TimeCardID  string `json:"time_card_id,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

func newTimeCardPayrollCertificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card payroll certification details",
		Long: `Show the full details of a time card payroll certification.

Output Fields:
  ID         Payroll certification identifier
  Time Card  Time card ID
  Created By User who certified the time card
  Created At Certification timestamp
  Updated At Last update timestamp

Arguments:
  <id>    The time card payroll certification ID (required). You can find IDs using the list command.`,
		Example: `  # Show a time card payroll certification
  xbe view time-card-payroll-certifications show 123

  # Get JSON output
  xbe view time-card-payroll-certifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardPayrollCertificationsShow,
	}
	initTimeCardPayrollCertificationsShowFlags(cmd)
	return cmd
}

func init() {
	timeCardPayrollCertificationsCmd.AddCommand(newTimeCardPayrollCertificationsShowCmd())
}

func initTimeCardPayrollCertificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardPayrollCertificationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTimeCardPayrollCertificationsShowOptions(cmd)
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
		return fmt.Errorf("time card payroll certification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-payroll-certifications/"+id, nil)
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

	details := buildTimeCardPayrollCertificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardPayrollCertificationDetails(cmd, details)
}

func parseTimeCardPayrollCertificationsShowOptions(cmd *cobra.Command) (timeCardPayrollCertificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardPayrollCertificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardPayrollCertificationDetails(resp jsonAPISingleResponse) timeCardPayrollCertificationDetails {
	resource := resp.Data
	attrs := resource.Attributes
	results := timeCardPayrollCertificationDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		results.TimeCardID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		results.CreatedByID = rel.Data.ID
	}

	return results
}

func renderTimeCardPayrollCertificationDetails(cmd *cobra.Command, details timeCardPayrollCertificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card: %s\n", details.TimeCardID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
