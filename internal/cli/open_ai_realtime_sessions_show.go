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

type openAiRealtimeSessionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type openAiRealtimeSessionDetails struct {
	ID                    string `json:"id"`
	Model                 string `json:"model,omitempty"`
	ClientFeature         string `json:"client_feature,omitempty"`
	ClientSecretValue     string `json:"client_secret_value,omitempty"`
	ClientSecretExpiresAt string `json:"client_secret_expires_at,omitempty"`
	Error                 string `json:"error,omitempty"`
	UserID                string `json:"user_id,omitempty"`
	UserName              string `json:"user_name,omitempty"`
	UserEmail             string `json:"user_email,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newOpenAiRealtimeSessionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show OpenAI realtime session details",
		Long: `Show the full details of a specific OpenAI realtime session.

Requires admin privileges.

Output Fields:
  ID                      Session identifier
  Model                   OpenAI model
  Client Feature          Client feature enum
  User                    User name, email, and ID (if set)
  Client Secret           Session client secret (if available)
  Client Secret Expires   Secret expiration timestamp
  Error                   Error message (if any)
  Created At              Creation timestamp
  Updated At              Last update timestamp

Arguments:
  <id>    The session ID (required). You can find IDs using the list command.`,
		Example: `  # View a session by ID
  xbe view open-ai-realtime-sessions show 123

  # Get JSON output
  xbe view open-ai-realtime-sessions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOpenAiRealtimeSessionsShow,
	}
	initOpenAiRealtimeSessionsShowFlags(cmd)
	return cmd
}

func init() {
	openAiRealtimeSessionsCmd.AddCommand(newOpenAiRealtimeSessionsShowCmd())
}

func initOpenAiRealtimeSessionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenAiRealtimeSessionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOpenAiRealtimeSessionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("open ai realtime session id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[open-ai-realtime-sessions]", "model,client-feature,client-secret-value,client-secret-expires-at,error,created-at,updated-at,user")
	query.Set("fields[users]", "name,email-address")
	query.Set("include", "user")

	body, _, err := client.Get(cmd.Context(), "/v1/open-ai-realtime-sessions/"+id, query)
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

	details := buildOpenAiRealtimeSessionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOpenAiRealtimeSessionDetails(cmd, details)
}

func parseOpenAiRealtimeSessionsShowOptions(cmd *cobra.Command) (openAiRealtimeSessionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openAiRealtimeSessionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOpenAiRealtimeSessionDetails(resp jsonAPISingleResponse) openAiRealtimeSessionDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := openAiRealtimeSessionDetails{
		ID:                    resp.Data.ID,
		Model:                 strings.TrimSpace(stringAttr(attrs, "model")),
		ClientFeature:         strings.TrimSpace(stringAttr(attrs, "client-feature")),
		ClientSecretValue:     strings.TrimSpace(stringAttr(attrs, "client-secret-value")),
		ClientSecretExpiresAt: formatDateTime(stringAttr(attrs, "client-secret-expires-at")),
		Error:                 strings.TrimSpace(stringAttr(attrs, "error")),
		CreatedAt:             formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:             formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			details.UserEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
		}
	}

	return details
}

func renderOpenAiRealtimeSessionDetails(cmd *cobra.Command, details openAiRealtimeSessionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Model != "" {
		fmt.Fprintf(out, "Model: %s\n", details.Model)
	}
	if details.ClientFeature != "" {
		fmt.Fprintf(out, "Client Feature: %s\n", details.ClientFeature)
	}
	if details.UserID != "" {
		userLabel := details.UserID
		if details.UserName != "" && details.UserEmail != "" {
			userLabel = fmt.Sprintf("%s <%s> (%s)", details.UserName, details.UserEmail, details.UserID)
		} else if details.UserName != "" {
			userLabel = fmt.Sprintf("%s (%s)", details.UserName, details.UserID)
		}
		fmt.Fprintf(out, "User: %s\n", userLabel)
	}
	if details.ClientSecretValue != "" {
		fmt.Fprintf(out, "Client Secret: %s\n", details.ClientSecretValue)
	}
	if details.ClientSecretExpiresAt != "" {
		fmt.Fprintf(out, "Client Secret Expires: %s\n", details.ClientSecretExpiresAt)
	}
	if details.Error != "" {
		fmt.Fprintf(out, "Error: %s\n", details.Error)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
