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

type doTicketReportImportsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	TicketReport string
}

func newDoTicketReportImportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a ticket report import",
		Long: `Create a ticket report import.

Required flags:
  --ticket-report  Ticket report ID to import (required)

Imports run asynchronously and may take time to complete. Only one import can be
pending or processing for a given ticket report.`,
		Example: `  # Create a ticket report import
  xbe do ticket-report-imports create --ticket-report 123

  # Output as JSON
  xbe do ticket-report-imports create --ticket-report 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTicketReportImportsCreate,
	}
	initDoTicketReportImportsCreateFlags(cmd)
	return cmd
}

func init() {
	doTicketReportImportsCmd.AddCommand(newDoTicketReportImportsCreateCmd())
}

func initDoTicketReportImportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("ticket-report", "", "Ticket report ID to import (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTicketReportImportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTicketReportImportsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TicketReport) == "" {
		err := fmt.Errorf("--ticket-report is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"ticket-report": map[string]any{
			"data": map[string]any{
				"type": "ticket-reports",
				"id":   opts.TicketReport,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "ticket-report-imports",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/ticket-report-imports", jsonBody)
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

	row := ticketReportImportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created ticket report import %s\n", row.ID)
	return nil
}

func parseDoTicketReportImportsCreateOptions(cmd *cobra.Command) (doTicketReportImportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	ticketReport, _ := cmd.Flags().GetString("ticket-report")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTicketReportImportsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		TicketReport: ticketReport,
	}, nil
}
