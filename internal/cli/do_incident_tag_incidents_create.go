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

type doIncidentTagIncidentsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Incident    string
	IncidentTag string
}

func newDoIncidentTagIncidentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an incident tag incident link",
		Long: `Create an incident tag incident link.

Required flags:
  --incident      Incident ID (required)
  --incident-tag  Incident tag ID (required)

Incident tags must be valid for the incident kind; the server will reject
incompatible tag assignments.`,
		Example: `  # Tag an incident
  xbe do incident-tag-incidents create --incident 123 --incident-tag 456`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentTagIncidentsCreate,
	}
	initDoIncidentTagIncidentsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentTagIncidentsCmd.AddCommand(newDoIncidentTagIncidentsCreateCmd())
}

func initDoIncidentTagIncidentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("incident", "", "Incident ID (required)")
	cmd.Flags().String("incident-tag", "", "Incident tag ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentTagIncidentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentTagIncidentsCreateOptions(cmd)
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

	if opts.Incident == "" {
		err := fmt.Errorf("--incident is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.IncidentTag == "" {
		err := fmt.Errorf("--incident-tag is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"incident": map[string]any{
			"data": map[string]any{
				"type": "incidents",
				"id":   opts.Incident,
			},
		},
		"incident-tag": map[string]any{
			"data": map[string]any{
				"type": "incident-tags",
				"id":   opts.IncidentTag,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "incident-tag-incidents",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/incident-tag-incidents", jsonBody)
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

	row := incidentTagIncidentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident tag incident %s\n", row.ID)
	return nil
}

func parseDoIncidentTagIncidentsCreateOptions(cmd *cobra.Command) (doIncidentTagIncidentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	incident, _ := cmd.Flags().GetString("incident")
	incidentTag, _ := cmd.Flags().GetString("incident-tag")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentTagIncidentsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Incident:    incident,
		IncidentTag: incidentTag,
	}, nil
}
