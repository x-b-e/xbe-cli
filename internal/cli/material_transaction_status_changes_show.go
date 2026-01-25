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

type materialTransactionStatusChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionStatusChangeDetails struct {
	ID                    string `json:"id"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
	Status                string `json:"status,omitempty"`
	ChangedAt             string `json:"changed_at,omitempty"`
	Comment               string `json:"comment,omitempty"`
	ChangedByID           string `json:"changed_by_id,omitempty"`
	ChangedByName         string `json:"changed_by_name,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newMaterialTransactionStatusChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction status change details",
		Long: `Show the full details of a material transaction status change.

Output Fields:
  ID                    Status change identifier
  Material Transaction  Material transaction ID
  Status                Status value
  Changed At            When the status change occurred
  Comment               Status change comment
  Changed By            User who made the change (when available)
  Created At            Created timestamp
  Updated At            Updated timestamp

Arguments:
  <id>    The material transaction status change ID (required). You can find IDs using the list command.`,
		Example: `  # Show a status change
  xbe view material-transaction-status-changes show 123

  # Get JSON output
  xbe view material-transaction-status-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionStatusChangesShow,
	}
	initMaterialTransactionStatusChangesShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionStatusChangesCmd.AddCommand(newMaterialTransactionStatusChangesShowCmd())
}

func initMaterialTransactionStatusChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionStatusChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionStatusChangesShowOptions(cmd)
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
		return fmt.Errorf("material transaction status change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-status-changes]", "status,changed-at,comment,created-at,updated-at,material-transaction,changed-by")
	query.Set("fields[material-transactions]", "ticket-number,transaction-at")
	query.Set("fields[users]", "name")
	query.Set("include", "changed-by,material-transaction")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-status-changes/"+id, query)
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

	details := buildMaterialTransactionStatusChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionStatusChangeDetails(cmd, details)
}

func parseMaterialTransactionStatusChangesShowOptions(cmd *cobra.Command) (materialTransactionStatusChangesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return materialTransactionStatusChangesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return materialTransactionStatusChangesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return materialTransactionStatusChangesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return materialTransactionStatusChangesShowOptions{}, err
	}

	return materialTransactionStatusChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionStatusChangeDetails(resp jsonAPISingleResponse) materialTransactionStatusChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialTransactionStatusChangeDetails{
		ID:        resource.ID,
		Status:    strings.TrimSpace(stringAttr(attrs, "status")),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		details.MaterialTransactionID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ChangedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return details
}

func renderMaterialTransactionStatusChangeDetails(cmd *cobra.Command, details materialTransactionStatusChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialTransactionID != "" {
		fmt.Fprintf(out, "Material Transaction: %s\n", details.MaterialTransactionID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ChangedAt != "" {
		fmt.Fprintf(out, "Changed At: %s\n", details.ChangedAt)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	if details.ChangedByName != "" && details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s (%s)\n", details.ChangedByName, details.ChangedByID)
	} else if details.ChangedByName != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByName)
	} else if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
