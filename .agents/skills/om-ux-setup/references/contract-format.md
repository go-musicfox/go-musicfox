# Contract format — .uxproof/

The manual-extraction fallback writes the same four files `npx uxproof init`
generates. Shapes below are the minimum; extra fields are allowed.

## contract.json

```json
{
  "version": 1,
  "framework": "next | react | vite-react | unknown",
  "styling": { "system": "tailwind | css | css-in-js", "tokenSources": [{ "file": "src/app/globals.css", "tokenCount": 42 }] },
  "componentRoots": [{ "dir": "src/components", "fileCount": 12 }],
  "nativeEquivalents": [{ "element": "button", "component": "Button" }],
  "archetypes": [{ "kind": "list", "count": 8, "examples": ["src/app/orders/page.tsx"] }],
  "counts": { "tokens": 42, "colorTokens": 18, "components": 12 }
}
```

## tokens.json

Array of `{ "name": "primary", "value": "#111827", "kind": "color | size | shadow | font | alias | other", "source": "src/app/globals.css" }`.

## components.json

Array of `{ "name": "Button", "file": "src/components/Button.tsx", "root": "src/components" }`.

## conventions.md

Human-readable house rules with generated sections (stack, tokens, components,
screen shapes) and a fenced manual section:

```markdown
<!-- uxproof:manual-start -->
The team's judgment calls live here and survive regeneration.
<!-- uxproof:manual-end -->
```

The manual section is the local-override surface for every UX skill: it
extends the generated rules and, on conflict, wins.
