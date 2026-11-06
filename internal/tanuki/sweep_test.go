package tanuki

import (
	"strings"
	"testing"
)

// TestCloakSweepClearsOrgRewriteFranken reproduces the hang-then-refuse the
// proxy hit on a real hunt: a target hostname glued inside a report filename.
// The host-bounded domain rewrite declines the glued host, but the org rewrite
// still replaces the bare org, leaving "api.<fictionorg>.example", which reads
// as a real domain. The staged pipeline regenerates it every pass, so the cloak
// loop could never clear it and refused the body. The direct sweep must clear
// it, and the real org must be gone from the output.
func TestCloakSweepClearsOrgRewriteFranken(t *testing.T) {
	t.Setenv("TANUKI_ON_LEAK", onLeakCloak)

	setupEngagement(t, "realcorp.example")

	body := []byte(`{"model":"m","max_tokens":1,"messages":[{"role":"user",` +
		`"content":[{"type":"tool_result","tool_use_id":"t1",` +
		`"content":"findings in F52_api.realcorp.example_VAPT_Report.md and host admin.realcorp.example"}]}]}`)

	out, err := anonymizeUpstreamBody(body)
	if err != nil {
		t.Fatalf("cloak must never refuse a value it can hide itself: %v", err)
	}

	if strings.Contains(strings.ToLower(string(out)), "realcorp") {
		t.Fatalf("the real org leaked into the forwarded body: %s", out)
	}
}

// TestCloakNeverRefusesUnderCloakPolicy is a broad guard: with a target whose
// org is present both cleanly and glued, the cloak policy forwards rather than
// refusing, and the real org never leaves.
func TestCloakNeverRefusesUnderCloakPolicy(t *testing.T) {
	t.Setenv("TANUKI_ON_LEAK", onLeakCloak)

	setupEngagement(t, "acmetech.example")

	body := []byte(`{"model":"m","max_tokens":1,"messages":[{"role":"user",` +
		`"content":[{"type":"text","text":"acmetech.example, api.acmetech.example, ` +
		`file_sub.acmetech.example_notes.txt, and bare acmetech"}]}]}`)

	out, err := anonymizeUpstreamBody(body)
	if err != nil {
		t.Fatalf("cloak must never refuse: %v", err)
	}

	if strings.Contains(strings.ToLower(string(out)), "acmetech") {
		t.Fatalf("real org leaked: %s", out)
	}
}
