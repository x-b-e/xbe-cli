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

type doTimeCardPayrollCertificationsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	TimeCardID string
}

func newDoTimeCardPayrollCertificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time card payroll certification",
		Long: `Create a time card payroll certification.

Required flags:
  --time-card  Time card ID (required)`,
		Example: `  # Certify a time card for payroll
  xbe do time-card-payroll-certifications create --time-card 123

  # Output as JSON
  xbe do time-card-payroll-certifications create --time-card 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardPayrollCertificationsCreate,
	}
	initDoTimeCardPayrollCertificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardPayrollCertificationsCmd.AddCommand(newDoTimeCardPayrollCertificationsCreateCmd())
}

func initDoTimeCardPayrollCertificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("time-card")
}

func runDoTimeCardPayrollCertificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardPayrollCertificationsCreateOptions(cmd)
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

	if opts.TimeCardID == "" {
		err := fmt.Errorf("--time-card is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"time-card": map[string]any{
			"data": map[string]any{
				"type": "time-cards",
				"id":   opts.TimeCardID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-card-payroll-certifications",
			"attributes":    map[string]any{},
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-payroll-certifications", jsonBody)
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

	row := buildTimeCardPayrollCertificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card payroll certification %s\n", row.ID)
	return nil
}

func parseDoTimeCardPayrollCertificationsCreateOptions(cmd *cobra.Command) (doTimeCardPayrollCertificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCardID, _ := cmd.Flags().GetString("time-card")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardPayrollCertificationsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		TimeCardID: timeCardID,
	}, nil
}
