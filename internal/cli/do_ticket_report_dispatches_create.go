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

type doTicketReportDispatchesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	TicketReport string
}

func newDoTicketReportDispatchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a ticket report dispatch",
		Long: `Create a ticket report dispatch.

Required flags:
  --ticket-report  Ticket report ID to dispatch (required)

Dispatches trigger fulfillment for the specified ticket report. The ticket report
must have a successful transform result with no errors.`,
		Example: `  # Dispatch a ticket report
  xbe do ticket-report-dispatches create --ticket-report 123

  # Output as JSON
  xbe do ticket-report-dispatches create --ticket-report 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTicketReportDispatchesCreate,
	}
	initDoTicketReportDispatchesCreateFlags(cmd)
	return cmd
}

func init() {
	doTicketReportDispatchesCmd.AddCommand(newDoTicketReportDispatchesCreateCmd())
}

func initDoTicketReportDispatchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("ticket-report", "", "Ticket report ID to dispatch (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTicketReportDispatchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTicketReportDispatchesCreateOptions(cmd)
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
			"type":          "ticket-report-dispatches",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/ticket-report-dispatches", jsonBody)
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

	row := ticketReportDispatchRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created ticket report dispatch %s\n", row.ID)
	return nil
}

func parseDoTicketReportDispatchesCreateOptions(cmd *cobra.Command) (doTicketReportDispatchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	ticketReport, _ := cmd.Flags().GetString("ticket-report")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTicketReportDispatchesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		TicketReport: ticketReport,
	}, nil
}
