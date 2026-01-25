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

type truckerBrokeragesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type truckerBrokerageDetails struct {
	ID                string `json:"id"`
	TruckerID         string `json:"trucker_id,omitempty"`
	BrokeredTruckerID string `json:"brokered_trucker_id,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

func newTruckerBrokeragesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trucker brokerage details",
		Long: `Show the full details of a trucker brokerage.

Output Fields:
  ID
  Trucker ID
  Brokered Trucker ID
  Created At
  Updated At

Arguments:
  <id>    The trucker brokerage ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a trucker brokerage
  xbe view trucker-brokerages show 123

  # Get JSON output
  xbe view trucker-brokerages show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTruckerBrokeragesShow,
	}
	initTruckerBrokeragesShowFlags(cmd)
	return cmd
}

func init() {
	truckerBrokeragesCmd.AddCommand(newTruckerBrokeragesShowCmd())
}

func initTruckerBrokeragesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerBrokeragesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTruckerBrokeragesShowOptions(cmd)
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
		return fmt.Errorf("trucker brokerage id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-brokerages/"+id, nil)
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

	details := buildTruckerBrokerageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTruckerBrokerageDetails(cmd, details)
}

func parseTruckerBrokeragesShowOptions(cmd *cobra.Command) (truckerBrokeragesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerBrokeragesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTruckerBrokerageDetails(resp jsonAPISingleResponse) truckerBrokerageDetails {
	attrs := resp.Data.Attributes
	details := truckerBrokerageDetails{
		ID:        resp.Data.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["brokered-trucker"]; ok && rel.Data != nil {
		details.BrokeredTruckerID = rel.Data.ID
	}

	return details
}

func renderTruckerBrokerageDetails(cmd *cobra.Command, details truckerBrokerageDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.BrokeredTruckerID != "" {
		fmt.Fprintf(out, "Brokered Trucker ID: %s\n", details.BrokeredTruckerID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
