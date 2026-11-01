# Rule Catalog

Conventions a project may adopt as `.claude/rules/*.md`. The surveyor reads this table alone, proposes the entries that match what it found, and reads an entry's file only to instantiate it once the user confirms.

| Entry | Adopt when | Pairs with |
|---|---|---|
| `code-domain-layer` | The project keeps a domain layer split into bounded contexts (DDD tactical) | `code-app-usecases`, one `code-*-idiomatic` entry |
| `code-app-usecases` | An application layer orchestrates use cases over those bounded contexts | `code-domain-layer`, one `code-*-idiomatic` entry |
| `code-ts-idiomatic` | TypeScript source under a DDD layout | `code-domain-layer`, `code-app-usecases` |
| `code-py-idiomatic` | Python source under a DDD layout | `code-domain-layer`, `code-app-usecases` |
| `code-english-only` | Identifiers and runtime strings must be English whatever language the team speaks | none |

## Format of an entry

- An entry's name is its file name, and it is also the rule's file name in a project. The frontmatter holds `installs-as` (that same name), `applies-when`, and `paths-default`: the layout the entry was written against. The last is a hint to propose from, never a decision.
- Everything above the `<!-- rationale -->` line is the rule, written without folder names. Everything below it is the reasoning, for people and for tuning the rule; it is never copied into a project.
- To instantiate: copy the rule verbatim, replace the frontmatter with a `paths:` list of folders that exist or that the user stated, and name the file `installs-as` plus `.md`. Those names follow the shape of the project's other rules: an area word, then one or two topic words, hyphenated, about 15 to 17 characters.
