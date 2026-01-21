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

type doStakeholderClassificationsCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	Title                       string
	LeverageFactor              int
	ObjectivesNarrativeExplicit string
}

func newDoStakeholderClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new stakeholder classification",
		Long: `Create a new stakeholder classification.

Required flags:
  --title  The classification title (required)

Optional flags:
  --leverage-factor              Influence level (integer)
  --objectives-narrative-explicit  Explicit objectives narrative

Note: Only admin users can create stakeholder classifications.`,
		Example: `  # Create a basic classification
  xbe do stakeholder-classifications create --title "Project Owner"

  # Create with leverage factor
  xbe do stakeholder-classifications create --title "Key Stakeholder" --leverage-factor 5

  # Create with narrative
  xbe do stakeholder-classifications create --title "Sponsor" --objectives-narrative-explicit "Provides funding and strategic direction"

  # Get JSON output
  xbe do stakeholder-classifications create --title "Test" --json`,
		Args: cobra.NoArgs,
		RunE: runDoStakeholderClassificationsCreate,
	}
	initDoStakeholderClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doStakeholderClassificationsCmd.AddCommand(newDoStakeholderClassificationsCreateCmd())
}

func initDoStakeholderClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Classification title (required)")
	cmd.Flags().Int("leverage-factor", 0, "Leverage factor (influence level)")
	cmd.Flags().String("objectives-narrative-explicit", "", "Explicit objectives narrative")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoStakeholderClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoStakeholderClassificationsCreateOptions(cmd)
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

	if opts.Title == "" {
		err := fmt.Errorf("--title is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"title": opts.Title,
	}
	if cmd.Flags().Changed("leverage-factor") {
		attributes["leverage-factor"] = opts.LeverageFactor
	}
	if opts.ObjectivesNarrativeExplicit != "" {
		attributes["objectives-narrative-explicit"] = opts.ObjectivesNarrativeExplicit
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "stakeholder-classifications",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/stakeholder-classifications", jsonBody)
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

	row := buildStakeholderClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created stakeholder classification %s (%s)\n", row.ID, row.Title)
	return nil
}

func parseDoStakeholderClassificationsCreateOptions(cmd *cobra.Command) (doStakeholderClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	leverageFactor, _ := cmd.Flags().GetInt("leverage-factor")
	objectivesNarrativeExplicit, _ := cmd.Flags().GetString("objectives-narrative-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doStakeholderClassificationsCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		Title:                       title,
		LeverageFactor:              leverageFactor,
		ObjectivesNarrativeExplicit: objectivesNarrativeExplicit,
	}, nil
}

func buildStakeholderClassificationRowFromSingle(resp jsonAPISingleResponse) stakeholderClassificationRow {
	attrs := resp.Data.Attributes

	row := stakeholderClassificationRow{
		ID:                          resp.Data.ID,
		Title:                       stringAttr(attrs, "title"),
		Slug:                        stringAttr(attrs, "slug"),
		ObjectivesNarrativeExplicit: stringAttr(attrs, "objectives-narrative-explicit"),
		ObjectivesNarrative:         stringAttr(attrs, "objectives-narrative"),
	}

	if lf, ok := attrs["leverage-factor"]; ok && lf != nil {
		if lfFloat, ok := lf.(float64); ok {
			lfInt := int(lfFloat)
			row.LeverageFactor = &lfInt
		}
	}

	return row
}
