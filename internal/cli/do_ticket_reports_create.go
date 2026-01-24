package cli

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doTicketReportsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	BrokerID         string
	TicketReportType string
	FilePath         string
	FileName         string
	FileBase64       string
}

func newDoTicketReportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ticket report",
		Long: `Create a new ticket report.

Required flags:
  --ticket-report-type  Ticket report type ID
  --file-path           Path to the ticket file (or use --file)

Optional flags:
  --broker              Broker ID (defaults to ticket report type broker)
  --file-name           Override file name (defaults to the file name from --file-path)
  --file                Base64-encoded file contents (use with --file-name)`,
		Example: `  # Create a ticket report from a file
  xbe do ticket-reports create --ticket-report-type 456 --file-path ./ticket.csv

  # Create with explicit broker and file name
  xbe do ticket-reports create --ticket-report-type 456 --broker 123 \\
    --file-path ./ticket.csv --file-name "jan-report.csv"

  # Create from base64 file contents
  xbe do ticket-reports create --ticket-report-type 456 --file "BASE64..." --file-name "ticket.csv"

  # JSON output
  xbe do ticket-reports create --ticket-report-type 456 --file-path ./ticket.csv --json`,
		RunE: runDoTicketReportsCreate,
	}
	initDoTicketReportsCreateFlags(cmd)
	return cmd
}

func init() {
	doTicketReportsCmd.AddCommand(newDoTicketReportsCreateCmd())
}

func initDoTicketReportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("ticket-report-type", "", "Ticket report type ID (required)")
	cmd.Flags().String("file-path", "", "Path to the ticket file")
	cmd.Flags().String("file-name", "", "Override file name")
	cmd.Flags().String("file", "", "Base64-encoded file contents")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("ticket-report-type")
}

func runDoTicketReportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTicketReportsCreateOptions(cmd)
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

	filePayload, fileName, err := resolveTicketReportFilePayload(opts)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"file":      filePayload,
		"file-name": fileName,
	}

	relationships := map[string]any{
		"ticket-report-type": map[string]any{
			"data": map[string]any{
				"type": "ticket-report-types",
				"id":   opts.TicketReportType,
			},
		},
	}
	if strings.TrimSpace(opts.BrokerID) != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "ticket-reports",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/ticket-reports", jsonBody)
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

	if row.FileName != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created ticket report %s (%s)\n", row.ID, row.FileName)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Created ticket report %s\n", row.ID)
	}
	return nil
}

func parseDoTicketReportsCreateOptions(cmd *cobra.Command) (doTicketReportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	ticketReportType, _ := cmd.Flags().GetString("ticket-report-type")
	filePath, _ := cmd.Flags().GetString("file-path")
	fileName, _ := cmd.Flags().GetString("file-name")
	fileBase64, _ := cmd.Flags().GetString("file")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTicketReportsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		BrokerID:         brokerID,
		TicketReportType: ticketReportType,
		FilePath:         filePath,
		FileName:         fileName,
		FileBase64:       fileBase64,
	}, nil
}

func resolveTicketReportFilePayload(opts doTicketReportsCreateOptions) (string, string, error) {
	if strings.TrimSpace(opts.FilePath) != "" && strings.TrimSpace(opts.FileBase64) != "" {
		return "", "", fmt.Errorf("use either --file-path or --file, not both")
	}

	if strings.TrimSpace(opts.FilePath) == "" && strings.TrimSpace(opts.FileBase64) == "" {
		return "", "", fmt.Errorf("--file-path or --file is required")
	}

	if strings.TrimSpace(opts.FilePath) != "" {
		contents, err := os.ReadFile(opts.FilePath)
		if err != nil {
			return "", "", err
		}
		encoded := base64.StdEncoding.EncodeToString(contents)
		fileName := strings.TrimSpace(opts.FileName)
		if fileName == "" {
			fileName = filepath.Base(opts.FilePath)
		}
		if fileName == "" || fileName == "." || fileName == string(filepath.Separator) {
			return "", "", fmt.Errorf("--file-name is required when file path has no base name")
		}
		return encoded, fileName, nil
	}

	fileName := strings.TrimSpace(opts.FileName)
	if fileName == "" {
		return "", "", fmt.Errorf("--file-name is required when using --file")
	}

	return opts.FileBase64, fileName, nil
}
