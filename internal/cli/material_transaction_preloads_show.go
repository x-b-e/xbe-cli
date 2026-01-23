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

type materialTransactionPreloadsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionPreloadDetails struct {
	ID                              string `json:"id"`
	PreloadedAt                     string `json:"preloaded_at,omitempty"`
	PreloadMinutes                  int    `json:"preload_minutes,omitempty"`
	MaterialTransactionID           string `json:"material_transaction_id,omitempty"`
	MaterialTransactionTicketNumber string `json:"material_transaction_ticket_number,omitempty"`
	MaterialTransactionAt           string `json:"material_transaction_at,omitempty"`
	TrailerID                       string `json:"trailer_id,omitempty"`
	TrailerNumber                   string `json:"trailer_number,omitempty"`
}

func newMaterialTransactionPreloadsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction preload details",
		Long: `Show the full details of a material transaction preload.

Output Fields:
  ID
  Preloaded At
  Preload Minutes
  Material Transaction ID
  Material Transaction Ticket Number
  Material Transaction Timestamp
  Trailer ID
  Trailer Number

Arguments:
  <id>    The preload ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show preload details
  xbe view material-transaction-preloads show 123

  # Get JSON output
  xbe view material-transaction-preloads show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionPreloadsShow,
	}
	initMaterialTransactionPreloadsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionPreloadsCmd.AddCommand(newMaterialTransactionPreloadsShowCmd())
}

func initMaterialTransactionPreloadsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionPreloadsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionPreloadsShowOptions(cmd)
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
		return fmt.Errorf("material transaction preload id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-preloads]", "preloaded-at,preload-minutes,material-transaction,trailer")
	query.Set("fields[material-transactions]", "ticket-number,transaction-at")
	query.Set("fields[trailers]", "number")
	query.Set("include", "material-transaction,trailer")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-preloads/"+id, query)
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

	details := buildMaterialTransactionPreloadDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionPreloadDetails(cmd, details)
}

func parseMaterialTransactionPreloadsShowOptions(cmd *cobra.Command) (materialTransactionPreloadsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionPreloadsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionPreloadDetails(resp jsonAPISingleResponse) materialTransactionPreloadDetails {
	resource := resp.Data
	attrs := resource.Attributes
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	trailerID := relationshipIDFromMap(resource.Relationships, "trailer")
	materialTransactionID := relationshipIDFromMap(resource.Relationships, "material-transaction")

	return materialTransactionPreloadDetails{
		ID:                              resource.ID,
		PreloadedAt:                     formatDateTime(stringAttr(attrs, "preloaded-at")),
		PreloadMinutes:                  intAttr(attrs, "preload-minutes"),
		MaterialTransactionID:           materialTransactionID,
		MaterialTransactionTicketNumber: resolveMaterialTransactionTicketNumber(materialTransactionID, included),
		MaterialTransactionAt:           formatDateTime(resolveMaterialTransactionAt(materialTransactionID, included)),
		TrailerID:                       trailerID,
		TrailerNumber:                   resolveTrailerNumber(trailerID, included),
	}
}

func renderMaterialTransactionPreloadDetails(cmd *cobra.Command, details materialTransactionPreloadDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PreloadedAt != "" {
		fmt.Fprintf(out, "Preloaded At: %s\n", details.PreloadedAt)
	}
	if details.PreloadMinutes > 0 {
		fmt.Fprintf(out, "Preload Minutes: %d\n", details.PreloadMinutes)
	}
	if details.MaterialTransactionID != "" {
		fmt.Fprintf(out, "Material Transaction ID: %s\n", details.MaterialTransactionID)
	}
	if details.MaterialTransactionTicketNumber != "" {
		fmt.Fprintf(out, "Material Transaction Ticket: %s\n", details.MaterialTransactionTicketNumber)
	}
	if details.MaterialTransactionAt != "" {
		fmt.Fprintf(out, "Material Transaction At: %s\n", details.MaterialTransactionAt)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer ID: %s\n", details.TrailerID)
	}
	if details.TrailerNumber != "" {
		fmt.Fprintf(out, "Trailer Number: %s\n", details.TrailerNumber)
	}

	return nil
}
