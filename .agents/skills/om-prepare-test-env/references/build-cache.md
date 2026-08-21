# Generic build cache — technology-agnostic, embedded in the script

The reusable build-cache mechanism the generated entrypoint embeds. Called from
the `om-prepare-test-env` body / entrypoint contract step 4 during Phase 2
generation; only the three variable lists are project-specific (2.2 discovers
them).

Compiling, codegen, and package builds are the most expensive bootstrap steps.
The cache mechanism is the same for every stack — only the three variable lists
differ, and 2.2 discovers those. Embed this (adapted) in every generated script:

```sh
# --- build cache (generic; only the three lists are project-specific) ---
BUILD_INPUTS="src package.json package-lock.json"   # source dirs, lockfiles, build configs
BUILD_ENV_VARS="NODE_ENV"                           # env vars that shape the build output
ARTIFACTS="dist"                                    # outputs that must exist and be non-empty

fp_file() { stat -f '%z:%m' "$1" 2>/dev/null || stat -c '%s:%Y' "$1" 2>/dev/null; }
fingerprint() {
  {
    for p in $BUILD_INPUTS; do
      if [ -d "$p" ]; then
        find "$p" -type f \
          ! -path '*/node_modules/*' ! -path '*/.git/*' ! -path '*/dist/*' \
          ! -path '*/.cache/*' ! -path '*/coverage/*'
      elif [ -f "$p" ]; then echo "$p"; fi
    done | LC_ALL=C sort | while IFS= read -r f; do printf '%s:%s\n' "$f" "$(fp_file "$f")"; done
    for v in $BUILD_ENV_VARS; do eval "printf 'env:%s=%s\n' \"$v\" \"\${$v:-}\""; done
  } | cksum | awk '{print $1"-"$2}'
}

build_needed() {
  [ "${FORCE_REBUILD:-0}" = 1 ] && return 0
  [ -f "$BUILD_CACHE" ] || return 0
  CACHED_FP=$(sed -n 's/.*"sourceFingerprint": *"\([^"]*\)".*/\1/p' "$BUILD_CACHE")
  CACHED_ROOT=$(sed -n 's/.*"projectRoot": *"\([^"]*\)".*/\1/p' "$BUILD_CACHE")
  [ "$CACHED_FP" = "$(fingerprint)" ] || return 0
  [ "$CACHED_ROOT" = "$(pwd)" ] || return 0          # a worktree never inherits another checkout's cache
  for a in $ARTIFACTS; do [ -s "$a" ] || [ -d "$a" ] || return 0; done
  return 1                                           # cache valid -> skip the build
}

if build_needed; then
  # <project's preparation chain: install -> codegen -> build, discovered in 2.2>
  printf '{ "builtAt": "%s", "sourceFingerprint": "%s", "projectRoot": "%s", "artifactPaths": "%s" }\n' \
    "$(date -u +%FT%TZ)" "$(fingerprint)" "$(pwd)" "$ARTIFACTS" > "$BUILD_CACHE"
fi
```

Adaptation notes for generation:

- `path:size:mtime` per file — cheap, no file reads, portable (BSD `stat -f`
  first, GNU `stat -c` fallback; both tested — Git Bash and WSL2 take the GNU
  branch). Add the repo's own output/cache
  dirs to the `find` exclusions (`target`, `build`, `.next`, `vendor`, …).
- In the PowerShell flavor, implement the same `fingerprint`/`build_needed`
  logic with the equivalents table above (`Get-ChildItem -Recurse` with the
  same exclusions, `size:mtime` per file, a hash of the sorted joined listing,
  `ConvertFrom-Json` for the cache file). The cache file format is identical,
  so the two flavors share `$BUILD_CACHE` state.
- The env fingerprint is part of the hash: a build made under different flags is
  a different build.
- **Bias toward rebuilding**: any unreadable, mismatched, or ambiguous cache
  state returns "build needed" — a fast bootstrap is never worth testing stale
  artifacts. The cache covers preparation/build only; database provisioning,
  migration, and seeding still run fresh per environment.
- State is **per checkout**: a worktree's first boot pays the full chain; the
  cache makes every subsequent boot cheap.
- When the repo's own tooling already implements build caching or env reuse
  (its own state file, reuse flags, cache TTL), the script calls **its**
  mechanism with its flags instead of duplicating it — this block then only
  covers whatever the repo's tooling does not.
