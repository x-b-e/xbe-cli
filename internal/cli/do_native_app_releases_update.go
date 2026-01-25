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

type doNativeAppReleasesUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	ID                    string
	ReleaseChannelDetails string
	Notes                 string
}

func newDoNativeAppReleasesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a native app release",
		Long: `Update an existing native app release.

Writable attributes:
  --release-channel-details  JSON array of release channel detail objects
  --notes                    Notes about the release

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update release notes
  xbe do native-app-releases update 123 --notes "Uploaded to app stores"

  # Update release channel details
  xbe do native-app-releases update 123 \
    --release-channel-details '[{"channel":"apple-app-store","status":"released"}]'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoNativeAppReleasesUpdate,
	}
	initDoNativeAppReleasesUpdateFlags(cmd)
	return cmd
}

func init() {
	doNativeAppReleasesCmd.AddCommand(newDoNativeAppReleasesUpdateCmd())
}

func initDoNativeAppReleasesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("release-channel-details", "", "Release channel details JSON array")
	cmd.Flags().String("notes", "", "Notes about the release")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoNativeAppReleasesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoNativeAppReleasesUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("native app release id is required")
	}

	attributes := map[string]any{}
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

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "native-app-releases",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/native-app-releases/"+id, jsonBody)
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
		fmt.Fprintf(cmd.OutOrStdout(), "Updated native app release %s (%s)\n", row.ID, label)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated native app release %s\n", row.ID)
	return nil
}

func parseDoNativeAppReleasesUpdateOptions(cmd *cobra.Command, args []string) (doNativeAppReleasesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	releaseChannelDetails, _ := cmd.Flags().GetString("release-channel-details")
	notes, _ := cmd.Flags().GetString("notes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doNativeAppReleasesUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		ID:                    args[0],
		ReleaseChannelDetails: releaseChannelDetails,
		Notes:                 notes,
	}, nil
}
