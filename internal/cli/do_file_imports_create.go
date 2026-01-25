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

type doFileImportsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	BrokerID         string
	FileAttachmentID string
	ProcessedAt      string
	Note             string
}

func newDoFileImportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new file import",
		Long: `Create a new file import.

Required flags:
  --broker           Broker ID
  --file-attachment  File attachment ID

Optional flags:
  --processed-at     Processed at timestamp (ISO 8601)
  --note             Import note`,
		Example: `  # Create a file import
  xbe do file-imports create --broker 123 --file-attachment 456

  # Create with note and processed-at
  xbe do file-imports create --broker 123 --file-attachment 456 \\
    --note "Initial import" --processed-at 2024-01-02T03:04:05Z

  # JSON output
  xbe do file-imports create --broker 123 --file-attachment 456 --json`,
		RunE: runDoFileImportsCreate,
	}
	initDoFileImportsCreateFlags(cmd)
	return cmd
}

func init() {
	doFileImportsCmd.AddCommand(newDoFileImportsCreateCmd())
}

func initDoFileImportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("file-attachment", "", "File attachment ID (required)")
	cmd.Flags().String("processed-at", "", "Processed at timestamp (ISO 8601)")
	cmd.Flags().String("note", "", "Import note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("file-attachment")
}

func runDoFileImportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoFileImportsCreateOptions(cmd)
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
	if opts.ProcessedAt != "" {
		attributes["processed-at"] = opts.ProcessedAt
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
		"file-attachment": map[string]any{
			"data": map[string]any{
				"type": "file-attachments",
				"id":   opts.FileAttachmentID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "file-imports",
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

	body, _, err := client.Post(cmd.Context(), "/v1/file-imports", jsonBody)
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

	row := buildFileImportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created file import %s\n", row.ID)
	return nil
}

func parseDoFileImportsCreateOptions(cmd *cobra.Command) (doFileImportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	fileAttachmentID, _ := cmd.Flags().GetString("file-attachment")
	processedAt, _ := cmd.Flags().GetString("processed-at")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFileImportsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		BrokerID:         brokerID,
		FileAttachmentID: fileAttachmentID,
		ProcessedAt:      processedAt,
		Note:             note,
	}, nil
}
