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

type materialTransactionsExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionsExportDetails struct {
	ID                      string   `json:"id"`
	Status                  string   `json:"status,omitempty"`
	FileName                string   `json:"file_name,omitempty"`
	Body                    string   `json:"body,omitempty"`
	MimeType                string   `json:"mime_type,omitempty"`
	FormatterErrorsDetails  any      `json:"formatter_errors_details,omitempty"`
	OrganizationType        string   `json:"organization_type,omitempty"`
	OrganizationID          string   `json:"organization_id,omitempty"`
	BrokerID                string   `json:"broker_id,omitempty"`
	OrganizationFormatterID string   `json:"organization_formatter_id,omitempty"`
	CreatedByID             string   `json:"created_by_id,omitempty"`
	MaterialTransactionIDs  []string `json:"material_transaction_ids,omitempty"`
}

func newMaterialTransactionsExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction export details",
		Long: `Show the full details of a material transaction export.

Output Fields:
  ID
  Status
  File Name
  Body
  Mime Type
  Formatter Errors Details
  Organization (type + ID)
  Broker ID
  Organization Formatter ID
  Created By
  Material Transaction IDs

Arguments:
  <id>    The export ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an export
  xbe view material-transactions-exports show 123

  # JSON output
  xbe view material-transactions-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionsExportsShow,
	}
	initMaterialTransactionsExportsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionsExportsCmd.AddCommand(newMaterialTransactionsExportsShowCmd())
}

func initMaterialTransactionsExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionsExportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionsExportsShowOptions(cmd)
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
		return fmt.Errorf("export id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transactions-exports]", "body,file-name,formatter-errors-details,mime-type,status,broker,created-by,organization,organization-formatter,material-transactions")
	query.Set("fields[material-transactions]", "ticket-number")
	query.Set("include", "material-transactions")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transactions-exports/"+id, query)
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

	details := buildMaterialTransactionsExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionsExportDetails(cmd, details)
}

func parseMaterialTransactionsExportsShowOptions(cmd *cobra.Command) (materialTransactionsExportsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return materialTransactionsExportsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return materialTransactionsExportsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return materialTransactionsExportsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return materialTransactionsExportsShowOptions{}, err
	}

	return materialTransactionsExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionsExportDetails(resp jsonAPISingleResponse) materialTransactionsExportDetails {
	resource := resp.Data
	attrs := resource.Attributes

	materialTransactionIDs := relationshipIDsFromMap(resource.Relationships, "material-transactions")
	if len(materialTransactionIDs) == 0 && len(resp.Included) > 0 {
		for _, included := range resp.Included {
			if included.Type == "material-transactions" && included.ID != "" {
				materialTransactionIDs = append(materialTransactionIDs, included.ID)
			}
		}
	}

	details := materialTransactionsExportDetails{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		FileName:                stringAttr(attrs, "file-name"),
		Body:                    stringAttr(attrs, "body"),
		MimeType:                stringAttr(attrs, "mime-type"),
		FormatterErrorsDetails:  attrs["formatter-errors-details"],
		BrokerID:                relationshipIDFromMap(resource.Relationships, "broker"),
		OrganizationFormatterID: relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:             relationshipIDFromMap(resource.Relationships, "created-by"),
		MaterialTransactionIDs:  materialTransactionIDs,
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	return details
}

func renderMaterialTransactionsExportDetails(cmd *cobra.Command, details materialTransactionsExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.FileName != "" {
		fmt.Fprintf(out, "File Name: %s\n", details.FileName)
	}
	if details.MimeType != "" {
		fmt.Fprintf(out, "Mime Type: %s\n", details.MimeType)
	}
	if details.OrganizationType != "" || details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization: %s:%s\n", details.OrganizationType, details.OrganizationID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.OrganizationFormatterID != "" {
		fmt.Fprintf(out, "Organization Formatter: %s\n", details.OrganizationFormatterID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}

	if details.Body != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Body:")
		fmt.Fprintln(out, details.Body)
	}
	if details.FormatterErrorsDetails != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Formatter Errors Details:")
		fmt.Fprintln(out, formatJSONBlock(details.FormatterErrorsDetails, "  "))
	}
	if len(details.MaterialTransactionIDs) > 0 {
		fmt.Fprintf(out, "Material Transaction IDs: %s\n", strings.Join(details.MaterialTransactionIDs, ", "))
	}

	return nil
}
