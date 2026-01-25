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

type doIncidentRequestApprovalsCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	IncidentRequest string
	Comment         string
}

func newDoIncidentRequestApprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Approve an incident request",
		Long: `Approve a submitted incident request.

Required flags:
  --incident-request   Incident request ID (required)

Optional flags:
  --comment            Approval comment`,
		Example: `  # Approve an incident request
  xbe do incident-request-approvals create --incident-request 12345

  # Approve with a comment
  xbe do incident-request-approvals create --incident-request 12345 --comment "Approved"

  # JSON output
  xbe do incident-request-approvals create --incident-request 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentRequestApprovalsCreate,
	}
	initDoIncidentRequestApprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentRequestApprovalsCmd.AddCommand(newDoIncidentRequestApprovalsCreateCmd())
}

func initDoIncidentRequestApprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("incident-request", "", "Incident request ID (required)")
	cmd.Flags().String("comment", "", "Approval comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentRequestApprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentRequestApprovalsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.IncidentRequest) == "" {
		err := fmt.Errorf("--incident-request is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"incident-request": map[string]any{
			"data": map[string]any{
				"type": "incident-requests",
				"id":   opts.IncidentRequest,
			},
		},
	}

	data := map[string]any{
		"type":          "incident-request-approvals",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/incident-request-approvals", jsonBody)
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

	row := buildIncidentRequestApprovalRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.IncidentRequestID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created incident request approval %s for incident request %s\n", row.ID, row.IncidentRequestID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident request approval %s\n", row.ID)
	return nil
}

func parseDoIncidentRequestApprovalsCreateOptions(cmd *cobra.Command) (doIncidentRequestApprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	incidentRequest, _ := cmd.Flags().GetString("incident-request")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentRequestApprovalsCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		IncidentRequest: incidentRequest,
		Comment:         comment,
	}, nil
}
