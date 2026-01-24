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

type doBrokerTruckerRatingsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Rating  int
}

func newDoBrokerTruckerRatingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker trucker rating",
		Long: `Update a broker trucker rating.

Arguments:
  <id>  The broker trucker rating ID (required)

Optional flags:
  --rating  Rating (1-5)`,
		Example: `  # Update a broker trucker rating
  xbe do broker-trucker-ratings update 123 --rating 4`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerTruckerRatingsUpdate,
	}
	initDoBrokerTruckerRatingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerTruckerRatingsCmd.AddCommand(newDoBrokerTruckerRatingsUpdateCmd())
}

func initDoBrokerTruckerRatingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Int("rating", 0, "Rating (1-5)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerTruckerRatingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerTruckerRatingsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.ID) == "" {
		return fmt.Errorf("broker trucker rating id is required")
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("rating") {
		attributes["rating"] = opts.Rating
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "broker-trucker-ratings",
		"id":         opts.ID,
		"attributes": attributes,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-trucker-ratings/"+opts.ID, jsonBody)
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

	row := brokerTruckerRatingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker trucker rating %s\n", row.ID)
	return nil
}

func parseDoBrokerTruckerRatingsUpdateOptions(cmd *cobra.Command, args []string) (doBrokerTruckerRatingsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rating, _ := cmd.Flags().GetInt("rating")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerTruckerRatingsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Rating:  rating,
	}, nil
}
