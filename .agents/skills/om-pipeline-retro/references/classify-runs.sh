#!/usr/bin/env sh
# Deterministic run classifier, opened by the classify step. Pure stdin -> stdout:
# it reads tracker data, writes a classification, and contacts nothing.
#
# Input: one JSON array of pull requests on stdin, carrying the fields the skill
# requests through the list-prs and get-pr tracker operations:
#
#   number, state, createdAt, mergedAt, closedAt, additions,
#   labels:   [{"name": ...}, ...]
#   reviews:  [{"state": ..., "body": ...}, ...]
#   comments: [{"body": ..., "createdAt": ...}, ...]
#
# Every field except `number` is optional and degrades to a neutral value; a
# field of the wrong type is treated as absent rather than aborting the run.
#
# Classification rules, so an agent that cannot execute this file can apply the
# same rules inline and reach the same verdict:
#
#   1. A run is counted from its OPENING comment: a marker comment whose line
#      says a run started, took over, or was resumed ("started by", "starting
#      <skill> run", "taking over"). One opener line opens exactly ONE run per
#      skill it names, however many times it names it — the 🤖 marker prefix and
#      the "starting `<skill>` run" phrase both carry the same name, and counting
#      occurrences would report a single run as a second pass. Completion notes,
#      label rationales, run summaries and evidence posts belong to the run
#      already open and start nothing. A skill that posts marker comments but
#      never an opener falls back
#      to time clustering: comments more than --gap-minutes apart count as
#      separate runs, and with no timestamps each counts as its own run, which is
#      an upper bound that `timestampCoverage.reliable` reports as false.
#   2. Only agent-authored text is evidence. A body counts when it carries an
#      agent marker; quoted lines and fenced blocks are stripped first, so a human
#      quoting a bot never creates a run or a cause.
#   3. A pull request took a SECOND PASS when any of these hold: two or more
#      review runs, any resume run, any drive-to-merge-ready run, two or more QA
#      runs, or the request was closed without merging.
#   4. A second pass is HARD RECOVERY when the record states a reason: recovery
#      language in agent text, a formal changes-requested review, a declared
#      non-clean Outcome line, or closure without merging.
#   5. It is LOOP CHECKPOINTS only when the loop-mode skills are the ONLY
#      second-pass evidence, since they post checkpoints by design.
#   6. Anything else that took a second pass is UNEXPLAINED: the record shows the
#      cost and not the cause.
#   7. Wall-clock cost runs from opened to merged, or to closed when the request
#      never merged. A request's excess is its time beyond the median clean
#      request, split evenly across the causes it carries so the ranked column
#      sums to the time actually lost. With no clean request in the window there
#      is no baseline: excess is null and causes rank by request count.
#   8. A request still carrying the in-progress label, and not yet merged, is not
#      a finished run. It is reported separately and takes part in no count.
#
# Usage:  cat tracker-data.json | sh references/classify-runs.sh [--gap-minutes 60] [--in-progress-label in-progress]
# Exit:   0 classified · 2 unusable input · 3 jq missing

set -u

GAP_MINUTES=60
IN_PROGRESS_LABEL=in-progress
while [ $# -gt 0 ]; do
  case "$1" in
    --gap-minutes)
      [ $# -ge 2 ] || { echo "classify-runs: --gap-minutes needs a value" >&2; exit 2; }
      GAP_MINUTES="$2"; shift 2 ;;
    --gap-minutes=*) GAP_MINUTES="${1#*=}"; shift ;;
    --in-progress-label)
      [ $# -ge 2 ] || { echo "classify-runs: --in-progress-label needs a value" >&2; exit 2; }
      IN_PROGRESS_LABEL="$2"; shift 2 ;;
    --in-progress-label=*) IN_PROGRESS_LABEL="${1#*=}"; shift ;;
    -h|--help) sed -n '2,/^$/p' "$0"; exit 0 ;;
    *) echo "classify-runs: unknown argument '$1'" >&2; exit 2 ;;
  esac
done
case "$GAP_MINUTES" in
  ''|*[!0-9]*) echo "classify-runs: --gap-minutes must be a whole number of minutes" >&2; exit 2 ;;
esac

command -v jq >/dev/null 2>&1 || {
  echo "classify-runs: jq is required (the tracker descriptor already depends on it)" >&2
  exit 3
}

input=$(cat)
case "$input" in
  '') echo "classify-runs: stdin was empty; expected a JSON array of pull requests" >&2; exit 2 ;;
esac

printf '%s' "$input" | jq --arg gapminutes "$GAP_MINUTES" --arg inprogress "$IN_PROGRESS_LABEL" '

def astext: if type == "string" then . else "" end;
def asarray: if type == "array" then . else [] end;
def asnumber: if type == "number" then . elif type == "string" then (try tonumber catch null) else null end;

# Quoted lines and fenced blocks are other people speaking; drop them first.
def unquoted:
  astext | split("\n")
  | reduce .[] as $line ({ fence: false, out: [] };
      if ($line | test("^[ \t]*```")) then { fence: (.fence | not), out: .out }
      elif .fence or ($line | test("^[ \t]*>")) then .
      else { fence: .fence, out: (.out + [$line]) } end)
  | .out | join("\n");

def markers: [ match("(^|\n)[ \t#*]*🤖[^\n]*"; "g") | .string ];
def marker_skills: [ markers[] | match("om-[a-z0-9-]+"; "g") | .string ] | unique;
def is_agent_text: (markers | length) > 0;

# A marker line that opens a run. The lifecycle verbs come from the claim
# boilerplate of this collection; every other marker line continues a run
# already open. (No apostrophe may appear anywhere below: this jq program is a
# single-quoted shell string, so one would close it mid-script.)
# One opener line opens ONE run for a skill however many times it names it: the
# marker prefix and the "starting `<skill>` run" phrase both carry the name.
def opener_lines:
  [ markers[] | select(test("started by|starting[^\n]*run|taking over|resum(ing|ed)[^\n]*run"; "i")) ];

def stamp:
  if type != "string" or . == "" then null
  else ( sub("\\.[0-9]+"; "")
         | sub("[+-][0-9]{2}:[0-9]{2}$"; "Z")
         | sub("[+-][0-9]{4}$"; "Z")
         | if test("Z$") then . else . + "Z" end
         | try fromdateiso8601 catch null )
  end;

def agent_comments:
  [ (.comments | asarray)[] | select(type == "object")
    | { t: (.createdAt | stamp), body: (.body | unquoted) }
    | select(.body | is_agent_text) ];

def opener_runs($cs):
  [ $cs[] | .body | opener_lines[] | [ match("om-[a-z0-9-]+"; "g").string ] | unique[] ]
  | group_by(.) | map({ key: .[0], value: length }) | from_entries;

def cluster($events; $gap):
  ($events | sort_by(.t // 0))
  | reduce .[] as $e ({ n: 0, last: null };
      if .last == null or $e.t == null or ($e.t - .last) > $gap
      then { n: (.n + 1), last: $e.t }
      else { n: .n, last: $e.t } end)
  | .n;

def runs($cs; $gap):
  (opener_runs($cs)) as $opened
  | [ $cs[] | . as $c | ($c.body | marker_skills)[] | { skill: ., t: $c.t } ]
  | group_by(.skill)
  | map( .[0].skill as $s
         | { key: $s,
             value: (if ($opened[$s] // 0) > 0 then $opened[$s] else cluster(.; $gap) end) } )
  | from_entries;

def recovery_language:
  test("\\bmerge conflicts?\\b|CONFLICTING|conflicts resolved|(?<!un)\\binterrupted\\b|\\bstalled\\b|timed out|\\bcrashed\\b|context (limit|exhaust)|rate limit|failed run|re-?run after"; "i");

def declared_outcomes:
  [ match("(^|\n)[ \t]*(?:[-*][ \t]*)?\\**Outcome:\\**[ \t]*(recovered|blocked|clean)\\b[^\n]*"; "g")
    | .string | sub("^\n"; "") | sub("^[ \t]+"; "") | sub("[ \t]+$"; "") ];

def negated($phrase): test("\\b(no|not|never|without)\\b[^.\n]{0,24}" + $phrase; "i");

def causes($changes_requested):
  . as $t
  | [ (if ($t | test("\\bmerge conflicts?\\b|CONFLICTING|conflicts resolved|\\brebased?\\b"; "i"))
          and (($t | negated("merge conflict")) | not)
       then "base moved under the change" else empty end),
      (if ([ $t | split("\n")[]
             | select(test("self-approval|approve your own|self-review"; "i"))
             | select(test("block|reject|refus|forbid|not allow|not permitted|not possible|unavailable|cannot|can not|impossible"; "i")) ] | length) > 0
       then "review could not be recorded" else empty end),
      (if ($t | test("(?<!un)\\binterrupted\\b|\\bstalled\\b|timed out|\\bcrashed\\b|context (limit|exhaust)|rate limit"; "i"))
       then "run did not finish" else empty end),
      (if $changes_requested or ($t | test("completed: CHANGES REQUESTED|verdict:[^\n]*request changes"; "i"))
       then "change requested by a reviewer" else empty end) ];

def median:
  if length == 0 then null
  else sort as $s | ($s | length) as $n
    | if ($n % 2) == 1 then $s[($n / 2 | floor)]
      else (($s[($n / 2 | floor) - 1] + $s[($n / 2 | floor)]) / 2) end
  end;

def p90:
  if length < 5 then null
  else sort as $s | ($s | length) as $n
    | (0.9 * ($n - 1)) as $i | ($i | floor) as $lo
    | ($s[$lo] + ($i - $lo) * ($s[([$lo + 1, $n - 1] | min)] - $s[$lo]))
  end;

def round1: if . == null then null else (. * 10 | round) / 10 end;

def classify($gap):
  . as $pr
  | (agent_comments) as $cs
  | (runs($cs; $gap)) as $runs
  | ([ $cs[] | .body ] + [ ((.reviews | asarray)[] | select(type == "object") | .body | unquoted) ] | join("\n")) as $text
  | ([ ((.reviews | asarray)[] | select(type == "object") | .state | astext | ascii_upcase) ] | any(. == "CHANGES_REQUESTED")) as $changes_requested
  | ($text | declared_outcomes) as $declared
  | ($pr.state | astext | ascii_upcase) as $state
  | [ ((.labels | asarray)[] | select(type == "object") | .name | astext) ] as $labels
  | [ (if (($runs["om-auto-review-pr"] // 0) >= 2) then "second review round" else empty end),
      (if ((($runs["om-auto-continue-pr"] // 0) + ($runs["om-auto-continue-pr-loop"] // 0)) > 0) then "resumed" else empty end),
      (if (($runs["om-auto-fix-pr"] // 0) > 0) then "driven to merge-ready" else empty end),
      (if (($runs["om-auto-qa-pr"] // 0) >= 2) then "second QA round" else empty end),
      (if ($state == "CLOSED") then "closed unmerged" else empty end) ] as $signals
  | [ (if ((($runs["om-auto-create-pr-loop"] // 0) + ($runs["om-auto-continue-pr-loop"] // 0)) > 0) then "resumed" else empty end) ] as $loop_only
  | ($pr.createdAt | stamp) as $opened
  | (($pr.mergedAt | stamp) // ($pr.closedAt | stamp)) as $finished
  | {
      number: $pr.number,
      class: (
        if ($labels | any(. == $inprogress)) and $state != "MERGED" then "in flight"
        elif ($signals | length) == 0 then "clean"
        elif ($text | recovery_language) or $changes_requested or $state == "CLOSED"
             or ($declared | any(test("Outcome:\\**[ \t]*(recovered|blocked)"; "i"))) then "hard recovery"
        elif (($signals - $loop_only) | length) == 0 then "loop checkpoints"
        else "unexplained" end),
      signals: $signals,
      runs: $runs,
      causes: ($text | causes($changes_requested)),
      declaredOutcomes: $declared,
      hoursToMerge: (if $opened != null and $finished != null and $finished > $opened
                     then (($finished - $opened) / 3600 | round1) else null end),
      additions: ($pr.additions | asnumber),
      labels: $labels
    };

def bucket($rows; $label; $low; $high):
  ($rows | map(select(.additions != null and .additions >= $low and .additions < $high))) as $in
  | ($in | map(select(.class == "hard recovery")) | length) as $hard
  | { addedLines: $label, prs: ($in | length), hardRecovery: $hard,
      hardRecoveryPct: (if ($in | length) == 0 then null else (($hard * 100) / ($in | length) | round) end) };

def summarize($rows; $inflight; $timestamped; $total_marker_comments):
  ($rows | map(select(.class == "clean") | .hoursToMerge) | map(select(. != null))) as $clean_hours
  | (if ($clean_hours | length) == 0 then null else ($clean_hours | median) end) as $baseline
  | ($rows | map(select(.class != "clean"))) as $second
  | ([ $second[]
       | . as $r
       | (if (.causes | length) == 0 then ["not stated"] else .causes end) as $cs
       | $cs[]
       | { cause: ., number: $r.number,
           excess: (if $baseline == null or $r.hoursToMerge == null then null
                    else (([($r.hoursToMerge - $baseline), 0] | max) / ($cs | length)) end) } ]
     | group_by(.cause)
     | map({ cause: .[0].cause, prs: map(.number), prCount: length,
             excessHours: (if $baseline == null then null else (map(.excess // 0) | add | round1) end) })
     | sort_by([- (.excessHours // 0), - .prCount])) as $ranked
  | ($rows | map(select(.additions == null)) | length) as $unsized
  | ($rows | map(select(.hoursToMerge != null)) | length) as $timed
  | {
      prsWithRunMarkers: ($rows | length),
      inFlight: { count: ($inflight | length), prs: ($inflight | map(.number)) },
      byClass: ( reduce ["clean", "hard recovery", "loop checkpoints", "unexplained"][] as $k ({};
          . + { ($k): ( ($rows | map(select(.class == $k))) as $m
                        | ($m | map(.hoursToMerge) | map(select(. != null))) as $h
                        | { count: ($m | length), prs: ($m | map(.number)),
                            medianHoursToMerge: ($h | median | round1),
                            p90HoursToMerge: ($h | p90 | round1) } ) })),
      rankedCauses: $ranked,
      bySize: ([ bucket($rows; "0-200"; 0; 200), bucket($rows; "200-600"; 200; 600), bucket($rows; "600+"; 600; 1e12) ]
               + (if $unsized > 0 then [{ addedLines: "size unknown", prs: $unsized, hardRecovery: null, hardRecoveryPct: null }] else [] end)),
      declaredOutcomeCoverage: {
        prs: ($rows | map(select((.declaredOutcomes | length) > 0)) | length),
        pct: (if ($rows | length) == 0 then 0 else (($rows | map(select((.declaredOutcomes | length) > 0)) | length) * 100 / ($rows | length) | round) end),
        unexplainedSecondPasses: ($rows | map(select(.class == "unexplained")) | length)
      },
      timestampCoverage: { markerComments: $total_marker_comments, withTimestamp: $timestamped,
        reliable: ($total_marker_comments == 0 or $timestamped == $total_marker_comments),
        note: (if $total_marker_comments == 0 then "no marker comment to time"
               elif $timestamped == 0
               then "no marker comment carries a timestamp, so a skill that posts no opening comment could not be clustered: its marker comments counted as separate runs and the counts below are an upper bound"
               elif $timestamped < $total_marker_comments
               then "some marker comments carry no timestamp, so run counts are an upper bound"
               else "every marker comment carried a timestamp" end) },
      fieldCoverage: {
        withTiming: $timed, withSize: (($rows | length) - $unsized),
        note: (if $timed == 0
               then "no request carried both an opening and a finishing timestamp, so every hour figure is null and causes rank by request count alone"
               elif $unsized == ($rows | length)
               then "no request carried a size, so the size table is empty"
               else "timestamps and sizes present" end) },
      cleanMedianHoursToMerge: ($baseline | round1)
    };

if type != "array" then
  ("classify-runs: expected a JSON array of pull requests" | halt_error(2))
else
  (map(select(type == "object"))) as $prs
  | ( [ $prs[] | (.comments | asarray)[] | select(type == "object")
        | { t: (.createdAt | stamp), body: (.body | unquoted) }
        | select(.body | is_agent_text) ] ) as $marker_comments
  | ($marker_comments | length) as $total_marker_comments
  | ($marker_comments | map(select(.t != null)) | length) as $timestamped
  | ( [ $prs[] | select((agent_comments | length) > 0) | classify($gapminutes | tonumber * 60) ] ) as $classified
  | ($classified | map(select(.class == "in flight"))) as $inflight
  | ($classified | map(select(.class != "in flight"))) as $rows
  | if ($rows | length) == 0 and ($inflight | length) == 0
    then { prsWithRunMarkers: 0, note: "no pull request carried an agent run marker" }
    else { summary: summarize($rows; $inflight; $timestamped; $total_marker_comments), pullRequests: $rows }
    end
end
'
