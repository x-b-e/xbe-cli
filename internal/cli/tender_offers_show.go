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

type tenderOffersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderOfferDetails struct {
	ID                          string `json:"id"`
	TenderType                  string `json:"tender_type,omitempty"`
	TenderID                    string `json:"tender_id,omitempty"`
	Comment                     string `json:"comment,omitempty"`
	SkipCertificationValidation bool   `json:"skip_certification_validation"`
}

func newTenderOffersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender offer details",
		Long: `Show full details of a tender offer.

Output Fields:
  ID           Offer identifier
  Tender Type  Tender type (broker-tenders, customer-tenders)
  Tender ID    Tender ID
  Comment      Comment (if provided)
  Skip Cert    Skip certification validation (true/false)

Arguments:
  <id>    Tender offer ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a tender offer
  xbe view tender-offers show 123

  # JSON output
  xbe view tender-offers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderOffersShow,
	}
	initTenderOffersShowFlags(cmd)
	return cmd
}

func init() {
	tenderOffersCmd.AddCommand(newTenderOffersShowCmd())
}

func initTenderOffersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderOffersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTenderOffersShowOptions(cmd)
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
		return fmt.Errorf("tender offer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tender-offers]", "tender,comment,skip-certification-validation")

	body, status, err := client.Get(cmd.Context(), "/v1/tender-offers/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTenderOffersShowUnavailable(cmd, opts.JSON)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTenderOfferDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderOfferDetails(cmd, details)
}

func renderTenderOffersShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), tenderOfferDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Tender offers are write-only; show is not available.")
	return nil
}

func parseTenderOffersShowOptions(cmd *cobra.Command) (tenderOffersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderOffersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderOfferDetails(resp jsonAPISingleResponse) tenderOfferDetails {
	attrs := resp.Data.Attributes
	details := tenderOfferDetails{
		ID:                          resp.Data.ID,
		Comment:                     strings.TrimSpace(stringAttr(attrs, "comment")),
		SkipCertificationValidation: boolAttr(attrs, "skip-certification-validation"),
	}

	if rel, ok := resp.Data.Relationships["tender"]; ok && rel.Data != nil {
		details.TenderType = rel.Data.Type
		details.TenderID = rel.Data.ID
	}

	return details
}

func renderTenderOfferDetails(cmd *cobra.Command, details tenderOfferDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderID != "" {
		fmt.Fprintf(out, "Tender Type: %s\n", details.TenderType)
		fmt.Fprintf(out, "Tender ID: %s\n", details.TenderID)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))
	fmt.Fprintf(out, "Skip Certification Validation: %t\n", details.SkipCertificationValidation)

	return nil
}
