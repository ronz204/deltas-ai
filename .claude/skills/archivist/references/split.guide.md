# Fitting an artifact to its context budget

Read this when the budget script reports an artifact over its ceiling. The goal is to lower what is loaded without losing any information: content moves to a place that loads less often, it is never discarded.

## The principle

Split by how often content is used, not by how big it is.

- Stays hot: anything needed on every load of the artifact (the decisions, the steps, the rules that apply every time).
- Moves out: anything needed only for some uses (a type-specific procedure, a template, a long example, an edge case, a table used occasionally).
- Moving something every use needs saves nothing and adds a read step, so do not.

## The procedure

1. Run the budget script and note the artifact's class, its size, and its ceiling.
2. Read the artifact whole and mark each section as used every time or only sometimes. When unsure, keep it hot.
3. Draft a split plan: what stays, what moves, the new file's name and location, and the exact pointer that replaces each moved part.
4. Show the plan to the user and wait for a go-ahead. Nothing is rewritten before that.
5. Apply it: move the text verbatim, then add the pointers.
6. Run the script again. The artifact must be at or under its ceiling, no pointer may be broken, and the always-loaded total must not have grown.
7. Compare what moved against the original once: every moved section is present in exactly one new place.

## Pointers must be steps

An agent that is not told to read a moved file will often not read it. Write each pointer as an explicit step that names the trigger, for example "when writing a subagent, read the agents guide before starting". A vague "see also" is not a pointer.

## Per artifact type

| Over its ceiling | How to fit it | Watch for |
|---|---|---|
| Skill body | Keep the numbered steps and decisions. Move type-specific procedures, templates, long examples, and rare cases into reference files, one per topic, each named after the kind of thing it covers in the same pattern as the existing reference files. | A skill's instructions must still read as a complete procedure without the references. |
| Rule with `paths:` | Split along the areas of the codebase it governs, one rule per concern, each with its own narrower `paths:`. Name the new rules in the same pattern as the existing rule names. | A rule that cannot be split by area is usually two conventions; split by convention instead. |
| Rule without `paths:` | First ask whether it needs to be global. Scoping it is the cheapest fix. Otherwise move the reasoning and examples to a scoped rule and keep only the directive. | Global rules cost the most; do not grow them. |
| `CLAUDE.md` | Reduce it to a router: orientation, layout, commands, pointers. Move durable content into the matching doc and leave a pointer behind. | It is the one file allowed to point at everything. |
| A doc | Evaluate which sections stand alone. Keep the doc as the entry point holding the overview and an index, and move the detailed sections into files inside a folder named after the doc. Each detail file stands alone and does not point to its siblings; only the entry doc points to its own details. | Keep file and folder names short and in the same length band as the existing doc names, one topic word each. The doc set stays the same size at the top level. |
| Subagent body | Replace restated boilerplate with pointers to the rules and docs that already hold it. | A subagent runs in its own context, so anything it needs must be reachable from there. |
| Description of a skill or subagent | Do not split. See the context budget rule. | Shortening can stop it triggering. |
| Spec or design | Do not split mechanically. Hand the question to the specifier skill: does this slice span more than one capability? An optional context section that is large can move out of the spec. | Never split a slice by layer. |

## Naming

Follow the naming pattern of the siblings already in the folder. New sibling names stay close in length to the existing ones, so a listing reads evenly, and each name states its topic in a word or two.
