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

type doNewslettersCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Body        string
	Summary     string
	PublishedOn string
	IsPublished bool
	IsPublic    bool
	Broker      string
}

func newDoNewslettersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new newsletter",
		Long: `Create a new newsletter.

Required flags:
  --body          Newsletter body content

Optional flags:
  --summary       Newsletter summary
  --published-on  Publication date (YYYY-MM-DD)
  --is-published  Mark as published
  --is-public     Make publicly accessible
  --broker        Broker ID`,
		Example: `  # Create a newsletter
  xbe do newsletters create --body "Newsletter content here" --broker 123

  # Create a published newsletter
  xbe do newsletters create --body "Content" --is-published --published-on 2025-01-20`,
		RunE: runDoNewslettersCreate,
	}
	initDoNewslettersCreateFlags(cmd)
	return cmd
}

func init() {
	doNewslettersCmd.AddCommand(newDoNewslettersCreateCmd())
}

func initDoNewslettersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("body", "", "Newsletter body content (required)")
	cmd.Flags().String("summary", "", "Newsletter summary")
	cmd.Flags().String("published-on", "", "Publication date (YYYY-MM-DD)")
	cmd.Flags().Bool("is-published", false, "Mark as published")
	cmd.Flags().Bool("is-public", false, "Make publicly accessible")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("body")
}

func runDoNewslettersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoNewslettersCreateOptions(cmd)
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
		"body": opts.Body,
	}

	if opts.Summary != "" {
		attributes["summary"] = opts.Summary
	}
	if opts.PublishedOn != "" {
		attributes["published-on"] = opts.PublishedOn
	}
	if cmd.Flags().Changed("is-published") {
		attributes["is-published"] = opts.IsPublished
	}
	if cmd.Flags().Changed("is-public") {
		attributes["is-public"] = opts.IsPublic
	}

	relationships := map[string]any{}

	if opts.Broker != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}

	data := map[string]any{
		"type":       "newsletters",
		"attributes": attributes,
	}

	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Post(cmd.Context(), "/v1/newsletters", jsonBody)
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
			"id": resp.Data.ID,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created newsletter %s\n", resp.Data.ID)
	return nil
}

func parseDoNewslettersCreateOptions(cmd *cobra.Command) (doNewslettersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	body, _ := cmd.Flags().GetString("body")
	summary, _ := cmd.Flags().GetString("summary")
	publishedOn, _ := cmd.Flags().GetString("published-on")
	isPublished, _ := cmd.Flags().GetBool("is-published")
	isPublic, _ := cmd.Flags().GetBool("is-public")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doNewslettersCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Body:        body,
		Summary:     summary,
		PublishedOn: publishedOn,
		IsPublished: isPublished,
		IsPublic:    isPublic,
		Broker:      broker,
	}, nil
}
