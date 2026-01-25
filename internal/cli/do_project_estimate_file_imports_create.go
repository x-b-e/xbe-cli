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

type doProjectEstimateFileImportsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	FileImportType             string
	IsDryRun                   bool
	ShouldUpdateFileExtraction bool
	FileImportID               string
	ProjectID                  string
}

type projectEstimateFileImportDetails struct {
	ID                         string `json:"id"`
	FileImportType             string `json:"file_import_type,omitempty"`
	IsDryRun                   bool   `json:"is_dry_run,omitempty"`
	ShouldUpdateFileExtraction bool   `json:"should_update_file_extraction,omitempty"`
	FileImportID               string `json:"file_import_id,omitempty"`
	ProjectID                  string `json:"project_id,omitempty"`
}

func newDoProjectEstimateFileImportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Import a project estimate file",
		Long: `Import a project estimate file.

Required flags:
  --project           Project ID
  --file-import       File import ID
  --file-import-type  File import type (depends on broker)

Optional flags:
  --is-dry-run                    Validate without importing
  --should-update-file-extraction Update PDF extraction data when importing`,
		Example: `  # Import a project estimate file
  xbe do project-estimate-file-imports create \
    --project 123 \
    --file-import 456 \
    --file-import-type Bid2Win

  # Dry run the import
  xbe do project-estimate-file-imports create \
    --project 123 \
    --file-import 456 \
    --file-import-type Bid2Win \
    --is-dry-run`,
		Args: cobra.NoArgs,
		RunE: runDoProjectEstimateFileImportsCreate,
	}
	initDoProjectEstimateFileImportsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectEstimateFileImportsCmd.AddCommand(newDoProjectEstimateFileImportsCreateCmd())
}

func initDoProjectEstimateFileImportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("file-import-type", "", "File import type (required)")
	cmd.Flags().Bool("is-dry-run", false, "Validate without importing")
	cmd.Flags().Bool("should-update-file-extraction", false, "Update PDF extraction data when importing")
	cmd.Flags().String("file-import", "", "File import ID (required)")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("file-import-type")
	cmd.MarkFlagRequired("file-import")
	cmd.MarkFlagRequired("project")
}

func runDoProjectEstimateFileImportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectEstimateFileImportsCreateOptions(cmd)
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
	if cmd.Flags().Changed("should-update-file-extraction") {
		attributes["should-update-file-extraction"] = opts.ShouldUpdateFileExtraction
	}

	relationships := map[string]any{
		"file-import": map[string]any{
			"data": map[string]any{
				"type": "file-imports",
				"id":   opts.FileImportID,
			},
		},
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.ProjectID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-estimate-file-imports",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-estimate-file-imports", jsonBody)
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

	details := buildProjectEstimateFileImportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectEstimateFileImportDetails(cmd, details)
}

func parseDoProjectEstimateFileImportsCreateOptions(cmd *cobra.Command) (doProjectEstimateFileImportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	fileImportType, _ := cmd.Flags().GetString("file-import-type")
	isDryRun, _ := cmd.Flags().GetBool("is-dry-run")
	shouldUpdateFileExtraction, _ := cmd.Flags().GetBool("should-update-file-extraction")
	fileImportID, _ := cmd.Flags().GetString("file-import")
	projectID, _ := cmd.Flags().GetString("project")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectEstimateFileImportsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		FileImportType:             fileImportType,
		IsDryRun:                   isDryRun,
		ShouldUpdateFileExtraction: shouldUpdateFileExtraction,
		FileImportID:               fileImportID,
		ProjectID:                  projectID,
	}, nil
}

func buildProjectEstimateFileImportDetails(resp jsonAPISingleResponse) projectEstimateFileImportDetails {
	attrs := resp.Data.Attributes
	details := projectEstimateFileImportDetails{
		ID:                         resp.Data.ID,
		FileImportType:             stringAttr(attrs, "file-import-type"),
		IsDryRun:                   boolAttr(attrs, "is-dry-run"),
		ShouldUpdateFileExtraction: boolAttr(attrs, "should-update-file-extraction"),
	}

	if rel, ok := resp.Data.Relationships["file-import"]; ok && rel.Data != nil {
		details.FileImportID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}

	return details
}

func renderProjectEstimateFileImportDetails(cmd *cobra.Command, details projectEstimateFileImportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FileImportType != "" {
		fmt.Fprintf(out, "File Import Type: %s\n", details.FileImportType)
	} else {
		fmt.Fprintln(out, "File Import Type: (none)")
	}
	fmt.Fprintf(out, "Dry Run: %s\n", formatBool(details.IsDryRun))
	fmt.Fprintf(out, "Update File Extraction: %s\n", formatBool(details.ShouldUpdateFileExtraction))
	if details.FileImportID != "" {
		fmt.Fprintf(out, "File Import ID: %s\n", details.FileImportID)
	} else {
		fmt.Fprintln(out, "File Import ID: (none)")
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	} else {
		fmt.Fprintln(out, "Project ID: (none)")
	}

	return nil
}
