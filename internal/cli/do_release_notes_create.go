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

type doReleaseNotesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Headline    string
	Description string
	ReleasedOn  string
	IsPublished bool
	Scopes      string
	IsArchived  bool
}

func newDoReleaseNotesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new release note",
		Long: `Create a new release note.

Required flags:
  --headline      Release note headline

Optional flags:
  --description   Full description
  --released-on   Release date (YYYY-MM-DD)
  --is-published  Mark as published
  --scopes        Comma-separated scopes
  --is-archived   Mark as archived`,
		Example: `  # Create a release note
  xbe do release-notes create --headline "New Feature: Dashboard"

  # Create with full details
  xbe do release-notes create --headline "Bug Fix" --description "Fixed login issue" --released-on 2025-01-20 --is-published`,
		RunE: runDoReleaseNotesCreate,
	}
	initDoReleaseNotesCreateFlags(cmd)
	return cmd
}

func init() {
	doReleaseNotesCmd.AddCommand(newDoReleaseNotesCreateCmd())
}

func initDoReleaseNotesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("headline", "", "Release note headline (required)")
	cmd.Flags().String("description", "", "Full description")
	cmd.Flags().String("released-on", "", "Release date (YYYY-MM-DD)")
	cmd.Flags().Bool("is-published", false, "Mark as published")
	cmd.Flags().String("scopes", "", "Comma-separated scopes")
	cmd.Flags().Bool("is-archived", false, "Mark as archived")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("headline")
}

func runDoReleaseNotesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoReleaseNotesCreateOptions(cmd)
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

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.ReleasedOn != "" {
		attributes["released-on"] = opts.ReleasedOn
	}
	if cmd.Flags().Changed("is-published") {
		attributes["is-published"] = opts.IsPublished
	}
	if opts.Scopes != "" {
		attributes["scopes"] = strings.Split(opts.Scopes, ",")
	}
	if cmd.Flags().Changed("is-archived") {
		attributes["is-archived"] = opts.IsArchived
	}

	data := map[string]any{
		"type":       "release-notes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/release-notes", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created release note %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "headline"))
	return nil
}

func parseDoReleaseNotesCreateOptions(cmd *cobra.Command) (doReleaseNotesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	headline, _ := cmd.Flags().GetString("headline")
	description, _ := cmd.Flags().GetString("description")
	releasedOn, _ := cmd.Flags().GetString("released-on")
	isPublished, _ := cmd.Flags().GetBool("is-published")
	scopes, _ := cmd.Flags().GetString("scopes")
	isArchived, _ := cmd.Flags().GetBool("is-archived")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doReleaseNotesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Headline:    headline,
		Description: description,
		ReleasedOn:  releasedOn,
		IsPublished: isPublished,
		Scopes:      scopes,
		IsArchived:  isArchived,
	}, nil
}
