package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type userUiToursShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type userUiTourDetails struct {
	ID                 string `json:"id"`
	CompletedAt        string `json:"completed_at,omitempty"`
	SkippedAt          string `json:"skipped_at,omitempty"`
	UserID             string `json:"user_id,omitempty"`
	UserName           string `json:"user_name,omitempty"`
	UserEmail          string `json:"user_email,omitempty"`
	UiTourID           string `json:"ui_tour_id,omitempty"`
	UiTourName         string `json:"ui_tour_name,omitempty"`
	UiTourAbbreviation string `json:"ui_tour_abbreviation,omitempty"`
	UiTourDescription  string `json:"ui_tour_description,omitempty"`
}

func newUserUiToursShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show user UI tour details",
		Long: `Show the full details of a user UI tour.

Output Fields:
  ID                   User UI tour identifier
  Completed At         Completion timestamp
  Skipped At           Skipped timestamp
  User ID              User ID
  User Name            User name
  User Email           User email address
  UI Tour ID           UI tour ID
  UI Tour Name         UI tour name
  UI Tour Abbreviation UI tour abbreviation
  UI Tour Description  UI tour description

Arguments:
  <id>  The user UI tour ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a user UI tour
  xbe view user-ui-tours show 123

  # Output as JSON
  xbe view user-ui-tours show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUserUiToursShow,
	}
	initUserUiToursShowFlags(cmd)
	return cmd
}

func init() {
	userUiToursCmd.AddCommand(newUserUiToursShowCmd())
}

func initUserUiToursShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserUiToursShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseUserUiToursShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("user UI tour id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-ui-tours]", "completed-at,skipped-at,user,ui-tour")
	query.Set("include", "user,ui-tour")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[ui-tours]", "name,abbreviation,description")

	body, _, err := client.Get(cmd.Context(), "/v1/user-ui-tours/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildUserUiTourDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUserUiTourDetails(cmd, details)
}

func parseUserUiToursShowOptions(cmd *cobra.Command) (userUiToursShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userUiToursShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUserUiTourDetails(resp jsonAPISingleResponse) userUiTourDetails {
	attrs := resp.Data.Attributes
	details := userUiTourDetails{
		ID:          resp.Data.ID,
		CompletedAt: strings.TrimSpace(stringAttr(attrs, "completed-at")),
		SkippedAt:   strings.TrimSpace(stringAttr(attrs, "skipped-at")),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["ui-tour"]; ok && rel.Data != nil {
		details.UiTourID = rel.Data.ID
	}

	for _, inc := range resp.Included {
		switch inc.Type {
		case "users":
			if inc.ID != details.UserID {
				continue
			}
			details.UserName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			details.UserEmail = strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
		case "ui-tours":
			if inc.ID != details.UiTourID {
				continue
			}
			details.UiTourName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			details.UiTourAbbreviation = strings.TrimSpace(stringAttr(inc.Attributes, "abbreviation"))
			details.UiTourDescription = strings.TrimSpace(stringAttr(inc.Attributes, "description"))
		}
	}

	return details
}

func renderUserUiTourDetails(cmd *cobra.Command, details userUiTourDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CompletedAt != "" {
		fmt.Fprintf(out, "Completed At: %s\n", details.CompletedAt)
	}
	if details.SkippedAt != "" {
		fmt.Fprintf(out, "Skipped At: %s\n", details.SkippedAt)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.UiTourID != "" {
		fmt.Fprintf(out, "UI Tour ID: %s\n", details.UiTourID)
	}
	if details.UiTourName != "" {
		fmt.Fprintf(out, "UI Tour Name: %s\n", details.UiTourName)
	}
	if details.UiTourAbbreviation != "" {
		fmt.Fprintf(out, "UI Tour Abbrev: %s\n", details.UiTourAbbreviation)
	}
	if details.UiTourDescription != "" {
		fmt.Fprintf(out, "UI Tour Description: %s\n", details.UiTourDescription)
	}

	return nil
}
