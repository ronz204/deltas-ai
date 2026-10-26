---
id: lang-en
installs-as: code-english-only
applies-when: identifiers and runtime strings must be English whatever language the team speaks
paths-default:
  - "*/cmd/**"
  - "*/source/**"
  - "*/drizzle/**"
---

# Code Language Conventions

This governs the language used in code itself: identifiers, and any user-facing string a piece of code produces at runtime, error messages above all, regardless of what language the team communicates in day to day.

---

## English everywhere, including error messages

- Every identifier, and every string a class or function produces at runtime, is written in English.
- This covers every error message a class throws or returns, not only identifiers and structure: an English class name does not make a non-English thrown string acceptable.

---

## Non-goals

- Says nothing about comments in conversation, commit messages or any artifact outside the code itself.

<!-- rationale -->

## Why the rule is shaped this way

- A mixed-language codebase fragments tooling that assumes one language: full-text search, log aggregation and error-tracking dashboards that group incidents by message text silently split one error type into two buckets as soon as one instance's message differs by language.
- It also raises the bar for anyone debugging an incident from a stack trace or log line alone, without the surrounding code, if they do not read the second language.
