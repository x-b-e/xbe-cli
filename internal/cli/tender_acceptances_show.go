package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type tenderAcceptancesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderAcceptanceDetails struct {
	ID                                string   `json:"id"`
	TenderType                        string   `json:"tender_type,omitempty"`
	TenderID                          string   `json:"tender_id,omitempty"`
	Comment                           string   `json:"comment,omitempty"`
	SkipCertificationValidation       bool     `json:"skip_certification_validation"`
	RejectedTenderJobScheduleShiftIDs []string `json:"rejected_tender_job_schedule_shift_ids,omitempty"`
}

func newTenderAcceptancesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender acceptance details",
		Long: `Show full details of a tender acceptance.

Output Fields:
  ID               Acceptance identifier
  Tender Type      Tender type (broker-tenders, customer-tenders)
  Tender ID        Tender ID
  Comment          Comment (if provided)
  Skip Cert        Skip certification validation (true/false)
  Rejected Shifts  Rejected tender job schedule shift IDs

Arguments:
  <id>    Tender acceptance ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a tender acceptance
  xbe view tender-acceptances show 123

  # JSON output
  xbe view tender-acceptances show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderAcceptancesShow,
	}
	initTenderAcceptancesShowFlags(cmd)
	return cmd
}

func init() {
	tenderAcceptancesCmd.AddCommand(newTenderAcceptancesShowCmd())
}

func initTenderAcceptancesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderAcceptancesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTenderAcceptancesShowOptions(cmd)
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
		return fmt.Errorf("tender acceptance id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tender-acceptances]", "tender,comment,skip-certification-validation,rejected-tender-job-schedule-shift-ids")

	body, status, err := client.Get(cmd.Context(), "/v1/tender-acceptances/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTenderAcceptancesShowUnavailable(cmd, opts.JSON)
		}
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

	details := buildTenderAcceptanceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderAcceptanceDetails(cmd, details)
}

func renderTenderAcceptancesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), tenderAcceptanceDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Tender acceptances are write-only; show is not available.")
	return nil
}

func parseTenderAcceptancesShowOptions(cmd *cobra.Command) (tenderAcceptancesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderAcceptancesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderAcceptanceDetails(resp jsonAPISingleResponse) tenderAcceptanceDetails {
	attrs := resp.Data.Attributes
	details := tenderAcceptanceDetails{
		ID:                                resp.Data.ID,
		Comment:                           strings.TrimSpace(stringAttr(attrs, "comment")),
		SkipCertificationValidation:       boolAttr(attrs, "skip-certification-validation"),
		RejectedTenderJobScheduleShiftIDs: stringSliceAttr(attrs, "rejected-tender-job-schedule-shift-ids"),
	}

	if rel, ok := resp.Data.Relationships["tender"]; ok && rel.Data != nil {
		details.TenderType = rel.Data.Type
		details.TenderID = rel.Data.ID
	}

	return details
}

func renderTenderAcceptanceDetails(cmd *cobra.Command, details tenderAcceptanceDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderID != "" {
		fmt.Fprintf(out, "Tender Type: %s\n", details.TenderType)
		fmt.Fprintf(out, "Tender ID: %s\n", details.TenderID)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))
	fmt.Fprintf(out, "Skip Certification Validation: %t\n", details.SkipCertificationValidation)
	rejected := strings.Join(details.RejectedTenderJobScheduleShiftIDs, ", ")
	fmt.Fprintf(out, "Rejected Tender Job Schedule Shift IDs: %s\n", formatOptional(rejected))

	return nil
}
