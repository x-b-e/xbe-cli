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

type doIncidentRequestCancellationsCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	IncidentRequest string
	Comment         string
}

type incidentRequestCancellationRow struct {
	ID                string `json:"id"`
	IncidentRequestID string `json:"incident_request_id,omitempty"`
	Comment           string `json:"comment,omitempty"`
}

func newDoIncidentRequestCancellationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Cancel an incident request",
		Long: `Cancel an incident request.

Required flags:
  --incident-request  Incident request ID (required)

Optional flags:
  --comment  Cancellation comment`,
		Example: `  # Cancel an incident request
  xbe do incident-request-cancellations create --incident-request 123 --comment "No longer needed"`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentRequestCancellationsCreate,
	}
	initDoIncidentRequestCancellationsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentRequestCancellationsCmd.AddCommand(newDoIncidentRequestCancellationsCreateCmd())
}

func initDoIncidentRequestCancellationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("incident-request", "", "Incident request ID (required)")
	cmd.Flags().String("comment", "", "Cancellation comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentRequestCancellationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentRequestCancellationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "incident-request-cancellations",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/incident-request-cancellations", jsonBody)
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

	row := incidentRequestCancellationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident request cancellation %s\n", row.ID)
	return nil
}

func incidentRequestCancellationRowFromSingle(resp jsonAPISingleResponse) incidentRequestCancellationRow {
	attrs := resp.Data.Attributes
	row := incidentRequestCancellationRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resp.Data.Relationships["incident-request"]; ok && rel.Data != nil {
		row.IncidentRequestID = rel.Data.ID
	}

	return row
}

func parseDoIncidentRequestCancellationsCreateOptions(cmd *cobra.Command) (doIncidentRequestCancellationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	incidentRequest, _ := cmd.Flags().GetString("incident-request")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentRequestCancellationsCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		IncidentRequest: incidentRequest,
		Comment:         comment,
	}, nil
}
