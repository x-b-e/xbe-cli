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

type doShiftFeedbackReasonsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	Name             string
	Kind             string
	Slug             string
	DefaultRating    string
	CorrectiveAction string
}

func newDoShiftFeedbackReasonsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new shift feedback reason",
		Long: `Create a new shift feedback reason.

Required flags:
  --name  The reason name (required)
  --kind  The feedback kind (required, e.g., positive, negative)
  --slug  URL-friendly identifier (required)

Optional flags:
  --default-rating     Default rating value
  --corrective-action  Corrective action text`,
		Example: `  # Create a positive feedback reason
  xbe do shift-feedback-reasons create --name "Great Work" --kind positive --slug "great-work"

  # Create a negative feedback reason with corrective action
  xbe do shift-feedback-reasons create --name "Late Arrival" --kind negative --slug "late-arrival" --corrective-action "Please arrive on time"

  # Get JSON output
  xbe do shift-feedback-reasons create --name "Good Attitude" --kind positive --slug "good-attitude" --json`,
		Args: cobra.NoArgs,
		RunE: runDoShiftFeedbackReasonsCreate,
	}
	initDoShiftFeedbackReasonsCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftFeedbackReasonsCmd.AddCommand(newDoShiftFeedbackReasonsCreateCmd())
}

func initDoShiftFeedbackReasonsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Reason name (required)")
	cmd.Flags().String("kind", "", "Feedback kind (required, e.g., positive, negative)")
	cmd.Flags().String("slug", "", "URL-friendly identifier (required)")
	cmd.Flags().String("default-rating", "", "Default rating value")
	cmd.Flags().String("corrective-action", "", "Corrective action text")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftFeedbackReasonsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftFeedbackReasonsCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Kind == "" {
		err := fmt.Errorf("--kind is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Slug == "" {
		err := fmt.Errorf("--slug is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
		"kind": opts.Kind,
		"slug": opts.Slug,
	}

	if opts.DefaultRating != "" {
		attributes["default-rating"] = opts.DefaultRating
	}
	if opts.CorrectiveAction != "" {
		attributes["corrective-action"] = opts.CorrectiveAction
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "shift-feedback-reasons",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/shift-feedback-reasons", jsonBody)
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

	row := buildShiftFeedbackReasonRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created shift feedback reason %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoShiftFeedbackReasonsCreateOptions(cmd *cobra.Command) (doShiftFeedbackReasonsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	kind, _ := cmd.Flags().GetString("kind")
	slug, _ := cmd.Flags().GetString("slug")
	defaultRating, _ := cmd.Flags().GetString("default-rating")
	correctiveAction, _ := cmd.Flags().GetString("corrective-action")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftFeedbackReasonsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		Name:             name,
		Kind:             kind,
		Slug:             slug,
		DefaultRating:    defaultRating,
		CorrectiveAction: correctiveAction,
	}, nil
}

func buildShiftFeedbackReasonRowFromSingle(resp jsonAPISingleResponse) shiftFeedbackReasonRow {
	attrs := resp.Data.Attributes

	return shiftFeedbackReasonRow{
		ID:               resp.Data.ID,
		Name:             stringAttr(attrs, "name"),
		Kind:             stringAttr(attrs, "kind"),
		DefaultRating:    attrs["default-rating"],
		Slug:             stringAttr(attrs, "slug"),
		CorrectiveAction: stringAttr(attrs, "corrective-action"),
	}
}
