# The 7-Step Methodology

A structured approach to AI-assisted software development using Gas Town.

---

## Overview

This methodology ensures every piece of work follows a complete lifecycle:

```
Plan → Design → Build → Test → Review → Document → Deploy
```

Each phase has **gates** - mandatory questions that must be answered before proceeding. This prevents rushing into implementation without understanding, and ensures quality at every step.

### Why Gates?

Without gates, AI agents (and humans) tend to:
- Jump straight to coding without understanding requirements
- Skip testing because "it works on my machine"
- Forget documentation until it's painful
- Deploy without rollback plans

Gates force deliberate progression through each phase.

---

## Prerequisites

### Required Tools

```bash
# Gas Town CLI
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash

# Verify installation
bd --version
gt --version
```

### MCP Server Setup

MCP (Model Context Protocol) servers extend AI capabilities for testing and validation. The servers you need depend on your project type.

**Discovery Process:**

Before starting work, identify what MCP servers would help:

| Project Type | Recommended MCP Servers |
|--------------|------------------------|
| Web/Frontend | `cursor-browser-extension` - UI testing, visual verification |
| Database | `supabase` - Data operations, schema validation |
| Document Processing | `doclayer` - PDF/document extraction |
| API Services | Custom MCP for endpoint testing |

**Installation:**

MCP servers are configured in Cursor settings. For each server:

1. Open Cursor Settings → MCP
2. Add the server configuration
3. Restart Cursor to activate

Example for browser extension:
```json
{
  "mcpServers": {
    "cursor-browser-extension": {
      "command": "npx",
      "args": ["cursor-browser-extension"]
    }
  }
}
```

> **Note:** Search for specific MCP servers based on your use case. The ecosystem is growing rapidly.

---

## Phase 1: Plan

### Purpose

Understand the problem before solving it. Break down work into actionable items.

### Gate Questions

You **cannot proceed** until these are answered:

| Question | Why It Matters |
|----------|---------------|
| What problem are we solving? | Prevents solving the wrong problem |
| Who are the stakeholders? | Identifies who to consult and notify |
| What are the constraints? | Surfaces blockers early |
| What does success look like? | Defines clear acceptance criteria |
| What is the scope boundary? | Prevents scope creep |
| Are there existing solutions? | Avoids reinventing the wheel |

### Actions

```bash
# Create the planning issue
bd create --title "Plan: <feature name>" --type planning

# If complex, use the planning molecule
bd mol pour mol-plan --var issue=<id>
```

### Exit Criteria

- [ ] Problem statement written and agreed
- [ ] Scope explicitly defined (what's in, what's out)
- [ ] Success criteria are measurable
- [ ] Child issues created for discrete work items
- [ ] Dependencies identified and linked

### Handoff to Design

```bash
# Update issue with planning summary
bd update <id> --notes "Planning complete. See child issues for work breakdown."

# If design is needed, create design issue
bd create --title "Design: <feature name>" --type design --parent <id>
```

---

## Phase 2: Design

### Purpose

Make architectural decisions before writing code. Document trade-offs.

### Gate Questions

| Question | Why It Matters |
|----------|---------------|
| What are the key architectural decisions? | Forces explicit choices |
| What alternatives were considered? | Shows due diligence |
| What are the risks? | Enables mitigation planning |
| How does this integrate with existing code? | Prevents integration surprises |
| What will break if this fails? | Identifies blast radius |
| Is this reversible? | Informs deployment strategy |

### Actions

```bash
# For significant features, run the design convoy
gt formula run design --problem="<problem statement>"

# This spawns parallel analysis across dimensions:
# - API design
# - Data model
# - UX considerations
# - Scalability
# - Security
# - Integration
```

### Exit Criteria

- [ ] Design document exists in `.designs/`
- [ ] Key decisions documented with rationale
- [ ] Risks identified with mitigation plans
- [ ] Design reviewed by stakeholder (if required)
- [ ] No unresolved blocking questions

### Handoff to Build

```bash
# Link design to implementation issue
bd update <impl-issue> --notes "Design: .designs/<design-id>/design-doc.md"
```

---

## Phase 3: Build

### Purpose

Implement the solution according to the plan and design.

### Gate Questions

| Question | Why It Matters |
|----------|---------------|
| Is the design approved? | Prevents building wrong thing |
| Are all dependencies available? | Avoids mid-build blockers |
| Is the branch clean and up-to-date? | Prevents merge conflicts |
| Do I understand what "done" looks like? | Clear target |
| What's the first small step? | Enables incremental progress |

### Actions

```bash
# For polecat workers, the standard workflow
gt hook                    # Check assignment
bd ready                   # Find current step
bd show <step-id>          # Understand the step

# Implementation loop
# 1. Make changes
# 2. Commit frequently
git add <files>
git commit -m "feat: <description> (<issue-id>)"

# 3. Mark step complete
bd close <step-id> --continue
```

### Exit Criteria

- [ ] All planned changes implemented
- [ ] Code compiles/builds without errors
- [ ] No linter errors introduced
- [ ] Changes committed with clear messages
- [ ] Self-review completed (no obvious issues)

### Handoff to Test

```bash
# Ensure all changes are committed
git status  # Should show "working tree clean"
```

---

## Phase 4: Test

### Purpose

Verify the implementation works correctly and doesn't break existing functionality.

### Gate Questions

| Question | Why It Matters |
|----------|---------------|
| What test types are needed? | Unit, integration, e2e, manual? |
| What edge cases matter? | Boundary conditions, error paths |
| What is acceptable coverage? | Quality bar definition |
| What could break that isn't obvious? | Regression risks |
| How do I test this locally? | Verification before CI |

### Actions

```bash
# Run existing tests first
go test ./...

# Check coverage for changed files
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep <changed-file>

# For web apps, use MCP browser extension
# (See MCP Testing section below)
```

### MCP-Assisted Testing

If you have `cursor-browser-extension` configured:

```
Use the browser MCP to:
1. Navigate to the feature URL
2. Interact with UI elements
3. Verify expected behavior
4. Screenshot results for documentation
```

### Exit Criteria

- [ ] All existing tests pass
- [ ] New code has appropriate test coverage
- [ ] Edge cases have test coverage
- [ ] Manual testing completed (if applicable)
- [ ] No regressions introduced

### Handoff to Review

```bash
# Push branch for review
git push -u origin $(git branch --show-current)
```

---

## Phase 5: Review

### Purpose

Quality assurance through systematic code inspection.

### Gate Questions

| Question | Why It Matters |
|----------|---------------|
| Are all tests passing? | Basic quality bar |
| Is code properly documented? | Maintainability |
| Are there security concerns? | Vulnerability prevention |
| Does it follow codebase conventions? | Consistency |
| Is the change minimal and focused? | Reviewability |
| Would I be comfortable maintaining this? | Long-term ownership |

### Actions

```bash
# For comprehensive review, run the code-review convoy
gt formula run code-review --branch=$(git branch --show-current)

# This spawns parallel reviewers examining:
# - Correctness
# - Performance
# - Security
# - Elegance
# - Resilience
# - Style
# - Code smells
```

### Exit Criteria

- [ ] All critical issues resolved
- [ ] Major issues resolved or tracked
- [ ] Minor issues noted for future
- [ ] Security review passed
- [ ] Approval received (if required)

### Handoff to Document

```bash
# Review findings in .reviews/<review-id>/
# Address critical and major issues before proceeding
```

---

## Phase 6: Document

### Purpose

Ensure changes are discoverable and understandable by future readers.

### Gate Questions

| Question | Why It Matters |
|----------|---------------|
| What user-facing changes need documentation? | User communication |
| What API changes need updating? | Developer communication |
| Are inline comments sufficient? | Code understanding |
| Does CHANGELOG need an entry? | Release notes |
| Are there examples that need updating? | Practical guidance |

### Actions

```bash
# Check what documentation exists
ls docs/
cat README.md

# For significant features, document:
# 1. What it does
# 2. How to use it
# 3. Configuration options
# 4. Examples
```

### Documentation Checklist

- [ ] README updated (if user-facing)
- [ ] Inline godoc/JSDoc comments added
- [ ] CHANGELOG entry drafted
- [ ] Examples updated or added
- [ ] Migration guide (if breaking changes)

### Exit Criteria

- [ ] All public APIs documented
- [ ] User-facing changes have usage docs
- [ ] CHANGELOG entry ready

### Handoff to Deploy

```bash
# Commit documentation
git add docs/ README.md CHANGELOG.md
git commit -m "docs: update documentation for <feature>"
```

---

## Phase 7: Deploy

### Purpose

Ship the changes safely and reversibly.

### Gate Questions

| Question | Why It Matters |
|----------|---------------|
| Is the rollback plan ready? | Recovery capability |
| Are monitoring/alerts configured? | Issue detection |
| Who approves this release? | Accountability |
| What's the deployment window? | Timing considerations |
| Who needs to be notified? | Stakeholder communication |
| What could go wrong? | Risk awareness |

### Actions

```bash
# For Gas Town releases, use the release formula
bd mol pour beads-release --var version=<version>

# For standard work, submit to merge queue
gt done
```

### Exit Criteria

- [ ] Changes merged to main
- [ ] CI/CD pipeline passed
- [ ] Deployment successful
- [ ] Monitoring confirms healthy state
- [ ] Stakeholders notified

---

## Debugging Procedures

### Phase-Specific Debugging

| Phase | Common Issues | Debug Approach |
|-------|--------------|----------------|
| Plan | Unclear requirements | Re-interview stakeholders, find existing issues |
| Design | Missing considerations | Run design convoy, review `.designs/` output |
| Build | Implementation stuck | Check `gt mail inbox`, mail Witness for help |
| Test | Flaky tests | Isolate test, check for race conditions |
| Review | Conflicting feedback | Escalate to design phase for decision |
| Document | Outdated docs | Search codebase for recent changes |
| Deploy | Failed deployment | Check CI logs, rollback if needed |

### Real-Time Monitoring

```bash
# Watch work progress
bd activity --follow

# Check specific agent status
gt hook                    # Your current assignment
bd mol current             # Where in the molecule

# Check for stuck work
bd list --status=in_progress --stale=1h
```

### Log Analysis

```bash
# Daemon logs
cat ~/gt/daemon/daemon.log | tail -100

# Agent-specific logs (if configured)
cat ~/gt/<rig>/polecats/<name>/session.log
```

---

## Testing Procedures

### Test Pyramid

```
         /\
        /  \        E2E Tests (few, slow, high confidence)
       /----\
      /      \      Integration Tests (some, medium speed)
     /--------\
    /          \    Unit Tests (many, fast, focused)
   /______________\
```

### Coverage Requirements

| Change Type | Minimum Coverage |
|-------------|-----------------|
| New feature | 80% of new code |
| Bug fix | Regression test required |
| Refactor | Maintain existing coverage |
| Critical path | 90%+ coverage |

### MCP-Assisted Testing

For web applications with `cursor-browser-extension`:

1. **Visual Verification**
   - Navigate to changed pages
   - Verify layout renders correctly
   - Check responsive breakpoints

2. **Interaction Testing**
   - Click buttons, fill forms
   - Verify state changes
   - Check error states

3. **Screenshot Documentation**
   - Capture before/after for review
   - Document edge cases visually

### Running the Test Phase

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/feature/...

# Verbose output for debugging
go test -v ./...
```

---

## Quick Reference

### Gate Checklist (All Phases)

```
□ PLAN    - Problem defined? Scope clear? Success measurable?
□ DESIGN  - Decisions documented? Risks identified? Reviewed?
□ BUILD   - Design approved? Branch clean? Incremental commits?
□ TEST    - Tests pass? Coverage adequate? Edge cases covered?
□ REVIEW  - Security checked? Conventions followed? Approved?
□ DOCUMENT - APIs documented? CHANGELOG updated? Examples work?
□ DEPLOY  - Rollback ready? Monitoring active? Stakeholders notified?
```

### Essential Commands

```bash
# Planning
bd create --title "..." --type planning
bd mol pour mol-plan --var issue=<id>

# Design
gt formula run design --problem="..."

# Build
gt hook && bd ready && bd close <step> --continue

# Test
go test ./...

# Review
gt formula run code-review --branch=<branch>

# Document
# (Manual - update docs/, README, CHANGELOG)

# Deploy
gt done                              # Submit to merge queue
bd mol pour beads-release --var version=<v>  # Full release
```

---

## Next Steps

1. **Install Gas Town** if not already installed
2. **Configure MCP servers** for your project type
3. **Create your first planning issue** using the gates above
4. **Share this document** with your team for consistent process

Questions? Check the [Gas Town documentation](./understanding-gas-town.md) or mail your Witness.
