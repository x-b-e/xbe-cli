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

type doTransportReferencesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	SubjectType string
	SubjectID   string
	Key         string
	Value       string
	Position    int
}

func newDoTransportReferencesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a transport reference",
		Long: `Create a transport reference.

Required:
  --subject-type  Subject type (e.g., transport-orders)
  --subject-id    Subject ID
  --key           Reference key
  --value         Reference value

Optional:
  --position      Reference position within the subject`,
		Example: `  # Create a transport reference for a transport order
  xbe do transport-references create --subject-type transport-orders --subject-id 123 --key BOL --value "BOL-999"

  # Create with an explicit position
  xbe do transport-references create --subject-type transport-orders --subject-id 123 --key PO --value "PO-123" --position 1`,
		RunE: runDoTransportReferencesCreate,
	}
	initDoTransportReferencesCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportReferencesCmd.AddCommand(newDoTransportReferencesCreateCmd())
}

func initDoTransportReferencesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("subject-type", "", "Subject type (e.g., transport-orders)")
	cmd.Flags().String("subject-id", "", "Subject ID")
	cmd.Flags().String("key", "", "Reference key")
	cmd.Flags().String("value", "", "Reference value")
	cmd.Flags().Int("position", 0, "Reference position within the subject")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("subject-type")
	_ = cmd.MarkFlagRequired("subject-id")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("value")
}

func runDoTransportReferencesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTransportReferencesCreateOptions(cmd)
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

	attributes := map[string]any{
		"key":   opts.Key,
		"value": opts.Value,
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}

	relationships := map[string]any{
		"subject": map[string]any{
			"data": map[string]any{
				"type": opts.SubjectType,
				"id":   opts.SubjectID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "transport-references",
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

	body, _, err := client.Post(cmd.Context(), "/v1/transport-references", jsonBody)
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

	if opts.JSON {
		row := buildTransportReferenceRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created transport reference %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportReferencesCreateOptions(cmd *cobra.Command) (doTransportReferencesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	key, _ := cmd.Flags().GetString("key")
	value, _ := cmd.Flags().GetString("value")
	position, _ := cmd.Flags().GetInt("position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportReferencesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Key:         key,
		Value:       value,
		Position:    position,
	}, nil
}
