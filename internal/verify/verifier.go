// Package verify implements a deterministic citation gate for LLM output.
//
// The model is instructed to cite retrieved material by id — never to restate it.
// This package then does the restating: every [post:<uuid>] pointer in a generated
// answer is resolved against the exact set of records that were retrieved for this
// user, and replaced verbatim with canonical text drawn from the database.
//
// The effect is that a citation cannot be wrong. A pointer either resolves to a
// record the user is entitled to see, or the whole answer is rejected. The model
// chooses *which* sources to reference; it never supplies their contents.
//
// This is deliberately not a scoring or filtering approach. There is no threshold to
// tune and no probability to trade off — an unresolvable pointer is a hard failure.
package verify

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrUnverifiedCitation reports an answer referencing material that was not in the
// retrieved set. It means the model invented a source, or reached for one belonging
// to someone else.
var ErrUnverifiedCitation = errors.New("verify: answer cited material outside the retrieved set")

// SafeFallback replaces any answer that fails verification. It is returned to the
// caller instead of the model's text, never alongside it.
const SafeFallback = "I couldn't ground that answer in your tracked sources, so I'd rather not guess. Try narrowing the question, or check that the relevant source is still being followed."

// citationRe matches [post:<uuid>] pointers. The uuid shape is enforced here so that
// malformed pointers fail closed rather than being silently ignored.
var citationRe = regexp.MustCompile(`\[post:([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})\]`)

// Source is one record the model was allowed to draw on. These come from the
// user-scoped retrieval query, so membership in this set is also the authorization
// check — a pointer to a post the user does not follow cannot resolve.
type Source struct {
	ID    string
	Title string
	URL   string
}

// Result is the outcome of verifying one generated answer.
type Result struct {
	// Answer is the text to show the caller: either the model's answer with every
	// pointer replaced by canonical text, or SafeFallback if verification failed.
	Answer string

	// Verified reports whether every pointer resolved.
	Verified bool

	// Cited lists the ids that resolved, in order of first appearance. Useful for
	// returning structured references alongside the prose.
	Cited []string

	// Rejected lists pointers that did not resolve. Non-empty exactly when
	// Verified is false, and worth logging — a persistent id here usually means a
	// prompt problem, not an attack.
	Rejected []string
}

// Gate verifies citations against a Resolver.
type Gate struct {
	resolver Resolver
}

// NewGate builds a gate over the records retrieved for one request. Pass exactly the
// rows the retrieval query returned; anything absent here is unciteable by construction.
func NewGate(sources []Source) *Gate {
	return &Gate{resolver: NewSetResolver(sources)}
}

// NewGateWithResolver builds a gate over any Resolver — a shared corpus table, a
// remote service, a cache. Use this when the citeable set is not a per-request slice.
func NewGateWithResolver(r Resolver) *Gate {
	return &Gate{resolver: r}
}

// Verify checks every citation in answer and substitutes canonical text for each.
//
// An answer containing no citations is rejected: the model was asked to ground its
// response, and an ungrounded answer is the failure mode this package exists to catch.
// Returning it unverified would defeat the gate.
func (g *Gate) Verify(answer string) (Result, error) {
	return g.VerifyContext(context.Background(), answer)
}

// VerifyContext is Verify with a context, for resolvers that perform I/O.
func (g *Gate) VerifyContext(ctx context.Context, answer string) (Result, error) {
	matches := citationRe.FindAllStringSubmatch(answer, -1)
	if len(matches) == 0 {
		return Result{Answer: SafeFallback, Verified: false}, ErrUnverifiedCitation
	}

	var cited, rejected []string
	seen := make(map[string]bool)

	for _, m := range matches {
		id := strings.ToLower(m[1])
		if _, ok := g.resolver.Resolve(ctx, id); !ok {
			if !seen[id] {
				rejected = append(rejected, id)
				seen[id] = true
			}
			continue
		}
		if !seen[id] {
			cited = append(cited, id)
			seen[id] = true
		}
	}

	// Fail the whole answer, not just the offending sentence. A response that mixes
	// verified and invented sources is more dangerous than one that is plainly wrong,
	// because the verified parts lend it credibility.
	if len(rejected) > 0 {
		return Result{
			Answer:   SafeFallback,
			Verified: false,
			Cited:    cited,
			Rejected: rejected,
		}, fmt.Errorf("%w: %v", ErrUnverifiedCitation, rejected)
	}

	substituted := citationRe.ReplaceAllStringFunc(answer, func(tag string) string {
		id := strings.ToLower(citationRe.FindStringSubmatch(tag)[1])
		e, ok := g.resolver.Resolve(ctx, id)
		if !ok {
			// Unreachable: every id resolved in the pass above. Preserve the raw tag
			// rather than emitting empty text if a resolver is ever non-deterministic.
			return tag
		}
		return e.Render()
	})

	return Result{Answer: substituted, Verified: true, Cited: cited}, nil
}
