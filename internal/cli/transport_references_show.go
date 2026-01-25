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

type transportReferencesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type transportReferenceDetails struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Position    int    `json:"position"`
	SubjectType string `json:"subject_type,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

func newTransportReferencesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show transport reference details",
		Long: `Show the full details of a specific transport reference.

Output Fields:
  ID         Transport reference identifier
  Key        Reference key
  Value      Reference value
  Position   Reference position within the subject
  Subject    Subject type and ID
  Created At Creation timestamp
  Updated At Last update timestamp

Arguments:
  <id>  The transport reference ID (required). Use the list command to find IDs.`,
		Example: `  # View a transport reference by ID
  xbe view transport-references show 123

  # Get transport reference as JSON
  xbe view transport-references show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTransportReferencesShow,
	}
	initTransportReferencesShowFlags(cmd)
	return cmd
}

func init() {
	transportReferencesCmd.AddCommand(newTransportReferencesShowCmd())
}

func initTransportReferencesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportReferencesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTransportReferencesShowOptions(cmd)
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
		return fmt.Errorf("transport reference id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[transport-references]", "key,value,position,created-at,updated-at,subject")
	query.Set("include", "subject")

	body, _, err := client.Get(cmd.Context(), "/v1/transport-references/"+id, query)
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

	details := buildTransportReferenceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTransportReferenceDetails(cmd, details)
}

func parseTransportReferencesShowOptions(cmd *cobra.Command) (transportReferencesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return transportReferencesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return transportReferencesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return transportReferencesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return transportReferencesShowOptions{}, err
	}

	return transportReferencesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTransportReferenceDetails(resp jsonAPISingleResponse) transportReferenceDetails {
	attrs := resp.Data.Attributes

	details := transportReferenceDetails{
		ID:        resp.Data.ID,
		Key:       stringAttr(attrs, "key"),
		Value:     stringAttr(attrs, "value"),
		Position:  intAttr(attrs, "position"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
		details.SubjectType = rel.Data.Type
		details.SubjectID = rel.Data.ID
	}

	return details
}

func renderTransportReferenceDetails(cmd *cobra.Command, details transportReferenceDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Key != "" {
		fmt.Fprintf(out, "Key: %s\n", details.Key)
	}
	if details.Value != "" {
		fmt.Fprintf(out, "Value: %s\n", details.Value)
	}
	fmt.Fprintf(out, "Position: %d\n", details.Position)
	if details.SubjectType != "" && details.SubjectID != "" {
		fmt.Fprintf(out, "Subject: %s/%s\n", details.SubjectType, details.SubjectID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
