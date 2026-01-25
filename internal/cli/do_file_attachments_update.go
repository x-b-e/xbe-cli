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

type doFileAttachmentsUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	FileName  string
	ObjectKey string
}

func newDoFileAttachmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a file attachment",
		Long: `Update a file attachment.

Optional flags:
  --file-name    File name
  --object-key   S3 object key`,
		Example: `  # Update file name
  xbe do file-attachments update 123 --file-name "updated.pdf"

  # Update object key
  xbe do file-attachments update 123 --object-key "uploads/123/updated.pdf"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoFileAttachmentsUpdate,
	}
	initDoFileAttachmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doFileAttachmentsCmd.AddCommand(newDoFileAttachmentsUpdateCmd())
}

func initDoFileAttachmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("file-name", "", "File name")
	cmd.Flags().String("object-key", "", "S3 object key")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoFileAttachmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoFileAttachmentsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("object-key") {
		attributes["object-key"] = opts.ObjectKey
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "file-attachments",
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

	path := fmt.Sprintf("/v1/file-attachments/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	row := buildFileAttachmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated file attachment %s\n", row.ID)
	return nil
}

func parseDoFileAttachmentsUpdateOptions(cmd *cobra.Command, args []string) (doFileAttachmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	fileName, _ := cmd.Flags().GetString("file-name")
	objectKey, _ := cmd.Flags().GetString("object-key")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFileAttachmentsUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		FileName:  fileName,
		ObjectKey: objectKey,
	}, nil
}
