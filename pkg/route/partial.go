package route

import (
	"net/http"

	"github.com/cloudretic/router/pkg/middleware"
	"github.com/cloudretic/router/pkg/path"
	"github.com/cloudretic/router/pkg/router/params"
)

// =====PARTS=====

// partialEndPart implements Part to match against a specific subPart repeatedly, with a given optional route parameter.
type partialEndPart struct {
	param   string
	subPart Part
}

// parse a partialEndPart from a token.
func parse_partialEndPart(token string) (*partialEndPart, error) {
	result := &partialEndPart{}
	// get subToken from token (exclude +)
	subToken := token[:len(token)-1]
	// if subToken is empty, use an unqualified anyWord
	if subToken == "" {
		result.subPart = &regexPart{"", regexp_anyWord_compiled}
		return result, nil
	}
	// otherwise, parse out subToken
	subPart, err := parse(subToken)
	if err != nil {
		return nil, err
	}
	// If the subpart has a parameter, move it to the result.
	// This has no real effect if the subPart has an empty parameter (intended behavior)
	if subPartWithParam, ok := subPart.(paramPart); ok {
		result.param = subPartWithParam.ParameterName()
		subPartWithParam.SetParameterName("")
	}
	result.subPart = subPart

	return result, nil
}

// partialEndPart assumes that it's starting at the first partial token.
// For example, in route /file/[filename]{.+}, partialEndPart will start on any token after file
func (part *partialEndPart) Match(req *http.Request, token string) *http.Request {
	req = part.subPart.Match(req, token)
	if req == nil {
		return nil
	}
	if part.param == "" {
		return req
	}
	// If there's a match, get the current path from params and append the token
	currentPath, _ := params.Get(req, part.param)
	req = params.Set(req, part.param, currentPath+"/"+token)
	return req
}

func (part *partialEndPart) ParameterName() string {
	return part.param
}

func (part *partialEndPart) SetParameterName(s string) {
	part.param = s
}

// =====ROUTE=====

// Convenience function to determine if a route expression is partial.
func isPartialRouteExpr(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '+'
}

// partialRoute is specialized to allow routes that may match on extensions, rather than on
// an exact match
type partialRoute struct {
	origExpr string
	mws      []middleware.Middleware
	parts    []Part
}

// Tokenize and parse a route expression into a partialRoute.
//
// See interface Route.
func build_partialRoute(expr string) (*partialRoute, error) {
	tokens := path.TokenizeString(expr)
	route := &partialRoute{
		origExpr: expr,
		mws:      make([]middleware.Middleware, 0),
		parts:    make([]Part, 0),
	}
	for i, token := range tokens {
		var part Part
		var err error
		if i < len(tokens)-1 {
			part, err = parse(token)
		} else {
			part, err = parse_partialEndPart(token)
		}
		if err != nil {
			return nil, err
		} else {
			route.parts = append(route.parts, part)
		}
	}
	return route, nil
}

// Get a string value unique to the route.
//
// See interface Route.
func (route *partialRoute) Hash() string {
	return route.origExpr
}

// Get the length of the route.
// For partialRoutes, this is the number of *absolute* parts; the adaptive part at the end is excluded.
// This ensures that when matching for longest route, the more specialized route is always picked.
//
// See interface Route.
func (route *partialRoute) Length() int {
	return len(route.parts) - 1
}

// Attach middleware to the route. Middleware is handled in attachment order.
//
// See interface Route.
func (route *partialRoute) Attach(mw middleware.Middleware) {
	route.mws = append(route.mws, mw)
}

// Match a request and update its context.
// If the request path is longer than the route, partialRoute will do two things:
//   - Check each token beyond the last against the last Part
//   - If the last part is a Wildcard, stores the leftover route as the parameter
//
// See interface Route.
func (route *partialRoute) MatchAndUpdateContext(req *http.Request) *http.Request {
	req = req.Clone(req.Context())
	// check length; tokens should be > parts
	tokens := path.TokenizeString(req.URL.Path)
	if len(tokens) < len(route.parts)-1 {
		return nil
	}
	// Run any attached middleware
	for _, mw := range route.mws {
		if req = mw(req); req == nil {
			return nil
		}
	}
	// Iterate through tokens and match on the corresponding part, or the last part when extending past
	for i, token := range tokens {
		// Match against current part, or last available if tokens are extending past the length of the route
		var p Part
		if i < len(route.parts) {
			p = route.parts[i]
		} else {
			p = route.parts[len(route.parts)-1]
		}
		if req = p.Match(req, token); req == nil {
			return nil
		}
	}
	// If there were no empty tokens to begin with, run the last rou
	return req
}