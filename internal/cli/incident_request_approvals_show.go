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

type incidentRequestApprovalsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentRequestApprovalDetails struct {
	ID                string `json:"id"`
	IncidentRequestID string `json:"incident_request_id,omitempty"`
	Comment           string `json:"comment,omitempty"`
}

func newIncidentRequestApprovalsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident request approval details",
		Long: `Show full details of an incident request approval.

Output Fields:
  ID                Approval identifier
  Incident Request  Incident request ID
  Comment           Comment (if provided)

Arguments:
  <id>    Incident request approval ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an incident request approval
  xbe view incident-request-approvals show 123

  # JSON output
  xbe view incident-request-approvals show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentRequestApprovalsShow,
	}
	initIncidentRequestApprovalsShowFlags(cmd)
	return cmd
}

func init() {
	incidentRequestApprovalsCmd.AddCommand(newIncidentRequestApprovalsShowCmd())
}

func initIncidentRequestApprovalsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentRequestApprovalsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseIncidentRequestApprovalsShowOptions(cmd)
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
		return fmt.Errorf("incident request approval id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-request-approvals]", "incident-request,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/incident-request-approvals/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderIncidentRequestApprovalsShowUnavailable(cmd, opts.JSON)
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

	details := buildIncidentRequestApprovalDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIncidentRequestApprovalDetails(cmd, details)
}

func renderIncidentRequestApprovalsShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), incidentRequestApprovalDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Incident request approvals are write-only; show is not available.")
	return nil
}

func parseIncidentRequestApprovalsShowOptions(cmd *cobra.Command) (incidentRequestApprovalsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentRequestApprovalsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentRequestApprovalDetails(resp jsonAPISingleResponse) incidentRequestApprovalDetails {
	attrs := resp.Data.Attributes
	details := incidentRequestApprovalDetails{
		ID:      resp.Data.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["incident-request"]; ok && rel.Data != nil {
		details.IncidentRequestID = rel.Data.ID
	}

	return details
}

func renderIncidentRequestApprovalDetails(cmd *cobra.Command, details incidentRequestApprovalDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.IncidentRequestID != "" {
		fmt.Fprintf(out, "Incident Request: %s\n", details.IncidentRequestID)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
