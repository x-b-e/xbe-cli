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

type doPressReleasesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Slug         string
	Headline     string
	Subheadline  string
	Body         string
	ReleasedAt   string
	LocationName string
	Published    bool
}

func newDoPressReleasesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new press release",
		Long: `Create a new press release.

Required flags:
  --headline      Press release headline

Optional flags:
  --slug          URL-friendly slug
  --subheadline   Subheadline
  --body          Press release body
  --released-at   Release datetime (ISO 8601)
  --location-name Location name
  --published     Mark as published`,
		Example: `  # Create a press release
  xbe do press-releases create --headline "Company Announces New Product"

  # Create with full details
  xbe do press-releases create --headline "Big News" --subheadline "Details here" --body "Full story..." --published`,
		RunE: runDoPressReleasesCreate,
	}
	initDoPressReleasesCreateFlags(cmd)
	return cmd
}

func init() {
	doPressReleasesCmd.AddCommand(newDoPressReleasesCreateCmd())
}

func initDoPressReleasesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("slug", "", "URL-friendly slug")
	cmd.Flags().String("headline", "", "Press release headline (required)")
	cmd.Flags().String("subheadline", "", "Subheadline")
	cmd.Flags().String("body", "", "Press release body")
	cmd.Flags().String("released-at", "", "Release datetime (ISO 8601)")
	cmd.Flags().String("location-name", "", "Location name")
	cmd.Flags().Bool("published", false, "Mark as published")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("headline")
}

func runDoPressReleasesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPressReleasesCreateOptions(cmd)
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
		"headline": opts.Headline,
	}

	if opts.Slug != "" {
		attributes["slug"] = opts.Slug
	}
	if opts.Subheadline != "" {
		attributes["subheadline"] = opts.Subheadline
	}
	if opts.Body != "" {
		attributes["body"] = opts.Body
	}
	if opts.ReleasedAt != "" {
		attributes["released-at"] = opts.ReleasedAt
	}
	if opts.LocationName != "" {
		attributes["location-name"] = opts.LocationName
	}
	if cmd.Flags().Changed("published") {
		attributes["published"] = opts.Published
	}

	data := map[string]any{
		"type":       "press-releases",
		"attributes": attributes,
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

	body, _, err := client.Post(cmd.Context(), "/v1/press-releases", jsonBody)
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
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id":       resp.Data.ID,
			"headline": stringAttr(resp.Data.Attributes, "headline"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created press release %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "headline"))
	return nil
}

func parseDoPressReleasesCreateOptions(cmd *cobra.Command) (doPressReleasesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	slug, _ := cmd.Flags().GetString("slug")
	headline, _ := cmd.Flags().GetString("headline")
	subheadline, _ := cmd.Flags().GetString("subheadline")
	body, _ := cmd.Flags().GetString("body")
	releasedAt, _ := cmd.Flags().GetString("released-at")
	locationName, _ := cmd.Flags().GetString("location-name")
	published, _ := cmd.Flags().GetBool("published")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPressReleasesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Slug:         slug,
		Headline:     headline,
		Subheadline:  subheadline,
		Body:         body,
		ReleasedAt:   releasedAt,
		LocationName: locationName,
		Published:    published,
	}, nil
}
