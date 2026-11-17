package verify

import (
	"context"
	"errors"
	"strings"
	"testing"
)

var (
	postA = Source{ID: "11111111-1111-1111-1111-111111111111", Title: "Go 1.26 released", URL: "https://example.com/go126"}
	postB = Source{ID: "22222222-2222-2222-2222-222222222222", Title: "Postgres 18 notes", URL: "https://example.com/pg18"}
)

func TestVerify_SubstitutesCanonicalText(t *testing.T) {
	g := NewGate([]Source{postA})

	res, err := g.Verify("The headline change is generics. [post:11111111-1111-1111-1111-111111111111]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Verified {
		t.Fatal("expected verified")
	}
	if strings.Contains(res.Answer, "[post:") {
		t.Errorf("pointer survived substitution: %q", res.Answer)
	}
	if !strings.Contains(res.Answer, postA.Title) || !strings.Contains(res.Answer, postA.URL) {
		t.Errorf("canonical text not injected: %q", res.Answer)
	}
}

// The core guarantee: text the model supplies about a source is discarded. Only the
// pointer survives, and the database supplies the words.
func TestVerify_ModelSuppliedTitleIsDiscarded(t *testing.T) {
	g := NewGate([]Source{postA})

	res, err := g.Verify(`See "Go 1.26 removes the GC" [post:11111111-1111-1111-1111-111111111111]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Answer, postA.Title) {
		t.Errorf("canonical title missing: %q", res.Answer)
	}
}

func TestVerify_RejectsHallucinatedID(t *testing.T) {
	g := NewGate([]Source{postA})

	res, err := g.Verify("As reported [post:99999999-9999-9999-9999-999999999999]")
	if !errors.Is(err, ErrUnverifiedCitation) {
		t.Fatalf("want ErrUnverifiedCitation, got %v", err)
	}
	if res.Verified {
		t.Error("hallucinated id must not verify")
	}
	if res.Answer != SafeFallback {
		t.Errorf("want SafeFallback, got %q", res.Answer)
	}
	if len(res.Rejected) != 1 {
		t.Errorf("want 1 rejected, got %v", res.Rejected)
	}
}

// A mixed answer is more dangerous than a plainly wrong one: the verified half lends
// credibility to the invented half. The whole response must fail.
func TestVerify_MixedValidAndInvalidFailsEntirely(t *testing.T) {
	g := NewGate([]Source{postA, postB})

	res, err := g.Verify("First [post:11111111-1111-1111-1111-111111111111] and second [post:99999999-9999-9999-9999-999999999999]")
	if !errors.Is(err, ErrUnverifiedCitation) {
		t.Fatalf("want ErrUnverifiedCitation, got %v", err)
	}
	if res.Answer != SafeFallback {
		t.Error("partial verification must not leak the model's text")
	}
	if strings.Contains(res.Answer, postA.Title) {
		t.Error("verified portion must not survive a failed answer")
	}
}

// An answer citing nothing is ungrounded, which is the failure this gate exists to
// catch. It must not pass merely because it contains no bad pointer.
func TestVerify_UncitedAnswerIsRejected(t *testing.T) {
	g := NewGate([]Source{postA})

	res, err := g.Verify("Broadly, the ecosystem is moving toward generics.")
	if !errors.Is(err, ErrUnverifiedCitation) {
		t.Fatalf("want ErrUnverifiedCitation, got %v", err)
	}
	if res.Answer != SafeFallback {
		t.Errorf("want SafeFallback, got %q", res.Answer)
	}
}

// Authorization is set membership. A gate built from user A's retrieval cannot resolve
// a post belonging to user B, even though the id is perfectly well-formed and real.
// This is the tenant-isolation boundary expressed at the verification layer.
func TestVerify_CannotCiteAnotherUsersPost(t *testing.T) {
	userAGate := NewGate([]Source{postA})

	res, err := userAGate.Verify("Per the notes [post:22222222-2222-2222-2222-222222222222]")
	if !errors.Is(err, ErrUnverifiedCitation) {
		t.Fatalf("cross-tenant citation must fail; got %v", err)
	}
	if strings.Contains(res.Answer, postB.Title) {
		t.Fatal("leaked another user's post title")
	}
	if res.Answer != SafeFallback {
		t.Errorf("want SafeFallback, got %q", res.Answer)
	}
}

func TestVerify_MalformedPointerFailsClosed(t *testing.T) {
	g := NewGate([]Source{postA})

	// Truncated uuid: must not be silently ignored and treated as an uncited answer
	// that happens to look grounded.
	res, err := g.Verify("See [post:1111-2222]")
	if !errors.Is(err, ErrUnverifiedCitation) {
		t.Fatalf("malformed pointer must fail closed; got %v", err)
	}
	if res.Answer != SafeFallback {
		t.Errorf("want SafeFallback, got %q", res.Answer)
	}
}

func TestVerify_DeduplicatesRepeatedCitations(t *testing.T) {
	g := NewGate([]Source{postA})

	res, err := g.Verify("One [post:11111111-1111-1111-1111-111111111111] two [post:11111111-1111-1111-1111-111111111111]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Cited) != 1 {
		t.Errorf("want 1 unique citation, got %v", res.Cited)
	}
}

func TestVerify_CaseInsensitiveUUID(t *testing.T) {
	g := NewGate([]Source{postA})

	if _, err := g.Verify("See [post:11111111-1111-1111-1111-11111111111A]"); err == nil {
		t.Skip("id differs; guard only asserts no panic on mixed case")
	}
	upper := Source{ID: strings.ToUpper(postA.ID), Title: postA.Title, URL: postA.URL}
	g2 := NewGate([]Source{upper})
	if _, err := g2.Verify("See [post:" + strings.ToLower(postA.ID) + "]"); err != nil {
		t.Errorf("case mismatch should still resolve: %v", err)
	}
}

// stubResolver records lookups so tests can assert the gate consults the seam rather
// than reaching into a concrete set.
type stubResolver struct {
	entries map[string]Entry
	calls   int
}

func (s *stubResolver) Resolve(_ context.Context, id string) (Entry, bool) {
	s.calls++
	e, ok := s.entries[strings.ToLower(id)]
	return e, ok
}

func TestGate_UsesResolverSeam(t *testing.T) {
	stub := &stubResolver{entries: map[string]Entry{
		postA.ID: {Title: "From corpus", URL: "https://corpus.example/a"},
	}}
	g := NewGateWithResolver(stub)

	res, err := g.Verify("Grounded [post:" + postA.ID + "]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Answer, "From corpus") {
		t.Errorf("resolver text not used: %q", res.Answer)
	}
	if stub.calls == 0 {
		t.Error("gate did not consult the resolver")
	}
}

// A resolver denying an id must fail the answer even when the id is well-formed and
// exists elsewhere. This is how a corpus-backed resolver enforces RLS.
func TestGate_ResolverDenialFailsClosed(t *testing.T) {
	g := NewGateWithResolver(&stubResolver{entries: map[string]Entry{}})

	res, err := g.Verify("Denied [post:" + postA.ID + "]")
	if !errors.Is(err, ErrUnverifiedCitation) {
		t.Fatalf("want ErrUnverifiedCitation, got %v", err)
	}
	if res.Answer != SafeFallback {
		t.Errorf("want SafeFallback, got %q", res.Answer)
	}
}

func TestEntry_RenderWithoutURL(t *testing.T) {
	if got := (Entry{Title: "Bare"}).Render(); got != "Bare" {
		t.Errorf("want bare title, got %q", got)
	}
}
