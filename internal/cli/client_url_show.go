package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

func maybeHandleClientURLShow(cmd *cobra.Command, args []string) (bool, error) {
	if !clientURLRequested(cmd) {
		return false, nil
	}
	resource, ok := resourceForSparseShow(cmd)
	if !ok {
		err := fmt.Errorf("--client-url is only supported on view <resource> show commands")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return true, err
	}
	if len(args) == 0 {
		err := fmt.Errorf("%s id is required", resource)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return true, err
	}
	id := strings.TrimSpace(args[0])
	if id == "" {
		err := fmt.Errorf("%s id is required", resource)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return true, err
	}

	handled, err := renderClientURLsFromIDIfPossible(cmd, resource, id)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return true, err
	}
	if handled {
		return true, nil
	}

	baseURL := strings.TrimSpace(getStringFlag(cmd, "base-url"))
	if baseURL == "" {
		baseURL = defaultBaseURL()
	}
	noAuth := getBoolFlag(cmd, "no-auth")
	token := strings.TrimSpace(getStringFlag(cmd, "token"))
	if noAuth {
		token = ""
	} else if token == "" {
		if resolved, _, err := auth.ResolveToken(baseURL, ""); err == nil {
			token = resolved
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return true, err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return true, err
		}
	}

	client := api.NewClient(baseURL, token)
	query := url.Values{}
	req, ok := clientURLRequirements(resource)
	fields := append([]string(nil), req.Fields...)
	include := append([]string(nil), req.Include...)
	if !ok || (len(fields) == 0 && len(include) == 0) {
		if resourceMap, err := loadResourceMap(); err == nil {
			if rels := resourceMap.Relationships[resource]; len(rels) > 0 {
				fields = nil
				include = nil
				for relName := range rels {
					fields = append(fields, relName)
					include = append(include, relName)
				}
			}
		}
	}
	fields = dedupeStrings(fields)
	include = dedupeStrings(include)
	sort.Strings(fields)
	sort.Strings(include)

	if len(fields) > 0 {
		query.Set("fields["+resource+"]", strings.Join(fields, ","))
	}
	if len(include) > 0 {
		query.Set("include", strings.Join(include, ","))
	}

	ctx := api.WithSparseFieldOverrides(cmd.Context(), api.SparseFieldOverrides{})
	body, _, err := client.Get(ctx, "/v1/"+resource+"/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return true, err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return true, err
	}
	if err := renderClientURLsForShow(cmd, resource, resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return true, err
	}
	return true, nil
}
