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

type doTicketReportsUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	FileName string
}

func newDoTicketReportsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a ticket report",
		Long: `Update a ticket report.

Optional flags:
  --file-name   Update the stored file name`,
		Example: `  # Update the file name
  xbe do ticket-reports update 123 --file-name "updated-report.csv"

  # JSON output
  xbe do ticket-reports update 123 --file-name "updated-report.csv" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTicketReportsUpdate,
	}
	initDoTicketReportsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTicketReportsCmd.AddCommand(newDoTicketReportsUpdateCmd())
}

func initDoTicketReportsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("file-name", "", "Update the stored file name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTicketReportsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTicketReportsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("file-name") {
		attributes["file-name"] = opts.FileName
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "ticket-reports",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/ticket-reports/"+opts.ID, jsonBody)
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

	row := buildTicketReportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated ticket report %s\n", row.ID)
	return nil
}

func parseDoTicketReportsUpdateOptions(cmd *cobra.Command, args []string) (doTicketReportsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	fileName, _ := cmd.Flags().GetString("file-name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTicketReportsUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		FileName: fileName,
	}, nil
}
