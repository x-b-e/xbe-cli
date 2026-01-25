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

type doOpenAiRealtimeSessionsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ClientFeature string
	Model         string
	User          string
}

type openAiRealtimeSessionCreateResult struct {
	ID                    string `json:"id"`
	Model                 string `json:"model,omitempty"`
	ClientFeature         string `json:"client_feature,omitempty"`
	ClientSecretValue     string `json:"client_secret_value,omitempty"`
	ClientSecretExpiresAt string `json:"client_secret_expires_at,omitempty"`
	Error                 string `json:"error,omitempty"`
	UserID                string `json:"user_id,omitempty"`
}

func newDoOpenAiRealtimeSessionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an OpenAI realtime session",
		Long: `Create an OpenAI realtime session.

Required flags:
  --client-feature   Client feature enum (required)

Optional flags:
  --model            OpenAI realtime model (defaults to server default)
  --user             User ID to associate with the session

The response includes a short-lived client secret for initiating a realtime
session with OpenAI.`,
		Example: `  # Create a session with the default model
  xbe do open-ai-realtime-sessions create --client-feature giant_anchor_prediction_creation

  # Create a session for a specific user and model
  xbe do open-ai-realtime-sessions create --client-feature slack_realtime_chat --model gpt-4o-realtime-preview --user 123

  # Output JSON
  xbe do open-ai-realtime-sessions create --client-feature development --json`,
		Args: cobra.NoArgs,
		RunE: runDoOpenAiRealtimeSessionsCreate,
	}
	initDoOpenAiRealtimeSessionsCreateFlags(cmd)
	return cmd
}

func init() {
	doOpenAiRealtimeSessionsCmd.AddCommand(newDoOpenAiRealtimeSessionsCreateCmd())
}

func initDoOpenAiRealtimeSessionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("client-feature", "", "Client feature enum (required)")
	cmd.Flags().String("model", "", "OpenAI realtime model (optional)")
	cmd.Flags().String("user", "", "User ID to associate with the session")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("client-feature")
}

func runDoOpenAiRealtimeSessionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOpenAiRealtimeSessionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ClientFeature) == "" {
		err := fmt.Errorf("--client-feature is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"client-feature": opts.ClientFeature,
	}
	if strings.TrimSpace(opts.Model) != "" {
		attributes["model"] = opts.Model
	}

	data := map[string]any{
		"type":       "open-ai-realtime-sessions",
		"attributes": attributes,
	}

	if strings.TrimSpace(opts.User) != "" {
		data["relationships"] = map[string]any{
			"user": map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.User,
				},
			},
		}
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/open-ai-realtime-sessions", jsonBody)
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

	result := buildOpenAiRealtimeSessionCreateResult(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created OpenAI realtime session %s\n", result.ID)
	if result.ClientSecretValue != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Client Secret: %s\n", result.ClientSecretValue)
	}
	if result.ClientSecretExpiresAt != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Expires At: %s\n", result.ClientSecretExpiresAt)
	}
	return nil
}

func parseDoOpenAiRealtimeSessionsCreateOptions(cmd *cobra.Command) (doOpenAiRealtimeSessionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	clientFeature, _ := cmd.Flags().GetString("client-feature")
	model, _ := cmd.Flags().GetString("model")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenAiRealtimeSessionsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ClientFeature: clientFeature,
		Model:         model,
		User:          user,
	}, nil
}

func buildOpenAiRealtimeSessionCreateResult(resp jsonAPISingleResponse) openAiRealtimeSessionCreateResult {
	attrs := resp.Data.Attributes

	result := openAiRealtimeSessionCreateResult{
		ID:                    resp.Data.ID,
		Model:                 strings.TrimSpace(stringAttr(attrs, "model")),
		ClientFeature:         strings.TrimSpace(stringAttr(attrs, "client-feature")),
		ClientSecretValue:     strings.TrimSpace(stringAttr(attrs, "client-secret-value")),
		ClientSecretExpiresAt: formatDateTime(stringAttr(attrs, "client-secret-expires-at")),
		Error:                 strings.TrimSpace(stringAttr(attrs, "error")),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		result.UserID = rel.Data.ID
	}

	return result
}
