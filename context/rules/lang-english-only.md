---
paths:
  - "*/cmd/**"
  - "*/source/**"
  - "*/drizzle/**"
---

# Code Language Conventions

This governs the language used in code itself — identifiers, and any runtime-user-facing string a piece of code produces (in particular, error messages) — regardless of what language the team communicates in day to day.

---

## English everywhere, including error messages

- Every identifier, and every string a class or function produces at runtime (error messages above all), is written in English. A mixed-language codebase — English identifiers next to a non-English error message — fragments tooling that assumes one language: full-text search, log aggregation, and error-tracking dashboards that group incidents by message text all silently split a single error type into two buckets the moment one instance's message differs by language.
- This applies to every error message a class throws, not only to identifiers and code structure — a class or function name being in English is not enough on its own if the string it throws or returns at runtime is not.
- It also raises the bar for anyone debugging an incident from a stack trace or log line alone, without the surrounding code, if they don't happen to read the second language.

---

## Non-goals

- This governs code only — it says nothing about comments in conversation, commit messages, or any other artifact outside the code itself.
