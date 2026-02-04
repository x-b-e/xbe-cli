package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

const clientURLMaxRelationshipDepth = 6

var clientURLAliases = map[string][]string{
	"branches":                              {"brokers"},
	"branch":                                {"brokers"},
	"shifts":                                {"job-schedule-shifts", "tender-job-schedule-shifts"},
	"shift":                                 {"job-schedule-shifts", "tender-job-schedule-shifts"},
	"revenue-items":                         {"project-revenue-items"},
	"revenue-item":                          {"project-revenue-items"},
	"cost-items":                            {"project-phase-cost-items"},
	"cost-item":                             {"project-phase-cost-items"},
	"referrals":                             {"trucker-referrals"},
	"referral":                              {"trucker-referrals"},
	"feedback":                              {"shift-feedbacks"},
	"feedbacks":                             {"shift-feedbacks"},
	"reference-types":                       {"developer-reference-types"},
	"reference-type":                        {"developer-reference-types"},
	"driver-day":                            {"driver-days"},
	"driver-days":                           {"driver-days"},
	"phases":                                {"project-phases"},
	"equipments":                            {"equipment"},
	"actual":                                {"project-phase-cost-item-actuals", "project-phase-revenue-item-actuals"},
	"driver-day-time-card-constraints":      {"shift-set-time-card-constraints"},
	"image":                                 {"file-attachments"},
	"transport-plan-projects":               {"projects"},
	"keep-truckin":                          {"go-motive-integrations"},
	"generation":                            {"invoice-generations"},
	"game-plan":                             {"objectives"},
	"eticketing":                            {"job-production-plans"},
	"safety-meeting":                        {"meetings"},
	"notification-feeds":                    {"notifications"},
	"laborer-crafts":                        {"craft-classes"},
	"dvir":                                  {"equipment"},
	"trucker-certification-classifications": {"developer-trucker-certification-classifications"},
	"proof-of-delivery":                     {"material-transactions"},
	"samsara":                               {"samsara-integrations"},
}

type clientRouteBinding struct {
	Route               clientRoute
	ParamBindings       []routeParamBinding
	ReferencedResources []string
}

type appURLRequirement struct {
	Fields  []string
	Include []string
}

type routeParamBinding struct {
	Name               string
	Segment            string
	ResourceCandidates []string
}

type clientURLResolver struct {
	cmd               *cobra.Command
	baseURL           string
	apiBaseURL        string
	token             string
	client            *api.Client
	resourceMap       resourceMap
	resourceSet       map[string]struct{}
	apiTypeToResource map[string]string
	routeBindings     map[string][]clientRouteBinding
	included          map[string]jsonAPIResource
	fetched           map[string]jsonAPIResource
	resolveCache      map[string]map[string][]string
}

type relatedRef struct {
	ID   string
	Type string
}

type resourceNode struct {
	Resource jsonAPIResource
	Type     string
	Depth    int
}

var clientURLRequirementCache = struct {
	sync.Mutex
	data      map[string]appURLRequirement
	supported map[string]bool
}{}

func clientURLRequested(cmd *cobra.Command) bool {
	return getBoolFlag(cmd, "client-url")
}

func clientURLRequirements(resource string) (appURLRequirement, bool) {
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return appURLRequirement{}, false
	}

	clientURLRequirementCache.Lock()
	if clientURLRequirementCache.data == nil {
		clientURLRequirementCache.data = map[string]appURLRequirement{}
		clientURLRequirementCache.supported = map[string]bool{}
	}
	if supported, ok := clientURLRequirementCache.supported[resource]; ok {
		req := clientURLRequirementCache.data[resource]
		clientURLRequirementCache.Unlock()
		return req, supported
	}
	clientURLRequirementCache.Unlock()

	req, supported, err := computeClientURLRequirements(resource)
	if err != nil {
		return appURLRequirement{}, false
	}

	clientURLRequirementCache.Lock()
	clientURLRequirementCache.data[resource] = req
	clientURLRequirementCache.supported[resource] = supported
	clientURLRequirementCache.Unlock()
	return req, supported
}

func applyClientURLOverrides(cmd *cobra.Command, resource string, overrides api.SparseFieldOverrides) api.SparseFieldOverrides {
	if !clientURLRequested(cmd) {
		return overrides
	}
	req, ok := clientURLRequirements(resource)
	if !ok {
		return overrides
	}
	if len(req.Fields) > 0 {
		overrides.FieldsSet = true
		overrides.Primary = mergeStringSlices(overrides.Primary, req.Fields)
	}
	if len(req.Include) > 0 {
		overrides.IncludeSet = true
		overrides.Include = mergeStringSlices(overrides.Include, req.Include)
	}
	return overrides
}

func renderClientURLsFromIDIfPossible(cmd *cobra.Command, resource, id string) bool {
	canResolve, bindings, err := clientURLCanResolveFromID(resource)
	if err != nil {
		logClientURLError(cmd, err)
		_ = writeClientURLOutput(cmd, nil)
		return true
	}
	if !canResolve {
		return false
	}
	urls, err := clientURLsFromID(cmd, resource, id, bindings)
	if err != nil {
		logClientURLError(cmd, err)
		_ = writeClientURLOutput(cmd, nil)
		return true
	}
	_ = writeClientURLOutput(cmd, urls)
	return true
}

func renderClientURLsForList(cmd *cobra.Command, resource string, resp jsonAPIResponse) error {
	urls, err := clientURLsForResources(cmd, resource, resp.Data, resp.Included)
	if err != nil {
		logClientURLError(cmd, err)
	}
	return writeClientURLOutput(cmd, urls)
}

func renderClientURLsForShow(cmd *cobra.Command, resource string, resp jsonAPISingleResponse) error {
	urls, err := clientURLsForResources(cmd, resource, []jsonAPIResource{resp.Data}, resp.Included)
	if err != nil {
		logClientURLError(cmd, err)
	}
	return writeClientURLOutput(cmd, urls)
}

func clientURLsForResources(cmd *cobra.Command, resource string, resources []jsonAPIResource, included []jsonAPIResource) ([]string, error) {
	resolver, err := newClientURLResolver(cmd, included)
	if err != nil {
		return nil, err
	}
	routes, err := resolver.routesForResource(resource)
	if err != nil {
		return nil, err
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("no client routes available for %q", resource)
	}

	urls := []string{}
	seen := map[string]struct{}{}
	var firstErr error
	for _, res := range resources {
		resURLs, err := resolver.urlsForResource(resource, res, routes)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s %s: %w", resource, res.ID, err)
			}
			continue
		}
		for _, url := range resURLs {
			url = strings.TrimSpace(url)
			if url == "" {
				continue
			}
			if _, exists := seen[url]; exists {
				continue
			}
			seen[url] = struct{}{}
			urls = append(urls, url)
		}
	}
	if len(urls) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, fmt.Errorf("no client URLs available for %q", resource)
	}
	return urls, nil
}

func newClientURLResolver(cmd *cobra.Command, included []jsonAPIResource) (*clientURLResolver, error) {
	resourceMap, err := loadResourceMap()
	if err != nil {
		return nil, err
	}
	resourceSet := make(map[string]struct{}, len(resourceMap.Resources))
	apiTypeToResource := map[string]string{}
	for name, spec := range resourceMap.Resources {
		resourceSet[name] = struct{}{}
		for _, serverType := range spec.ServerTypes {
			if serverType == "" {
				continue
			}
			if _, exists := apiTypeToResource[serverType]; !exists {
				apiTypeToResource[serverType] = name
			}
		}
	}

	baseURL := resolveClientBaseURL(cmd)
	apiBaseURL := strings.TrimSpace(getStringFlag(cmd, "base-url"))
	if apiBaseURL == "" {
		apiBaseURL = defaultBaseURL()
	}

	noAuth := getBoolFlag(cmd, "no-auth")
	token := strings.TrimSpace(getStringFlag(cmd, "token"))
	if noAuth {
		token = ""
	} else if token == "" {
		if resolved, _, err := auth.ResolveToken(apiBaseURL, ""); err == nil {
			token = resolved
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return nil, err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return nil, err
		}
	}

	var client *api.Client
	if token != "" {
		client = api.NewClient(apiBaseURL, token)
	}

	resolver := &clientURLResolver{
		cmd:               cmd,
		baseURL:           baseURL,
		apiBaseURL:        apiBaseURL,
		token:             token,
		client:            client,
		resourceMap:       resourceMap,
		resourceSet:       resourceSet,
		apiTypeToResource: apiTypeToResource,
		routeBindings:     map[string][]clientRouteBinding{},
		included:          map[string]jsonAPIResource{},
		fetched:           map[string]jsonAPIResource{},
		resolveCache:      map[string]map[string][]string{},
	}
	resolver.cacheIncluded(included)
	return resolver, nil
}

func (r *clientURLResolver) routesForResource(resource string) ([]clientRouteBinding, error) {
	if cached, ok := r.routeBindings[resource]; ok {
		return cached, nil
	}
	catalog, err := loadClientRoutes()
	if err != nil {
		return nil, err
	}
	bindings := []clientRouteBinding{}
	for _, route := range catalog.Routes {
		if route.Action {
			continue
		}
		binding := buildClientRouteBinding(route, r.resourceSet)
		if !binding.referencesResource(resource) {
			continue
		}
		bindings = append(bindings, binding)
	}
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Route.Path < bindings[j].Route.Path
	})
	r.routeBindings[resource] = bindings
	return bindings, nil
}

func buildClientRouteBinding(route clientRoute, resourceSet map[string]struct{}) clientRouteBinding {
	parts := splitRoutePath(route.Path)
	paramBindings := []routeParamBinding{}
	referenced := map[string]struct{}{}

	for idx, part := range parts {
		if isRouteParam(part) {
			paramName := strings.TrimLeft(part, ":*")
			segment := ""
			if idx > 0 {
				prev := parts[idx-1]
				if prev != "" && !isRouteParam(prev) {
					segment = prev
				}
			}
			candidates := resourceCandidatesForSegment(segment, resourceSet)
			if len(candidates) == 0 {
				candidates = resourceCandidatesForParam(paramName, resourceSet)
			}
			for _, candidate := range candidates {
				referenced[candidate] = struct{}{}
			}
			paramBindings = append(paramBindings, routeParamBinding{
				Name:               paramName,
				Segment:            segment,
				ResourceCandidates: candidates,
			})
			continue
		}
		for _, candidate := range resourceCandidatesForSegment(part, resourceSet) {
			referenced[candidate] = struct{}{}
		}
	}

	referencedList := make([]string, 0, len(referenced))
	for candidate := range referenced {
		referencedList = append(referencedList, candidate)
	}
	sort.Strings(referencedList)

	return clientRouteBinding{
		Route:               route,
		ParamBindings:       paramBindings,
		ReferencedResources: referencedList,
	}
}

func (binding clientRouteBinding) referencesResource(resource string) bool {
	for _, candidate := range binding.ReferencedResources {
		if candidate == resource {
			return true
		}
	}
	return false
}

func (r *clientURLResolver) urlsForResource(resource string, res jsonAPIResource, routes []clientRouteBinding) ([]string, error) {
	root := res
	if len(root.Relationships) == 0 {
		if fetched, err := r.fetchResource(res.Type, res.ID); err == nil {
			root = fetched
		}
	}

	urls := []string{}
	var firstErr error
	for _, binding := range routes {
		paramsList, err := r.resolveRouteParams(root, binding)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", binding.Route.Path, err)
			}
			continue
		}
		for _, params := range paramsList {
			resolved, err := applyRouteParams(binding.Route.Path, params)
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("%s: %w", binding.Route.Path, err)
				}
				continue
			}
			urls = append(urls, clientURL(r.baseURL, resolved))
		}
	}
	if len(urls) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, fmt.Errorf("no client URLs available for %q", resource)
	}
	return urls, nil
}

func (r *clientURLResolver) resolveRouteParams(resource jsonAPIResource, binding clientRouteBinding) ([]map[string]string, error) {
	if len(binding.ParamBindings) == 0 {
		return []map[string]string{{}}, nil
	}
	paramsList := []map[string]string{{}}
	for _, param := range binding.ParamBindings {
		values, err := r.resolveParamValues(resource, param)
		if err != nil {
			return nil, err
		}
		if len(values) == 0 {
			return nil, fmt.Errorf("missing value for %s", param.Name)
		}
		paramsList = combineParamValues(paramsList, param.Name, values)
		if len(paramsList) == 0 {
			return nil, fmt.Errorf("missing value for %s", param.Name)
		}
	}
	return paramsList, nil
}

func (r *clientURLResolver) resolveParamValues(resource jsonAPIResource, param routeParamBinding) ([]string, error) {
	values := []string{}
	if len(param.ResourceCandidates) > 0 {
		for _, candidate := range param.ResourceCandidates {
			ids, err := r.resolveResourceIDs(resource, candidate)
			if err != nil {
				return nil, err
			}
			values = append(values, ids...)
		}
	}
	values = append(values, attributeValuesForParam(resource, param.Name)...)
	values = dedupeStrings(values)
	return values, nil
}

func (r *clientURLResolver) resolveResourceIDs(resource jsonAPIResource, target string) ([]string, error) {
	cacheKey := resourceKey(resource.Type, resource.ID)
	if cached, ok := r.resolveCache[cacheKey]; ok {
		if values, ok := cached[target]; ok {
			return values, nil
		}
	}

	root := resource
	if len(root.Relationships) == 0 {
		if fetched, err := r.fetchResource(resource.Type, resource.ID); err == nil {
			root = fetched
		}
	}

	results, err := r.findReachableResources(root, target)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		results = append(results, attributeValuesForTarget(root, target)...)
	}
	results = dedupeStrings(results)

	if r.resolveCache[cacheKey] == nil {
		r.resolveCache[cacheKey] = map[string][]string{}
	}
	r.resolveCache[cacheKey][target] = results
	return results, nil
}

func (r *clientURLResolver) findReachableResources(root jsonAPIResource, target string) ([]string, error) {
	target = r.normalizeResourceType(target)
	queue := []resourceNode{{Resource: root, Type: r.normalizeResourceType(root.Type), Depth: 0}}
	visited := map[string]struct{}{resourceKey(root.Type, root.ID): {}}
	results := map[string]struct{}{}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		node.Resource = r.ensureRelationshipData(node.Resource, node.Type)

		if node.Type == target {
			results[node.Resource.ID] = struct{}{}
			continue
		}
		if node.Depth >= clientURLMaxRelationshipDepth {
			continue
		}

		rels := r.resourceMap.Relationships[node.Type]
		for relName, relSpec := range rels {
			refs := relationshipValues(node.Resource, relName, relSpec)
			for _, ref := range refs {
				if ref.ID == "" {
					continue
				}
				apiType := ref.Type
				if apiType == "" {
					apiType = relSpec.ResourceType()
				}
				if apiType == "" {
					continue
				}
				key := resourceKey(apiType, ref.ID)
				if _, ok := visited[key]; ok {
					continue
				}
				visited[key] = struct{}{}
				resource, err := r.getResource(apiType, ref.ID)
				if err != nil {
					return nil, err
				}
				nodeType := r.normalizeResourceType(resource.Type)
				queue = append(queue, resourceNode{Resource: resource, Type: nodeType, Depth: node.Depth + 1})
			}
		}
	}

	out := make([]string, 0, len(results))
	for id := range results {
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

func (r *clientURLResolver) ensureRelationshipData(resource jsonAPIResource, resourceType string) jsonAPIResource {
	if r.client == nil || resource.ID == "" || resourceType == "" {
		return resource
	}
	rels := r.resourceMap.Relationships[resourceType]
	if len(rels) == 0 {
		return resource
	}
	for relName, relSpec := range rels {
		if len(relationshipValues(resource, relName, relSpec)) > 0 {
			return resource
		}
	}

	fields := make([]string, 0, len(rels))
	for relName := range rels {
		fields = append(fields, relName)
	}
	sort.Strings(fields)
	fetched, err := r.fetchResourceWithFields(resource.Type, resource.ID, fields)
	if err != nil {
		return resource
	}
	return fetched
}

func relationshipValues(resource jsonAPIResource, relName string, relSpec relationshipSpec) []relatedRef {
	if rel, ok := resource.Relationships[relName]; ok && rel.Data != nil {
		return []relatedRef{{ID: rel.Data.ID, Type: rel.Data.Type}}
	}

	attrKey := relName + "-id"
	if id := stringAttr(resource.Attributes, attrKey); id != "" {
		if len(relSpec.Resources) == 1 {
			return []relatedRef{{ID: id, Type: relSpec.Resources[0]}}
		}
		if len(relSpec.Resources) > 1 {
			refs := make([]relatedRef, 0, len(relSpec.Resources))
			for _, resourceType := range relSpec.Resources {
				refs = append(refs, relatedRef{ID: id, Type: resourceType})
			}
			return refs
		}
		return []relatedRef{{ID: id}}
	}

	attrKey = relName + "-ids"
	ids := stringSliceAttr(resource.Attributes, attrKey)
	if len(ids) == 0 {
		return nil
	}

	refs := []relatedRef{}
	if len(relSpec.Resources) == 0 {
		for _, id := range ids {
			refs = append(refs, relatedRef{ID: id})
		}
		return refs
	}
	for _, resourceType := range relSpec.Resources {
		for _, id := range ids {
			refs = append(refs, relatedRef{ID: id, Type: resourceType})
		}
	}
	return refs
}

func (spec relationshipSpec) ResourceType() string {
	if len(spec.Resources) == 1 {
		return spec.Resources[0]
	}
	return ""
}

func (r *clientURLResolver) normalizeResourceType(apiType string) string {
	if apiType == "" {
		return ""
	}
	if _, ok := r.resourceSet[apiType]; ok {
		return apiType
	}
	if mapped, ok := r.apiTypeToResource[apiType]; ok {
		return mapped
	}
	return apiType
}

func (r *clientURLResolver) cacheIncluded(included []jsonAPIResource) {
	for _, res := range included {
		r.cacheResource(res)
	}
}

func (r *clientURLResolver) cacheResource(res jsonAPIResource) {
	if res.ID == "" {
		return
	}
	key := resourceKey(res.Type, res.ID)
	r.included[key] = res

	normalized := r.normalizeResourceType(res.Type)
	if normalized != res.Type {
		alt := resourceKey(normalized, res.ID)
		r.included[alt] = res
	}
}

func (r *clientURLResolver) fetchResource(typ, id string) (jsonAPIResource, error) {
	return r.fetchResourceWithFields(typ, id, nil)
}

func (r *clientURLResolver) fetchResourceWithFields(typ, id string, fields []string) (jsonAPIResource, error) {
	if typ == "" || id == "" {
		return jsonAPIResource{}, fmt.Errorf("missing resource type or id")
	}
	if resource, ok := r.lookupResource(typ, id); ok {
		return resource, nil
	}
	if r.client == nil {
		return jsonAPIResource{}, fmt.Errorf("cannot resolve related %s %s without auth", typ, id)
	}

	query := url.Values{}
	pathType := r.normalizeResourceType(typ)
	if pathType == "" {
		pathType = typ
	}
	if len(fields) > 0 {
		query.Set("fields["+pathType+"]", strings.Join(fields, ","))
	}
	ctx := api.WithSparseFieldOverrides(r.cmd.Context(), api.SparseFieldOverrides{})
	body, _, err := r.client.Get(ctx, "/v1/"+pathType+"/"+id, query)
	if err != nil {
		return jsonAPIResource{}, err
	}
	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return jsonAPIResource{}, err
	}
	r.cacheFetched(resp.Data)
	for _, inc := range resp.Included {
		r.cacheFetched(inc)
	}
	return resp.Data, nil
}

func (r *clientURLResolver) getResource(typ, id string) (jsonAPIResource, error) {
	if resource, ok := r.lookupResource(typ, id); ok {
		return resource, nil
	}
	return r.fetchResource(typ, id)
}

func (r *clientURLResolver) lookupResource(typ, id string) (jsonAPIResource, bool) {
	if typ == "" || id == "" {
		return jsonAPIResource{}, false
	}
	key := resourceKey(typ, id)
	if resource, ok := r.fetched[key]; ok {
		return resource, true
	}
	if resource, ok := r.included[key]; ok {
		return resource, true
	}

	normalized := r.normalizeResourceType(typ)
	if normalized != typ {
		key = resourceKey(normalized, id)
		if resource, ok := r.fetched[key]; ok {
			return resource, true
		}
		if resource, ok := r.included[key]; ok {
			return resource, true
		}
	}
	return jsonAPIResource{}, false
}

func (r *clientURLResolver) cacheFetched(res jsonAPIResource) {
	if res.ID == "" {
		return
	}
	key := resourceKey(res.Type, res.ID)
	r.fetched[key] = res

	normalized := r.normalizeResourceType(res.Type)
	if normalized != res.Type {
		alt := resourceKey(normalized, res.ID)
		r.fetched[alt] = res
	}
}

func attributeValuesForParam(resource jsonAPIResource, param string) []string {
	if param == "" {
		return nil
	}
	key := strings.ReplaceAll(param, "_", "-")
	values := attributeValues(resource, key)
	if len(values) > 0 {
		return values
	}
	return nil
}

func attributeValuesForTarget(resource jsonAPIResource, target string) []string {
	if target == "" {
		return nil
	}
	keys := []string{}
	singular := singularize(strings.ReplaceAll(target, "_", "-"))
	if singular != "" {
		keys = append(keys, singular+"-id", singular+"-ids")
	}
	keys = append(keys, target+"-id", target+"-ids")
	seen := map[string]struct{}{}
	for _, key := range keys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		values := attributeValues(resource, key)
		if len(values) > 0 {
			return values
		}
	}
	return nil
}

func attributeValues(resource jsonAPIResource, key string) []string {
	if key == "" {
		return nil
	}
	if value := stringAttr(resource.Attributes, key); value != "" {
		return []string{value}
	}
	values := stringSliceAttr(resource.Attributes, key)
	if len(values) > 0 {
		return values
	}
	return nil
}

func combineParamValues(existing []map[string]string, key string, values []string) []map[string]string {
	if len(existing) == 0 || len(values) == 0 {
		return nil
	}
	out := make([]map[string]string, 0, len(existing)*len(values))
	for _, entry := range existing {
		for _, value := range values {
			clone := make(map[string]string, len(entry)+1)
			for k, v := range entry {
				clone[k] = v
			}
			clone[key] = value
			out = append(out, clone)
		}
	}
	return out
}

func applyRouteParams(path string, params map[string]string) (string, error) {
	parts := splitRoutePath(path)
	for idx, part := range parts {
		if !isRouteParam(part) {
			continue
		}
		name := strings.TrimLeft(part, ":*")
		value := params[name]
		if value == "" {
			return "", fmt.Errorf("missing value for %s", name)
		}
		parts[idx] = value
	}
	return "/" + strings.Join(parts, "/"), nil
}

func splitRoutePath(path string) []string {
	trimmed := strings.TrimSpace(path)
	trimmed = strings.TrimPrefix(trimmed, "#")
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return []string{}
	}
	parts := strings.Split(trimmed, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func isRouteParam(part string) bool {
	return strings.HasPrefix(part, ":") || strings.HasPrefix(part, "*")
}

func resourceCandidatesForSegment(segment string, resourceSet map[string]struct{}) []string {
	if segment == "" {
		return nil
	}
	return resourceCandidatesForToken(segment, resourceSet)
}

func resourceCandidatesForParam(param string, resourceSet map[string]struct{}) []string {
	if param == "" {
		return nil
	}
	base := strings.TrimSuffix(param, "_id")
	base = strings.TrimSuffix(base, "_uuid")
	return resourceCandidatesForToken(base, resourceSet)
}

func resourceCandidatesForToken(token string, resourceSet map[string]struct{}) []string {
	normalized := normalizeToken(token)
	if normalized == "" {
		return nil
	}
	if aliases, ok := clientURLAliases[normalized]; ok {
		return filterResourceCandidates(aliases, resourceSet)
	}
	candidates := []string{}
	if _, ok := resourceSet[normalized]; ok {
		candidates = append(candidates, normalized)
	}
	plural := pluralizeToken(normalized)
	if plural != normalized {
		if _, ok := resourceSet[plural]; ok {
			candidates = append(candidates, plural)
		}
	}
	return dedupeStrings(candidates)
}

func filterResourceCandidates(candidates []string, resourceSet map[string]struct{}) []string {
	out := []string{}
	for _, candidate := range candidates {
		if _, ok := resourceSet[candidate]; ok {
			out = append(out, candidate)
		}
	}
	return dedupeStrings(out)
}

func normalizeToken(token string) string {
	return strings.ToLower(strings.ReplaceAll(token, "_", "-"))
}

func pluralizeToken(token string) string {
	if token == "" {
		return ""
	}
	if strings.HasSuffix(token, "s") {
		return token
	}
	if strings.HasSuffix(token, "y") && len(token) > 1 {
		return token[:len(token)-1] + "ies"
	}
	if strings.HasSuffix(token, "ch") || strings.HasSuffix(token, "sh") || strings.HasSuffix(token, "x") || strings.HasSuffix(token, "z") {
		return token + "es"
	}
	return token + "s"
}

func singularize(token string) string {
	if strings.HasSuffix(token, "ies") && len(token) > 3 {
		return token[:len(token)-3] + "y"
	}
	if strings.HasSuffix(token, "ses") || strings.HasSuffix(token, "xes") || strings.HasSuffix(token, "zes") || strings.HasSuffix(token, "ches") || strings.HasSuffix(token, "shes") {
		return token[:len(token)-2]
	}
	if strings.HasSuffix(token, "s") && len(token) > 1 {
		return token[:len(token)-1]
	}
	return token
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func writeClientURLOutput(cmd *cobra.Command, urls []string) error {
	if urls == nil {
		urls = []string{}
	}
	if getBoolFlag(cmd, "json") {
		payload := map[string][]string{"client": urls}
		return writeJSON(cmd.OutOrStdout(), payload)
	}
	for _, url := range urls {
		fmt.Fprintln(cmd.OutOrStdout(), url)
	}
	return nil
}

func logClientURLError(cmd *cobra.Command, err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(cmd.ErrOrStderr(), err)
}

func resolveClientBaseURL(cmd *cobra.Command) string {
	if value := strings.TrimSpace(os.Getenv("XBE_CLIENT_BASE_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}
	if value := strings.TrimSpace(os.Getenv("CLIENT_BASE_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}
	baseURL := strings.TrimSpace(getStringFlag(cmd, "base-url"))
	if strings.Contains(baseURL, "staging") {
		return "https://staging-client.x-b-e.com"
	}
	return "https://client.x-b-e.com"
}

func clientURL(baseURL, path string) string {
	base := strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + "/#" + path
}

func clientURLCanResolveFromID(resource string) (bool, []clientRouteBinding, error) {
	bindings, err := clientRouteBindingsForResource(resource)
	if err != nil {
		return false, nil, err
	}
	if len(bindings) == 0 {
		return false, nil, nil
	}
	for _, binding := range bindings {
		if !routeBindingIDOnly(resource, binding) {
			return false, nil, nil
		}
	}
	return true, bindings, nil
}

func clientURLsFromID(cmd *cobra.Command, resource, id string, bindings []clientRouteBinding) ([]string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%s id is required", resource)
	}
	baseURL := resolveClientBaseURL(cmd)
	urls := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		params := map[string]string{}
		for _, param := range binding.ParamBindings {
			if !containsString(param.ResourceCandidates, resource) {
				return nil, fmt.Errorf("%s: missing %s", binding.Route.Path, param.Name)
			}
			params[param.Name] = id
		}
		resolved, err := applyRouteParams(binding.Route.Path, params)
		if err != nil {
			return nil, err
		}
		urls = append(urls, clientURL(baseURL, resolved))
	}
	return dedupeStrings(urls), nil
}

func routeBindingIDOnly(resource string, binding clientRouteBinding) bool {
	for _, param := range binding.ParamBindings {
		if !containsString(param.ResourceCandidates, resource) {
			return false
		}
	}
	return true
}

func clientRouteBindingsForResource(resource string) ([]clientRouteBinding, error) {
	resourceMap, err := loadResourceMap()
	if err != nil {
		return nil, err
	}
	resourceSet := make(map[string]struct{}, len(resourceMap.Resources))
	for name := range resourceMap.Resources {
		resourceSet[name] = struct{}{}
	}

	catalog, err := loadClientRoutes()
	if err != nil {
		return nil, err
	}
	bindings := []clientRouteBinding{}
	for _, route := range catalog.Routes {
		if route.Action {
			continue
		}
		binding := buildClientRouteBinding(route, resourceSet)
		if !binding.referencesResource(resource) {
			continue
		}
		bindings = append(bindings, binding)
	}
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Route.Path < bindings[j].Route.Path
	})
	return bindings, nil
}

func computeClientURLRequirements(resource string) (appURLRequirement, bool, error) {
	bindings, err := clientRouteBindingsForResource(resource)
	if err != nil {
		return appURLRequirement{}, false, err
	}
	if len(bindings) == 0 {
		return appURLRequirement{}, false, nil
	}

	resourceMap, err := loadResourceMap()
	if err != nil {
		return appURLRequirement{}, false, err
	}
	spec, ok := resourceMap.Resources[resource]
	if !ok {
		return appURLRequirement{}, false, nil
	}
	attributeSet := map[string]struct{}{}
	for _, field := range appendUniversalFields(spec.Attributes) {
		attributeSet[field] = struct{}{}
	}
	for _, field := range spec.LabelFields {
		attributeSet[field] = struct{}{}
	}
	relations := resourceMap.Relationships[resource]
	requiredRelations := map[string]struct{}{}
	requiredAttrs := map[string]struct{}{}

	for _, binding := range bindings {
		for _, param := range binding.ParamBindings {
			if containsString(param.ResourceCandidates, resource) {
				continue
			}
			if len(param.ResourceCandidates) == 0 {
				field := normalizeParamField(param.Name)
				if field != "" {
					if _, ok := attributeSet[field]; ok {
						requiredAttrs[field] = struct{}{}
					}
				}
				continue
			}
			found := false
			for relName, relSpec := range relations {
				for _, candidate := range param.ResourceCandidates {
					if containsString(relSpec.Resources, candidate) {
						requiredRelations[relName] = struct{}{}
						found = true
					}
				}
			}
			if !found {
				field := normalizeParamField(param.Name)
				if field != "" {
					if _, ok := attributeSet[field]; ok {
						requiredAttrs[field] = struct{}{}
					}
				}
			}
		}
	}

	fields := make([]string, 0, len(requiredRelations)+len(requiredAttrs))
	for rel := range requiredRelations {
		fields = append(fields, rel)
	}
	for attr := range requiredAttrs {
		fields = append(fields, attr)
	}
	sort.Strings(fields)

	include := make([]string, 0, len(requiredRelations))
	for rel := range requiredRelations {
		include = append(include, rel)
	}
	sort.Strings(include)

	return appURLRequirement{Fields: fields, Include: include}, true, nil
}

func normalizeParamField(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), "_", "-")
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func mergeStringSlices(values ...[]string) []string {
	set := map[string]struct{}{}
	for _, slice := range values {
		for _, value := range slice {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			set[value] = struct{}{}
		}
	}
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
