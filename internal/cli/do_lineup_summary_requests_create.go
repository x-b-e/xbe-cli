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

type doLineupSummaryRequestsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	LevelType      string
	LevelID        string
	StartAtMin     string
	StartAtMax     string
	EmailTo        []string
	SendIfNoShifts bool
	Note           string
}

func newDoLineupSummaryRequestsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup summary request",
		Long: `Create a lineup summary request.

Required flags:
  --level-type     Level type (Broker or Customer)
  --level-id       Level ID
  --start-at-min   Minimum shift start time (ISO 8601)
  --start-at-max   Maximum shift start time (ISO 8601)

Optional flags:
  --email-to          Email recipients (comma-separated or repeated). If omitted, the server uses the level's defaults.
  --send-if-no-shifts Send the summary even if no shifts are found
  --note              Optional note for the summary

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Request a broker lineup summary
  xbe do lineup-summary-requests create \
    --level-type Broker \
    --level-id 123 \
    --start-at-min "2026-01-23T00:00:00Z" \
    --start-at-max "2026-01-24T00:00:00Z" \
    --email-to "ops@example.com,dispatch@example.com" \
    --send-if-no-shifts \
    --note "Morning lineup summary"

  # JSON output
  xbe do lineup-summary-requests create \
    --level-type Customer \
    --level-id 456 \
    --start-at-min "2026-01-23T00:00:00Z" \
    --start-at-max "2026-01-24T00:00:00Z" \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupSummaryRequestsCreate,
	}
	initDoLineupSummaryRequestsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupSummaryRequestsCmd.AddCommand(newDoLineupSummaryRequestsCreateCmd())
}

func initDoLineupSummaryRequestsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("level-type", "", "Level type (Broker or Customer) (required)")
	cmd.Flags().String("level-id", "", "Level ID (required)")
	cmd.Flags().String("start-at-min", "", "Minimum shift start time (ISO 8601) (required)")
	cmd.Flags().String("start-at-max", "", "Maximum shift start time (ISO 8601) (required)")
	cmd.Flags().StringSlice("email-to", nil, "Email recipients (comma-separated or repeated)")
	cmd.Flags().Bool("send-if-no-shifts", false, "Send the summary even if no shifts are found")
	cmd.Flags().String("note", "", "Optional note for the summary")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupSummaryRequestsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupSummaryRequestsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.LevelType) == "" {
		err := fmt.Errorf("--level-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.LevelID) == "" {
		err := fmt.Errorf("--level-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.StartAtMin) == "" {
		err := fmt.Errorf("--start-at-min is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.StartAtMax) == "" {
		err := fmt.Errorf("--start-at-max is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	levelType, err := parseLineupSummaryRequestLevelType(opts.LevelType)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"start-at-min": opts.StartAtMin,
		"start-at-max": opts.StartAtMax,
	}

	emails := normalizeLineupSummaryRequestEmails(opts.EmailTo)
	if len(emails) > 0 {
		attributes["email-to"] = emails
	}
	if cmd.Flags().Changed("send-if-no-shifts") {
		attributes["send-if-no-shifts"] = opts.SendIfNoShifts
	}
	setStringAttrIfPresent(attributes, "note", opts.Note)

	relationships := map[string]any{
		"level": map[string]any{
			"data": map[string]any{
				"type": levelType,
				"id":   opts.LevelID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-summary-requests",
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

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-summary-requests", jsonBody)
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

	row := buildLineupSummaryRequestRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup summary request %s\n", row.ID)
	return nil
}

func parseDoLineupSummaryRequestsCreateOptions(cmd *cobra.Command) (doLineupSummaryRequestsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	levelType, _ := cmd.Flags().GetString("level-type")
	levelID, _ := cmd.Flags().GetString("level-id")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	emailTo, _ := cmd.Flags().GetStringSlice("email-to")
	sendIfNoShifts, _ := cmd.Flags().GetBool("send-if-no-shifts")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupSummaryRequestsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		LevelType:      levelType,
		LevelID:        levelID,
		StartAtMin:     startAtMin,
		StartAtMax:     startAtMax,
		EmailTo:        emailTo,
		SendIfNoShifts: sendIfNoShifts,
		Note:           note,
	}, nil
}
