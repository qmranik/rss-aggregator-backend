package verify

import (
	"context"
	"strings"
)

// Resolver supplies canonical text for a citation id.
//
// This is the seam between the gate and whatever holds the truth. Two authorization
// models already exist behind it:
//
//   - Retrieval-scoped (SetResolver): the citeable set is whatever a user-scoped query
//     returned for this request. Authorization is set membership, so a pointer to
//     another user's record simply does not resolve.
//   - Corpus-scoped (a DB-backed resolver): the citeable set is a shared, admin-written
//     table with row-level security. Authorization lives in the database.
//
// Resolve must be conservative: return ok=false whenever the id is unknown, malformed,
// or not permitted for this caller. Never return a best guess — the gate treats ok=true
// as proof the text is canonical.
type Resolver interface {
	Resolve(ctx context.Context, id string) (Entry, bool)
}

// Entry is the canonical rendering of one citeable record.
type Entry struct {
	Title string
	URL   string
}

// Render produces the substituted text for a citation.
func (e Entry) Render() string {
	if e.URL == "" {
		return e.Title
	}
	return "[" + e.Title + "](" + e.URL + ")"
}

// SetResolver resolves against a fixed set held in memory — the retrieval-scoped case.
// It is what NewGate builds internally, and is the right resolver whenever the citeable
// set is small, per-request, and already filtered by the caller's identity.
type SetResolver struct {
	byID map[string]Entry
}

// NewSetResolver builds a resolver over exactly the records retrieved for one request.
func NewSetResolver(sources []Source) *SetResolver {
	byID := make(map[string]Entry, len(sources))
	for _, s := range sources {
		byID[strings.ToLower(s.ID)] = Entry{Title: s.Title, URL: s.URL}
	}
	return &SetResolver{byID: byID}
}

// Resolve implements Resolver.
func (r *SetResolver) Resolve(_ context.Context, id string) (Entry, bool) {
	e, ok := r.byID[strings.ToLower(id)]
	return e, ok
}
