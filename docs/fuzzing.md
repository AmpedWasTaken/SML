# Fuzzing the parser

`internal/engine/fuzz_test.go` carries one target, `FuzzParseNeverPanics`. It feeds arbitrary bytes to `Parse` and, when the parse succeeds, renders the result at width 80, so a panic anywhere on the untrusted-input path — XML decoding, component validation, limit checks, bounded rendering — surfaces as a crash.

The committed corpus lives next to it in `internal/engine/testdata/fuzz/FuzzParseNeverPanics/`.

## Running the target

```bash
go test ./internal/engine -run=^$ -fuzz=FuzzParseNeverPanics -fuzztime=30s
```

`-run=^$` keeps the ordinary tests out of the run; `-fuzztime` is the budget. Ten seconds is enough to cover the corpus and a few thousand mutations; thirty seconds is a reasonable pre-PR run (roughly 300k executions on a laptop). CI does **not** run the fuzzer — it runs the seeds as ordinary tests — so quote the run you did in the pull request.

## Seeds are regression tests

Every file under `testdata/fuzz/FuzzParseNeverPanics/` is executed by a plain `go test ./internal/engine` run, with no `-fuzz` flag: each one shows up as a sub-test named after its file. A seed is therefore a permanent pin, not a fuzzing aid.

| Seed | Pins |
| --- | --- |
| `duplicate-attribute` | A repeated attribute name is rejected rather than silently overwritten. |
| `duplicate-root` | Exactly one `<screen>` root. |
| `unknown-component` | Unknown element names are rejected, not dropped. |
| `namespace-prefix` | Namespaced elements are rejected. |
| `processing-instruction` | XML directives and processing instructions are rejected. |
| `invalid-named-entity` | A named entity the decoder does not know is rejected. |
| `out-of-range-char-entity` | A character reference outside Unicode (`&#x110000;`) is rejected. |
| `nul-byte` | A NUL byte inside text is rejected. |
| `invalid-utf8` | Invalid UTF-8 is rejected before decoding. |
| `unclosed-element` | Truncated markup is rejected. |
| `text-outside-root` | Text outside the root element is rejected. |
| `depth-limit` | Nesting deeper than `maxDepth` is rejected. |
| `node-limit` | More than `maxNodes` elements is rejected. |

Adding a seed:

- Keep it **under 64 KiB**. The target skips longer input before calling `Parse`, and `maxFileSize` is the same 64 KiB, so a larger seed would never reach the parser.
- Name it for the shape it pins, not for a hash — the name becomes the sub-test name in `go test -v` output.
- One file per distinct shape; a seed that only repeats another one adds runtime without adding coverage.
- The corpus format is Go's own: a `go test fuzz v1` header line, then `[]byte("<quoted input>")`. Let the fuzzer write the file when it finds a crash, then rename it.

## When the fuzzer finds something

1. Go writes the crashing input to `testdata/fuzz/FuzzParseNeverPanics/<hash>` and prints the name.
2. Reproduce it: `go test ./internal/engine -run=FuzzParseNeverPanics/<hash>`.
3. Go has already minimised the input when it wrote the file; shrink it further by hand only if the shape is still unclear.
4. Fix the parser, or add the explicit rejection the input deserves — a silent acceptance is as much a finding as a panic.
5. Rename the file to describe the shape, keep it in the corpus, and add a case to the package's tests that asserts the error rather than the absence of a panic.
6. If the input is security-sensitive — resource exhaustion, sanitisation bypass, anything that escapes the trust boundary in [security.md](security.md) — follow [SECURITY.md](../SECURITY.md) instead of opening a public pull request with the reproducer.

## What the target cannot reach

- **One document at a time.** There is no asset include graph, so include cycles and depth limits are covered by the package's ordinary tests, not here.
- **No host data.** `RenderChecked` is called with `nil` assets and no bindings, so paths that need a real asset root or a bound value are out of scope.
- **No filesystem or network access.** The target must stay deterministic and side-effect free; a seed that depends on the environment is a bug in the seed.

## Limits that must stay enforced

The target assumes these limits are in place and only checks that nothing panics; the limits themselves are asserted by the package tests.

| Limit | Value | Where |
| --- | --- | --- |
| View size | 64 KiB | `maxFileSize`, `internal/engine/types.go` |
| Nesting depth | 16 | `maxDepth`, `internal/engine/types.go` |
| Node count | 800 | `maxNodes`, `internal/engine/types.go` |

If a change moves one of those limits, update the table, the seeds that sit on the boundary, and the package tests together.
