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

type doTimeCardTimeChangesCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	TimeCard                string
	CreatedBy               string
	TimeChangesAttributes   string
	Comment                 string
	SkipTimeCardNotEditable string
	SkipQuantityValidation  string
}

func newDoTimeCardTimeChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time card time change",
		Long: `Create a time card time change.

Required flags:
  --time-card               Time card ID (required)
  --time-changes-attributes Time change attributes as JSON object (required)

Optional flags:
  --comment                     Comment describing the change
  --created-by                  Created by user ID (defaults to current user)
  --skip-time-card-not-editable Skip editable validation (true/false)
  --skip-quantity-validation    Skip quantity validation (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a time card time change
  xbe do time-card-time-changes create \
    --time-card 123 \
    --time-changes-attributes '{"down_minutes":15}'

  # Create with a comment
  xbe do time-card-time-changes create \
    --time-card 123 \
    --time-changes-attributes '{"start_at":"2026-01-23T12:00:00Z"}' \
    --comment "Adjusted start time"`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardTimeChangesCreate,
	}
	initDoTimeCardTimeChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardTimeChangesCmd.AddCommand(newDoTimeCardTimeChangesCreateCmd())
}

func initDoTimeCardTimeChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID (required)")
	cmd.Flags().String("time-changes-attributes", "", "Time change attributes as JSON object (required)")
	cmd.Flags().String("comment", "", "Comment describing the change")
	cmd.Flags().String("created-by", "", "Created by user ID (optional)")
	cmd.Flags().String("skip-time-card-not-editable", "", "Skip editable validation (true/false)")
	cmd.Flags().String("skip-quantity-validation", "", "Skip quantity validation (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardTimeChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardTimeChangesCreateOptions(cmd)
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

	timeCardID := strings.TrimSpace(opts.TimeCard)
	if timeCardID == "" {
		err := fmt.Errorf("--time-card is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TimeChangesAttributes) == "" {
		err := fmt.Errorf("--time-changes-attributes is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var timeChanges map[string]any
	if err := json.Unmarshal([]byte(opts.TimeChangesAttributes), &timeChanges); err != nil {
		err := fmt.Errorf("invalid time-changes-attributes JSON: %w", err)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if len(timeChanges) == 0 {
		err := fmt.Errorf("--time-changes-attributes must include at least one change")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"time-changes-attributes": timeChanges,
	}
	setStringAttrIfPresent(attributes, "comment", opts.Comment)
	setBoolAttrIfPresent(attributes, "skip-time-card-not-editable", opts.SkipTimeCardNotEditable)
	setBoolAttrIfPresent(attributes, "skip-quantity-validation", opts.SkipQuantityValidation)

	relationships := map[string]any{
		"time-card": map[string]any{
			"data": map[string]any{
				"type": "time-cards",
				"id":   timeCardID,
			},
		},
	}
	if createdBy := strings.TrimSpace(opts.CreatedBy); createdBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   createdBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-card-time-changes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-time-changes", jsonBody)
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

	row := buildTimeCardTimeChangeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card time change %s\n", row.ID)
	return nil
}

func parseDoTimeCardTimeChangesCreateOptions(cmd *cobra.Command) (doTimeCardTimeChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCard, _ := cmd.Flags().GetString("time-card")
	createdBy, _ := cmd.Flags().GetString("created-by")
	timeChangesAttributes, _ := cmd.Flags().GetString("time-changes-attributes")
	comment, _ := cmd.Flags().GetString("comment")
	skipTimeCardNotEditable, _ := cmd.Flags().GetString("skip-time-card-not-editable")
	skipQuantityValidation, _ := cmd.Flags().GetString("skip-quantity-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardTimeChangesCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		TimeCard:                timeCard,
		CreatedBy:               createdBy,
		TimeChangesAttributes:   timeChangesAttributes,
		Comment:                 comment,
		SkipTimeCardNotEditable: skipTimeCardNotEditable,
		SkipQuantityValidation:  skipQuantityValidation,
	}, nil
}

func buildTimeCardTimeChangeRowFromSingle(resp jsonAPISingleResponse) timeCardTimeChangeRow {
	resource := resp.Data
	attrs := resource.Attributes
	return timeCardTimeChangeRow{
		ID:          resource.ID,
		TimeCardID:  relationshipIDFromMap(resource.Relationships, "time-card"),
		CreatedByID: relationshipIDFromMap(resource.Relationships, "created-by"),
		IsProcessed: boolAttr(attrs, "is-processed"),
		Comment:     stringAttr(attrs, "comment"),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
	}
}
