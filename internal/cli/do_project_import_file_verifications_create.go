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

type doProjectImportFileVerificationsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	VerificationType string
	IsDryRun         bool
	FileImport       string
	Project          string
}

func newDoProjectImportFileVerificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project import file verification",
		Long: `Create a project import file verification.

Required flags:
  --verification-type  Verification type (e.g., Bid2Win)
  --file-import        File import ID
  --project            Project ID

Optional flags:
  --is-dry-run  Create a dry run verification (does not process the file)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Verify a project import file
  xbe do project-import-file-verifications create \
    --verification-type Bid2Win \
    --file-import 123 \
    --project 456

  # Dry run verification
  xbe do project-import-file-verifications create \
    --verification-type Bid2Win \
    --file-import 123 \
    --project 456 \
    --is-dry-run \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectImportFileVerificationsCreate,
	}
	initDoProjectImportFileVerificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectImportFileVerificationsCmd.AddCommand(newDoProjectImportFileVerificationsCreateCmd())
}

func initDoProjectImportFileVerificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("verification-type", "", "Verification type (required)")
	cmd.Flags().String("file-import", "", "File import ID (required)")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().Bool("is-dry-run", false, "Create a dry run verification")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectImportFileVerificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectImportFileVerificationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.VerificationType) == "" {
		err := fmt.Errorf("--verification-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.FileImport) == "" {
		err := fmt.Errorf("--file-import is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"verification-type": opts.VerificationType,
	}
	if cmd.Flags().Changed("is-dry-run") {
		attributes["is-dry-run"] = opts.IsDryRun
	}

	relationships := map[string]any{
		"file-import": map[string]any{
			"data": map[string]any{
				"type": "file-imports",
				"id":   opts.FileImport,
			},
		},
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-import-file-verifications",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-import-file-verifications", jsonBody)
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

	if opts.JSON {
		rows := buildProjectImportFileVerificationRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
		if len(rows) > 0 {
			return writeJSON(cmd.OutOrStdout(), rows[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project import file verification %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectImportFileVerificationsCreateOptions(cmd *cobra.Command) (doProjectImportFileVerificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	verificationType, _ := cmd.Flags().GetString("verification-type")
	fileImport, _ := cmd.Flags().GetString("file-import")
	project, _ := cmd.Flags().GetString("project")
	isDryRun, _ := cmd.Flags().GetBool("is-dry-run")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectImportFileVerificationsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		VerificationType: verificationType,
		IsDryRun:         isDryRun,
		FileImport:       fileImport,
		Project:          project,
	}, nil
}
