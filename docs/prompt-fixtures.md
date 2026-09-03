# Prompt test fixtures

Fixtures live in [`src/testdata/prompts/`](../src/testdata/prompts/). They exercise the lint engine (`Run` / `Apply`), CLI report output, and desktop Lint workspace behavior (score, sections, suggestions, approve toggles, diff).

Run from `src/`:

```bash
go run ./cmd/lint testdata/prompts/<fixture>.txt
go run ./cmd/lint -fix testdata/prompts/<fixture>.txt
go run ./cmd/lint -fix -approve-near-duplicates -approve-json-conversion \
  testdata/prompts/<fixture>.txt
```

Regression cases in `internal/regression` currently cover a subset of the original fixtures. Newer suite files below are primarily for manual / exploratory coverage until wired into that harness.

## Original single-issue fixtures

| File | Aspect | Expected signal |
|------|--------|-----------------|
| `exact_duplicates.txt` | Exact instruction repeats | Exact-dupe groups; `-fix` collapses them; distinct constraints survive |
| `near_duplicates.txt` | Cosmetic rewordings in System | Near-dupe groups; unchanged without `-approve-near-duplicates` |
| `example.txt` | Soft near-phrasing (markdown variants) | Near-style phrasing left alone at high confidence; nothing silently dropped |
| `redundant_examples.txt` | Exact + near few-shot blocks | Exact whole-block collapse; near blocks need approval; distinct Q/A survive |
| `structured_data.txt` | `Key: value` runs in System | Structured-data suggestion; JSON rewrite only with `-approve-json-conversion` |
| `long_history.txt` | Conversation history section | History present in section counts; question/system survive Apply |

## Robust suite (multi-aspect / edge)

| File | Aspect | Expected signal |
|------|--------|-----------------|
| `clean_baseline.txt` | Healthy prompt / score floor | ~100 score, 0% savings, no suggestions |
| `constraint_dense.txt` | Dense must-survive rules, little waste | No collapse; JSON / no-markdown / no-explain stay intact |
| `approval_matrix.txt` | Approve-toggle matrix | Near instructions + exact/near examples + structured JSON. `-fix` alone collapses exact examples only; both approve flags also fold near instructions/examples and JSON |
| `kitchen_sink.txt` | Full report surface | Exact + near + duplicate examples + two structured blocks + compress-history in one prompt |
| `shared_line_across_examples.txt` | Example-boundary safety | Shared `Notes:` line across different examples must not corrupt blocks on Apply |
| `multi_structured_blocks.txt` | Multiple JSON opportunities | Two `Key: value` blocks; similar phones may near-match — approving near-dupes is unsafe here |
| `history_over_threshold.txt` | History flagger (>10 turns) | “Compress history” suggestion |
| `history_under_threshold.txt` | History below threshold | No history suggestion |
| `unlabeled_with_dupes.txt` | Headerless system + labeled question | Exact instruction dupes; `Question:` still splits correctly |
| `question_only.txt` | Minimal / splitter edge | Whole text lands in System (no `Question:` header) |
| `nested_headers_in_history.txt` | Splitter trap | `Example N:` / `Question:` lines inside History get re-parsed as live sections |
| `code_context_review.txt` | Code / long History weight | Clean System / History / Question split; no false example promotion |

## Suggested manual pass (desktop or CLI)

1. **Baseline** — `clean_baseline.txt` then `kitchen_sink.txt` (empty vs full UI/report).
2. **Toggles** — `approval_matrix.txt` with neither approve flag, then near only, then both.
3. **Safety** — `shared_line_across_examples.txt` and `constraint_dense.txt` through `-fix` (and with approvals) and confirm required text survives.
4. **Edges** — `question_only.txt`, `nested_headers_in_history.txt`, `history_over_threshold.txt` vs `history_under_threshold.txt`.

## Confidence tiers (reminder)

- **Exact duplicates** (instructions or whole example blocks): high confidence → auto-applied by `-fix` / Apply.
- **Near duplicates** and **JSON conversion**: low confidence → require explicit approval flags / UI toggles.
- **History**: flagged when turn count exceeds the threshold; compression rewrite is not applied yet.
