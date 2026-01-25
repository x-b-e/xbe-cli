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

type doUiTourStepsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Name         string
	Content      string
	Abbreviation string
	Sequence     int
	UiTourID     string
}

func newDoUiTourStepsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UI tour step",
		Long: `Create a UI tour step.

Required flags:
  --name           Step name (required)
  --content        Step content (required)
  --abbreviation   Step abbreviation (required)
  --ui-tour        UI tour ID (required)

Optional flags:
  --sequence       Step sequence order`,
		Example: `  # Create a UI tour step
  xbe do ui-tour-steps create \
    --name "Welcome" \
    --content "Welcome to the app" \
    --abbreviation "welcome" \
    --ui-tour 123

  # Create with sequence
  xbe do ui-tour-steps create \
    --name "Next" \
    --content "Next step" \
    --abbreviation "next" \
    --ui-tour 123 \
    --sequence 2`,
		Args: cobra.NoArgs,
		RunE: runDoUiTourStepsCreate,
	}
	initDoUiTourStepsCreateFlags(cmd)
	return cmd
}

func init() {
	doUiTourStepsCmd.AddCommand(newDoUiTourStepsCreateCmd())
}

func initDoUiTourStepsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Step name (required)")
	cmd.Flags().String("content", "", "Step content (required)")
	cmd.Flags().String("abbreviation", "", "Step abbreviation (required)")
	cmd.Flags().Int("sequence", 0, "Step sequence order")
	cmd.Flags().String("ui-tour", "", "UI tour ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUiTourStepsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUiTourStepsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Name) == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Content) == "" {
		err := fmt.Errorf("--content is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Abbreviation) == "" {
		err := fmt.Errorf("--abbreviation is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.UiTourID) == "" {
		err := fmt.Errorf("--ui-tour is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name":         opts.Name,
		"content":      opts.Content,
		"abbreviation": opts.Abbreviation,
	}
	if cmd.Flags().Changed("sequence") {
		attributes["sequence"] = opts.Sequence
	}

	relationships := map[string]any{
		"ui-tour": map[string]any{
			"data": map[string]any{
				"type": "ui-tours",
				"id":   opts.UiTourID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "ui-tour-steps",
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

	body, _, err := client.Post(cmd.Context(), "/v1/ui-tour-steps", jsonBody)
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

	row := uiTourStepRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created UI tour step %s\n", row.ID)
	return nil
}

func parseDoUiTourStepsCreateOptions(cmd *cobra.Command) (doUiTourStepsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	content, _ := cmd.Flags().GetString("content")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	sequence, _ := cmd.Flags().GetInt("sequence")
	uiTourID, _ := cmd.Flags().GetString("ui-tour")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUiTourStepsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Name:         name,
		Content:      content,
		Abbreviation: abbreviation,
		Sequence:     sequence,
		UiTourID:     uiTourID,
	}, nil
}
