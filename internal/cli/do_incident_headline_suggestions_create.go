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

type doIncidentHeadlineSuggestionsCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	Incident string
	Options  string
	IsAsync  bool
}

func newDoIncidentHeadlineSuggestionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an incident headline suggestion",
		Long: `Create an AI-generated headline suggestion for an incident.

Required flags:
  --incident  Incident ID (required)

Optional flags:
  --options   JSON options for suggestion generation
  --is-async  Generate suggestion asynchronously (default: true)`,
		Example: `  # Create a headline suggestion
  xbe do incident-headline-suggestions create --incident 123

  # Create with custom options
  xbe do incident-headline-suggestions create --incident 123 --options '{"temperature":0.5,"max_tokens":256}'

  # Output as JSON
  xbe do incident-headline-suggestions create --incident 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentHeadlineSuggestionsCreate,
	}
	initDoIncidentHeadlineSuggestionsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentHeadlineSuggestionsCmd.AddCommand(newDoIncidentHeadlineSuggestionsCreateCmd())
}

func initDoIncidentHeadlineSuggestionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("incident", "", "Incident ID (required)")
	cmd.Flags().String("options", "", "JSON options for suggestion generation")
	cmd.Flags().Bool("is-async", true, "Generate suggestion asynchronously")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("incident")
}

func runDoIncidentHeadlineSuggestionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentHeadlineSuggestionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Incident) == "" {
		err := fmt.Errorf("--incident is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"is-async": opts.IsAsync,
	}

	if strings.TrimSpace(opts.Options) != "" {
		var parsedOptions map[string]any
		if err := json.Unmarshal([]byte(opts.Options), &parsedOptions); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["options"] = parsedOptions
	}

	relationships := map[string]any{
		"incident": map[string]any{
			"data": map[string]any{
				"type": "incidents",
				"id":   opts.Incident,
			},
		},
	}

	data := map[string]any{
		"type":          "incident-headline-suggestions",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/incident-headline-suggestions", jsonBody)
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

	row := incidentHeadlineSuggestionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident headline suggestion %s\n", row.ID)
	return nil
}

func parseDoIncidentHeadlineSuggestionsCreateOptions(cmd *cobra.Command) (doIncidentHeadlineSuggestionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	incident, _ := cmd.Flags().GetString("incident")
	options, _ := cmd.Flags().GetString("options")
	isAsync, _ := cmd.Flags().GetBool("is-async")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentHeadlineSuggestionsCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Incident: incident,
		Options:  options,
		IsAsync:  isAsync,
	}, nil
}
