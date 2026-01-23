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

type doProjectPhasesCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	Project          string
	Name             string
	Description      string
	SequencePosition int
}

func newDoProjectPhasesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase",
		Long: `Create a project phase.

Required:
  --project          Project ID
  --name             Phase name

Optional:
  --description      Phase description
  --sequence-position  Position in sequence`,
		Example: `  # Create a project phase
  xbe do project-phases create --project 123 --name "Phase 1"

  # Create with description
  xbe do project-phases create --project 123 --name "Design" --description "Design phase"

  # Create with sequence position
  xbe do project-phases create --project 123 --name "Phase 2" --sequence-position 2`,
		RunE: runDoProjectPhasesCreate,
	}
	initDoProjectPhasesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhasesCmd.AddCommand(newDoProjectPhasesCreateCmd())
}

func initDoProjectPhasesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("name", "", "Phase name")
	cmd.Flags().String("description", "", "Phase description")
	cmd.Flags().Int("sequence-position", 0, "Position in sequence")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("name")
}

func runDoProjectPhasesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhasesCreateOptions(cmd)
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
		"name": opts.Name,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("sequence-position") {
		attributes["sequence-position"] = opts.SequencePosition
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-phases",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phases", jsonBody)
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
		row := projectPhaseRow{
			ID:               resp.Data.ID,
			Name:             stringAttr(resp.Data.Attributes, "name"),
			Description:      stringAttr(resp.Data.Attributes, "description"),
			Sequence:         stringAttr(resp.Data.Attributes, "sequence"),
			SequencePosition: intAttr(resp.Data.Attributes, "sequence-position"),
		}
		if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectPhasesCreateOptions(cmd *cobra.Command) (doProjectPhasesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	sequencePosition, _ := cmd.Flags().GetInt("sequence-position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhasesCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		Project:          project,
		Name:             name,
		Description:      description,
		SequencePosition: sequencePosition,
	}, nil
}
