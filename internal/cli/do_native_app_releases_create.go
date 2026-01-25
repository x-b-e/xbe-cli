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

type doNativeAppReleasesCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	GitTag                string
	GitSHA                string
	BuildNumber           string
	ReleaseChannelDetails string
	Notes                 string
}

func newDoNativeAppReleasesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a native app release",
		Long: `Create a native app release.

Required flags:
  --git-sha          Git commit SHA
  --build-number     Build number

Optional flags:
  --git-tag                 Git tag
  --release-channel-details JSON array of release channel detail objects
  --notes                   Notes about the release

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a native app release
  xbe do native-app-releases create --git-sha abc123 --build-number 101

  # Create with release channel details
  xbe do native-app-releases create \
    --git-sha abc123 \
    --build-number 101 \
    --release-channel-details '[{"channel":"apple-app-store","status":"uploaded"}]'`,
		RunE: runDoNativeAppReleasesCreate,
	}
	initDoNativeAppReleasesCreateFlags(cmd)
	return cmd
}

func init() {
	doNativeAppReleasesCmd.AddCommand(newDoNativeAppReleasesCreateCmd())
}

func initDoNativeAppReleasesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("git-tag", "", "Git tag")
	cmd.Flags().String("git-sha", "", "Git commit SHA (required)")
	cmd.Flags().String("build-number", "", "Build number (required)")
	cmd.Flags().String("release-channel-details", "", "Release channel details JSON array")
	cmd.Flags().String("notes", "", "Notes about the release")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("git-sha")
	cmd.MarkFlagRequired("build-number")
}

func runDoNativeAppReleasesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoNativeAppReleasesCreateOptions(cmd)
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

	gitSHA := strings.TrimSpace(opts.GitSHA)
	if gitSHA == "" {
		err := fmt.Errorf("git-sha is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	buildNumber := strings.TrimSpace(opts.BuildNumber)
	if buildNumber == "" {
		err := fmt.Errorf("build-number is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"git-sha":      gitSHA,
		"build-number": buildNumber,
	}

	if strings.TrimSpace(opts.GitTag) != "" {
		attributes["git-tag"] = strings.TrimSpace(opts.GitTag)
	}
	if cmd.Flags().Changed("release-channel-details") {
		details, err := parseReleaseChannelDetails(opts.ReleaseChannelDetails)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["release-channel-details"] = details
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "native-app-releases",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/native-app-releases", jsonBody)
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

	row := nativeAppReleaseRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	label := row.BuildNumber
	if label == "" {
		label = row.GitSHA
	}
	if label != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created native app release %s (%s)\n", row.ID, label)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created native app release %s\n", row.ID)
	return nil
}

func parseDoNativeAppReleasesCreateOptions(cmd *cobra.Command) (doNativeAppReleasesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	gitTag, _ := cmd.Flags().GetString("git-tag")
	gitSHA, _ := cmd.Flags().GetString("git-sha")
	buildNumber, _ := cmd.Flags().GetString("build-number")
	releaseChannelDetails, _ := cmd.Flags().GetString("release-channel-details")
	notes, _ := cmd.Flags().GetString("notes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doNativeAppReleasesCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		GitTag:                gitTag,
		GitSHA:                gitSHA,
		BuildNumber:           buildNumber,
		ReleaseChannelDetails: releaseChannelDetails,
		Notes:                 notes,
	}, nil
}

func parseReleaseChannelDetails(raw string) ([]map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("--release-channel-details cannot be empty")
	}

	var details []map[string]any
	if err := json.Unmarshal([]byte(raw), &details); err != nil {
		return nil, fmt.Errorf("invalid release-channel-details JSON: %w", err)
	}

	return details, nil
}
