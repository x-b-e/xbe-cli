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

type retainerEarningStatusesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type retainerEarningStatusDetails struct {
	ID           string `json:"id"`
	RetainerID   string `json:"retainer_id,omitempty"`
	CalculatedOn string `json:"calculated_on,omitempty"`
	Expected     string `json:"expected,omitempty"`
	Actual       string `json:"actual,omitempty"`
}

func newRetainerEarningStatusesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show retainer earning status details",
		Long: `Show the full details of a retainer earning status.

Output Fields:
  ID
  Retainer ID
  Calculated On
  Expected
  Actual

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The retainer earning status ID (required). You can find IDs using the list command.`,
		Example: `  # Show a retainer earning status
  xbe view retainer-earning-statuses show 123

  # Get JSON output
  xbe view retainer-earning-statuses show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRetainerEarningStatusesShow,
	}
	initRetainerEarningStatusesShowFlags(cmd)
	return cmd
}

func init() {
	retainerEarningStatusesCmd.AddCommand(newRetainerEarningStatusesShowCmd())
}

func initRetainerEarningStatusesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerEarningStatusesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRetainerEarningStatusesShowOptions(cmd)
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
		return fmt.Errorf("retainer earning status id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[retainer-earning-statuses]", "expected,actual,calculated-on,retainer")

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-earning-statuses/"+id, query)
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

	details := buildRetainerEarningStatusDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRetainerEarningStatusDetails(cmd, details)
}

func parseRetainerEarningStatusesShowOptions(cmd *cobra.Command) (retainerEarningStatusesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerEarningStatusesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRetainerEarningStatusDetails(resp jsonAPISingleResponse) retainerEarningStatusDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := retainerEarningStatusDetails{
		ID:           resource.ID,
		CalculatedOn: formatDate(stringAttr(attrs, "calculated-on")),
		Expected:     stringAttr(attrs, "expected"),
		Actual:       stringAttr(attrs, "actual"),
	}

	if rel, ok := resource.Relationships["retainer"]; ok && rel.Data != nil {
		details.RetainerID = rel.Data.ID
	}

	return details
}

func renderRetainerEarningStatusDetails(cmd *cobra.Command, details retainerEarningStatusDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RetainerID != "" {
		fmt.Fprintf(out, "Retainer ID: %s\n", details.RetainerID)
	}
	if details.CalculatedOn != "" {
		fmt.Fprintf(out, "Calculated On: %s\n", details.CalculatedOn)
	}
	if details.Expected != "" {
		fmt.Fprintf(out, "Expected: %s\n", details.Expected)
	}
	if details.Actual != "" {
		fmt.Fprintf(out, "Actual: %s\n", details.Actual)
	}

	return nil
}
