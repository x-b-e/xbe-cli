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

type doProjectsFileImportsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	FileImportType string
	IsDryRun       bool
	FileImportID   string
	SubjectType    string
	SubjectID      string
}

func newDoProjectsFileImportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Import projects from a file",
		Long: `Import projects from a file.

Required flags:
  --file-import       File import ID
  --file-import-type  File import type (depends on broker)

Optional flags:
  --is-dry-run    Validate without importing
  --subject-type  Subject type (use with --subject-id)
  --subject-id    Subject ID (use with --subject-type)

If subject is omitted, the file import broker is used.`,
		Example: `  # Import projects from a file
  xbe do projects-file-imports create \
    --file-import 123 \
    --file-import-type SageProjectsFileImport

  # Dry run a projects file import
  xbe do projects-file-imports create \
    --file-import 123 \
    --file-import-type SageProjectsFileImport \
    --is-dry-run

  # Specify a subject
  xbe do projects-file-imports create \
    --file-import 123 \
    --file-import-type SageProjectsFileImport \
    --subject-type brokers \
    --subject-id 456`,
		Args: cobra.NoArgs,
		RunE: runDoProjectsFileImportsCreate,
	}
	initDoProjectsFileImportsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectsFileImportsCmd.AddCommand(newDoProjectsFileImportsCreateCmd())
}

func initDoProjectsFileImportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("file-import-type", "", "File import type (required)")
	cmd.Flags().Bool("is-dry-run", false, "Validate without importing")
	cmd.Flags().String("file-import", "", "File import ID (required)")
	cmd.Flags().String("subject-type", "", "Subject type (use with --subject-id)")
	cmd.Flags().String("subject-id", "", "Subject ID (use with --subject-type)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("file-import-type")
	cmd.MarkFlagRequired("file-import")
}

func runDoProjectsFileImportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectsFileImportsCreateOptions(cmd)
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

	attributes := map[string]any{
		"file-import-type": opts.FileImportType,
	}
	if cmd.Flags().Changed("is-dry-run") {
		attributes["is-dry-run"] = opts.IsDryRun
	}

	if (strings.TrimSpace(opts.SubjectType) != "") != (strings.TrimSpace(opts.SubjectID) != "") {
		err := fmt.Errorf("--subject-type and --subject-id must be set together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"file-import": map[string]any{
			"data": map[string]any{
				"type": "file-imports",
				"id":   opts.FileImportID,
			},
		},
	}

	if strings.TrimSpace(opts.SubjectType) != "" {
		relationships["subject"] = map[string]any{
			"data": map[string]any{
				"type": opts.SubjectType,
				"id":   opts.SubjectID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "projects-file-imports",
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

	body, _, err := client.Post(cmd.Context(), "/v1/projects-file-imports", jsonBody)
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

	details := buildProjectsFileImportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectsFileImportDetails(cmd, details)
}

func parseDoProjectsFileImportsCreateOptions(cmd *cobra.Command) (doProjectsFileImportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	fileImportType, _ := cmd.Flags().GetString("file-import-type")
	isDryRun, _ := cmd.Flags().GetBool("is-dry-run")
	fileImportID, _ := cmd.Flags().GetString("file-import")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectsFileImportsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		FileImportType: fileImportType,
		IsDryRun:       isDryRun,
		FileImportID:   fileImportID,
		SubjectType:    subjectType,
		SubjectID:      subjectID,
	}, nil
}
