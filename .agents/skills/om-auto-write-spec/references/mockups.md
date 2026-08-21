# UI mockups and current-app screenshots (step 5)

Visual evidence for a UI-facing spec: what the affected screens look like **today**, and what the spec proposes they look like. Both end up on the spec PR via **attach-image-evidence**.

## Preconditions

- `om-prepare-test-env` has produced (or can produce) the shared test-env descriptor, and `.ai/agentic.config.json` names a browser provider whose descriptor exists at `.ai/browsers/<provider>.md`.
- When either is missing: **skip visuals entirely**, add a `Mockups: skipped — {reason}` line to the PR body, and continue. A text-only spec PR is a valid outcome; a failed run is not.

## 1. Current-state screenshots

1. Boot the app through the test-env descriptor (same flow as `om-auto-qa-pr` step 4 — reuse a healthy running env when the descriptor says it is fresh).
2. From the spec's UI/UX section, list the existing screens/flows the feature touches (routes, admin pages, components). Cap at the 3–6 most relevant screens.
3. Drive each screen with the browser-provider operations (open → wait for load → screenshot) into `${SPECS_DIR}/assets/${SLUG}/current-NN-<screen>.png`. Verify each PNG is non-empty.
4. When the app cannot boot or a screen errors, capture what you can and note the gaps — partial evidence beats none.

## 2. Proposed-UI mockups

For each **new or materially changed** screen in the spec:

1. Author a minimal **static HTML file** in `${SPECS_DIR}/assets/${SLUG}/mockup-NN-<screen>.html` — self-contained (inline CSS, no build step, no app code, no external requests). Match the app's rough look (reuse its visible palette/typography from the current-state screenshots) but keep it obviously illustrative; realistic placeholder data, never real user data.
2. Render it with the browser provider (`open file://…` → screenshot) into `mockup-NN-<screen>.png`.
3. Reference each mockup from the spec's UI/UX section by relative path so the spec document itself shows the visuals.

Keep it cheap: mockups exist to communicate layout and flow, not pixel-perfect design. 2–4 mockups is the normal ceiling; skip mockups for standard CRUD the spec explicitly calls standard.

## 3. Publish

Commit the `assets/${SLUG}/` folder with the spec, then (after the PR exists) post one evidence comment via **attach-image-evidence**: `{prNumber}`, a short table mapping each image to its screen + current/proposed role, slug `spec-${SLUG}`, and the PNG paths. The tracker descriptor owns making images render inline; when it cannot, it posts links — surface the limitation and move on.
