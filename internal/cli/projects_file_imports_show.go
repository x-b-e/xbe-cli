package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectsFileImportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectsFileImportDetails struct {
	ID             string `json:"id"`
	FileImportType string `json:"file_import_type,omitempty"`
	IsDryRun       bool   `json:"is_dry_run,omitempty"`
	Status         string `json:"status,omitempty"`
	Results        any    `json:"results,omitempty"`
	FileImportID   string `json:"file_import_id,omitempty"`
	SubjectType    string `json:"subject_type,omitempty"`
	SubjectID      string `json:"subject_id,omitempty"`
}

func newProjectsFileImportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show projects file import details",
		Long: `Show full details of a projects file import.

Output Fields:
  ID          Import identifier
  Type        File import type
  Dry Run     Whether the import is a dry run
  Status      Processing status
  File Import File import ID
  Subject     Subject type and ID
  Results     Import results payload

Arguments:
  <id>    Projects file import ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a projects file import
  xbe view projects-file-imports show 123

  # JSON output
  xbe view projects-file-imports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectsFileImportsShow,
	}
	initProjectsFileImportsShowFlags(cmd)
	return cmd
}

func init() {
	projectsFileImportsCmd.AddCommand(newProjectsFileImportsShowCmd())
}

func initProjectsFileImportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectsFileImportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectsFileImportsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("projects file import id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[projects-file-imports]", "file-import-type,is-dry-run,status,results,file-import,subject")

	body, _, err := client.Get(cmd.Context(), "/v1/projects-file-imports/"+id, query)
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

func parseProjectsFileImportsShowOptions(cmd *cobra.Command) (projectsFileImportsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectsFileImportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectsFileImportDetails(resp jsonAPISingleResponse) projectsFileImportDetails {
	attrs := resp.Data.Attributes
	details := projectsFileImportDetails{
		ID:             resp.Data.ID,
		FileImportType: stringAttr(attrs, "file-import-type"),
		IsDryRun:       boolAttr(attrs, "is-dry-run"),
		Status:         stringAttr(attrs, "status"),
		Results:        anyAttr(attrs, "results"),
	}

	if rel, ok := resp.Data.Relationships["file-import"]; ok && rel.Data != nil {
		details.FileImportID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
		details.SubjectType = rel.Data.Type
		details.SubjectID = rel.Data.ID
	}

	return details
}

func renderProjectsFileImportDetails(cmd *cobra.Command, details projectsFileImportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FileImportType != "" {
		fmt.Fprintf(out, "File Import Type: %s\n", details.FileImportType)
	} else {
		fmt.Fprintln(out, "File Import Type: (none)")
	}
	fmt.Fprintf(out, "Dry Run: %s\n", formatBool(details.IsDryRun))
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	} else {
		fmt.Fprintln(out, "Status: (none)")
	}
	if details.FileImportID != "" {
		fmt.Fprintf(out, "File Import ID: %s\n", details.FileImportID)
	} else {
		fmt.Fprintln(out, "File Import ID: (none)")
	}
	if details.SubjectType != "" && details.SubjectID != "" {
		fmt.Fprintf(out, "Subject: %s/%s\n", details.SubjectType, details.SubjectID)
	} else {
		fmt.Fprintln(out, "Subject: (none)")
	}
	writeAnySection(out, "Results", details.Results)

	return nil
}
