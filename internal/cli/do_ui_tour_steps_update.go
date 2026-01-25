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

type doUiTourStepsUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Name         string
	Content      string
	Abbreviation string
	Sequence     int
}

func newDoUiTourStepsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a UI tour step",
		Long: `Update a UI tour step.

Optional flags:
  --name           Step name
  --content        Step content
  --abbreviation   Step abbreviation
  --sequence       Step sequence order`,
		Example: `  # Update a UI tour step name
  xbe do ui-tour-steps update 123 --name "Updated step"

  # Update sequence
  xbe do ui-tour-steps update 123 --sequence 3`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUiTourStepsUpdate,
	}
	initDoUiTourStepsUpdateFlags(cmd)
	return cmd
}

func init() {
	doUiTourStepsCmd.AddCommand(newDoUiTourStepsUpdateCmd())
}

func initDoUiTourStepsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Step name")
	cmd.Flags().String("content", "", "Step content")
	cmd.Flags().String("abbreviation", "", "Step abbreviation")
	cmd.Flags().Int("sequence", 0, "Step sequence order")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUiTourStepsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUiTourStepsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("content") {
		attributes["content"] = opts.Content
	}
	if cmd.Flags().Changed("abbreviation") {
		attributes["abbreviation"] = opts.Abbreviation
	}
	if cmd.Flags().Changed("sequence") {
		attributes["sequence"] = opts.Sequence
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "ui-tour-steps",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/ui-tour-steps/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated UI tour step %s\n", row.ID)
	return nil
}

func parseDoUiTourStepsUpdateOptions(cmd *cobra.Command, args []string) (doUiTourStepsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	content, _ := cmd.Flags().GetString("content")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	sequence, _ := cmd.Flags().GetInt("sequence")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUiTourStepsUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Name:         name,
		Content:      content,
		Abbreviation: abbreviation,
		Sequence:     sequence,
	}, nil
}
