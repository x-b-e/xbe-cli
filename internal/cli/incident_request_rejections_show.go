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

type incidentRequestRejectionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentRequestRejectionDetails struct {
	ID                string `json:"id"`
	IncidentRequestID string `json:"incident_request_id,omitempty"`
	Comment           string `json:"comment,omitempty"`
}

func newIncidentRequestRejectionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident request rejection details",
		Long: `Show full details of an incident request rejection.

Output Fields:
  ID                Rejection identifier
  Incident Request  Incident request ID
  Comment           Comment (if provided)

Arguments:
  <id>    Incident request rejection ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an incident request rejection
  xbe view incident-request-rejections show 123

  # JSON output
  xbe view incident-request-rejections show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentRequestRejectionsShow,
	}
	initIncidentRequestRejectionsShowFlags(cmd)
	return cmd
}

func init() {
	incidentRequestRejectionsCmd.AddCommand(newIncidentRequestRejectionsShowCmd())
}

func initIncidentRequestRejectionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentRequestRejectionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseIncidentRequestRejectionsShowOptions(cmd)
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
		return fmt.Errorf("incident request rejection id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-request-rejections]", "incident-request,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/incident-request-rejections/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderIncidentRequestRejectionsShowUnavailable(cmd, opts.JSON)
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

	details := buildIncidentRequestRejectionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIncidentRequestRejectionDetails(cmd, details)
}

func renderIncidentRequestRejectionsShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), incidentRequestRejectionDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Incident request rejections are write-only; show is not available.")
	return nil
}

func parseIncidentRequestRejectionsShowOptions(cmd *cobra.Command) (incidentRequestRejectionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentRequestRejectionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentRequestRejectionDetails(resp jsonAPISingleResponse) incidentRequestRejectionDetails {
	attrs := resp.Data.Attributes
	details := incidentRequestRejectionDetails{
		ID:      resp.Data.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["incident-request"]; ok && rel.Data != nil {
		details.IncidentRequestID = rel.Data.ID
	}

	return details
}

func renderIncidentRequestRejectionDetails(cmd *cobra.Command, details incidentRequestRejectionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.IncidentRequestID != "" {
		fmt.Fprintf(out, "Incident Request: %s\n", details.IncidentRequestID)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
