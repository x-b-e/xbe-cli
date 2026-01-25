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

type rawTransportProjectsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawTransportProjectDetails struct {
	ID                    string   `json:"id"`
	ExternalProjectNumber string   `json:"external_project_number,omitempty"`
	Importer              string   `json:"importer,omitempty"`
	ImportStatus          string   `json:"import_status,omitempty"`
	ImportErrors          []string `json:"import_errors,omitempty"`
	IsManaged             bool     `json:"is_managed"`
	TablesRowversionMin   string   `json:"tables_rowversion_min,omitempty"`
	TablesRowversionMax   string   `json:"tables_rowversion_max,omitempty"`
	BrokerID              string   `json:"broker_id,omitempty"`
	ProjectID             string   `json:"project_id,omitempty"`
	CreatedByID           string   `json:"created_by_id,omitempty"`
	Tables                any      `json:"tables,omitempty"`
}

func newRawTransportProjectsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw transport project details",
		Long: `Show the full details of a raw transport project.

Output Fields:
  ID
  External Project Number
  Importer
  Import Status
  Import Errors
  Is Managed
  Tables Rowversion Min
  Tables Rowversion Max
  Broker
  Project
  Created By
  Tables

Arguments:
  <id>    The raw transport project ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show raw transport project details
  xbe view raw-transport-projects show 123

  # Output as JSON
  xbe view raw-transport-projects show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawTransportProjectsShow,
	}
	initRawTransportProjectsShowFlags(cmd)
	return cmd
}

func init() {
	rawTransportProjectsCmd.AddCommand(newRawTransportProjectsShowCmd())
}

func initRawTransportProjectsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportProjectsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRawTransportProjectsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("raw transport project id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-transport-projects]", "external-project-number,importer,import-status,import-errors,is-managed,tables-rowversion-min,tables-rowversion-max,tables")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-projects/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildRawTransportProjectDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawTransportProjectDetails(cmd, details)
}

func parseRawTransportProjectsShowOptions(cmd *cobra.Command) (rawTransportProjectsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportProjectsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawTransportProjectDetails(resp jsonAPISingleResponse) rawTransportProjectDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := rawTransportProjectDetails{
		ID:                    resource.ID,
		ExternalProjectNumber: stringAttr(attrs, "external-project-number"),
		Importer:              stringAttr(attrs, "importer"),
		ImportStatus:          stringAttr(attrs, "import-status"),
		ImportErrors:          stringSliceAttr(attrs, "import-errors"),
		IsManaged:             boolAttr(attrs, "is-managed"),
		TablesRowversionMin:   stringAttr(attrs, "tables-rowversion-min"),
		TablesRowversionMax:   stringAttr(attrs, "tables-rowversion-max"),
		BrokerID:              relationshipIDFromMap(resource.Relationships, "broker"),
		ProjectID:             relationshipIDFromMap(resource.Relationships, "project"),
		CreatedByID:           relationshipIDFromMap(resource.Relationships, "created-by"),
		Tables:                attrs["tables"],
	}

	return details
}

func renderRawTransportProjectDetails(cmd *cobra.Command, details rawTransportProjectDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalProjectNumber != "" {
		fmt.Fprintf(out, "External Project Number: %s\n", details.ExternalProjectNumber)
	}
	if details.Importer != "" {
		fmt.Fprintf(out, "Importer: %s\n", details.Importer)
	}
	if details.ImportStatus != "" {
		fmt.Fprintf(out, "Import Status: %s\n", details.ImportStatus)
	}
	fmt.Fprintf(out, "Is Managed: %t\n", details.IsManaged)
	if details.TablesRowversionMin != "" {
		fmt.Fprintf(out, "Tables Rowversion Min: %s\n", details.TablesRowversionMin)
	}
	if details.TablesRowversionMax != "" {
		fmt.Fprintf(out, "Tables Rowversion Max: %s\n", details.TablesRowversionMax)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if len(details.ImportErrors) > 0 {
		fmt.Fprintf(out, "Import Errors: %s\n", strings.Join(details.ImportErrors, ", "))
	}
	if details.Tables != nil {
		fmt.Fprintln(out, "\nTables:")
		fmt.Fprintln(out, formatJSONBlock(details.Tables, "  "))
	}

	return nil
}
