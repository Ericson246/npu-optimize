# ADR-000: Architectural Decision Record Process

**Date:** 2026-06-14
**Status:** Accepted
**Repo:** `github.com/Ericson246/npu-optimize`

## Context

As the project grows, architectural decisions need to be documented systematically. Without a record, decisions are lost, new contributors cannot understand why things are the way they are, and the project lacks architectural traceability.

## Decision

ADR (Architectural Decision Record) process inspired by the [ADR GitHub organization](https://github.com/adr) and adapted to the project's size.

### What needs an ADR

An ADR is needed when a decision:
- Affects the public API (CLI flags, JSON schema, exit codes)
- Introduces a new package or internal interface
- Changes dependencies or tooling (CI/CD, linters, release process)
- Defines data formats or communication protocols
- Is revisited (replaces or deprecates a previous ADR)

An ADR is NOT needed for:
- Internal refactoring that doesn't change interfaces
- Bug fixes
- New tests
- Documentation updates without architectural impact
- Dependency updates

### ADR Lifecycle

ADR states (simplified, no RFC overhead):

```
[Propuesto] → [Aceptado] → [Obsoleto]
```

| State | Meaning |
|:------|:--------|
| **Propuesto** | Proposed but not yet approved. Open for discussion in the PR |
| **Aceptado** | Accepted. The decision is in effect and should be followed |
| **Obsoleto** | Superseded by a newer ADR. Must reference the replacing ADR |

### ADR Numbering

Consecutive numbers: `ADR-000`, `ADR-001`, `ADR-002`, etc.

ADR-000 is reserved to document the process itself.

### ADR Template

```markdown
# ADR-NNN: Title

**Date:** YYYY-MM-DD
**Status:** [Propuesto | Aceptado | Obsoleto]
**Replaces:** ADR-MMM (optional, only if status=Obsoleto)

## Context

Describe the problem and why a decision is needed.
Include relevant research and considered information.

## Decision

Describe the decision clearly. Include code snippets, diagrams, or
references if necessary.

## Consequences

What changes as a result of this decision?
What should the team/project do differently?

## Alternatives Considered

Optional section listing alternatives and why they were discarded.

## References

Links to related ADRs, external docs, or relevant code.
```

### ADR File Naming

`ADR-NNN-title-with-hyphens.md`

Examples:
- `ADR-000-adr-process.md`
- `ADR-001-npu-optimize-architecture.md`
- `ADR-002-benchmark-and-extrapolation.md`

### Where to store ADRs

All ADRs in `docs/` at the repository root.

### ADR Review Process

1. Create the ADR in a branch (following the template)
2. Open a PR with the ADR and any related implementation code
3. Discussion in the PR
4. The PR author updates the ADR based on feedback, or the ADR is closed if rejected
5. When merged → status changes from `Propuesto` to `Aceptado`
6. If a later ADR supersedes this one, its status changes to `Obsoleto` referencing the new ADR

## Consequences

- All new ADRs must follow this format
- ADR-000 is the exception: it documents the process itself and its own template
- The number of ADRs should be proportional to the project's complexity

## References

- [ADR GitHub organization](https://github.com/adr)
- [Documenting architecture decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
