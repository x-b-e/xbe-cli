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

type doRootCausesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Title        string
	Description  string
	IsTriaged    bool
	IncidentType string
	IncidentID   string
	RootCauseID  string
}

func newDoRootCausesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new root cause",
		Long: `Create a new root cause for an incident.

Required flags:
  --incident-type    Incident type (required, e.g., production-incidents)
  --incident-id      Incident ID (required)

Optional flags:
  --title           Title
  --description     Description
  --is-triaged       Mark as triaged
  --root-cause      Parent root cause ID`,
		Example: `  # Create a root cause
  xbe do root-causes create \
    --incident-type production-incidents \
    --incident-id 123 \
    --title "Mechanical failure" \
    --description "Hydraulic leak caused downtime" \
    --is-triaged

  # Create a child root cause
  xbe do root-causes create \
    --incident-type production-incidents \
    --incident-id 123 \
    --root-cause 456 \
    --title "Seal failure"

  # JSON output
  xbe do root-causes create \
    --incident-type production-incidents \
    --incident-id 123 \
    --title "Mechanical failure" \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoRootCausesCreate,
	}
	initDoRootCausesCreateFlags(cmd)
	return cmd
}

func init() {
	doRootCausesCmd.AddCommand(newDoRootCausesCreateCmd())
}

func initDoRootCausesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Root cause title")
	cmd.Flags().String("description", "", "Root cause description")
	cmd.Flags().Bool("is-triaged", false, "Mark as triaged")
	cmd.Flags().String("incident-type", "", "Incident type (required)")
	cmd.Flags().String("incident-id", "", "Incident ID (required)")
	cmd.Flags().String("root-cause", "", "Parent root cause ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRootCausesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRootCausesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.IncidentType) == "" {
		err := fmt.Errorf("--incident-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.IncidentID) == "" {
		err := fmt.Errorf("--incident-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Title) != "" {
		attributes["title"] = opts.Title
	}
	if strings.TrimSpace(opts.Description) != "" {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("is-triaged") {
		attributes["is-triaged"] = opts.IsTriaged
	}

	relationships := map[string]any{
		"incident": map[string]any{
			"data": map[string]any{
				"type": opts.IncidentType,
				"id":   opts.IncidentID,
			},
		},
	}

	if strings.TrimSpace(opts.RootCauseID) != "" {
		relationships["root-cause"] = map[string]any{
			"data": map[string]any{
				"type": "root-causes",
				"id":   opts.RootCauseID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "root-causes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/root-causes", jsonBody)
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

	row := buildRootCauseRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Title != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created root cause %s (%s)\n", row.ID, row.Title)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created root cause %s\n", row.ID)
	return nil
}

func parseDoRootCausesCreateOptions(cmd *cobra.Command) (doRootCausesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	isTriaged, _ := cmd.Flags().GetBool("is-triaged")
	incidentType, _ := cmd.Flags().GetString("incident-type")
	incidentID, _ := cmd.Flags().GetString("incident-id")
	rootCauseID, _ := cmd.Flags().GetString("root-cause")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRootCausesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Title:        title,
		Description:  description,
		IsTriaged:    isTriaged,
		IncidentType: incidentType,
		IncidentID:   incidentID,
		RootCauseID:  rootCauseID,
	}, nil
}
