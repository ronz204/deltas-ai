# Rule Catalog

Conventions a project may adopt as `.claude/rules/*.md`. The surveyor reads this table alone, proposes the entries that match what it found, and reads an entry's file only to instantiate it once the user confirms.

| Entry | Adopt when | Pairs with |
|---|---|---|
| `ddd-core` | The project keeps a domain layer split into bounded contexts (DDD tactical) | `ddd-app`, one `lang-*` entry |
| `ddd-app` | An application layer orchestrates use cases over those bounded contexts | `ddd-core`, one `lang-*` entry |
| `lang-ts` | TypeScript source under a DDD layout | `ddd-core`, `ddd-app` |
| `lang-py` | Python source under a DDD layout | `ddd-core`, `ddd-app` |
| `lang-en` | Identifiers and runtime strings must be English whatever language the team speaks | none |

## Format of an entry

- The frontmatter holds `id`, `installs-as` (the rule's file name in a project), `applies-when`, and `paths-default`: the layout the entry was written against. The last is a hint to propose from, never a decision.
- Everything above the `<!-- rationale -->` line is the rule, written without folder names. Everything below it is the reasoning, for people and for tuning the rule; it is never copied into a project.
- To instantiate: copy the rule verbatim, replace the frontmatter with a `paths:` list of folders that exist or that the user stated, and name the file `installs-as` plus `.md`. Those names follow the shape of the project's other rules: an area word, then one or two topic words, hyphenated, about 14 to 18 characters.
