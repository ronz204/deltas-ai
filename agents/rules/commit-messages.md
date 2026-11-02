# Commit Messages

Write every commit as `type(scope): subject`, in English.

- The type is one of `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.
- The scope is the part of the project the commit touches, in parentheses. Leave it out only when nothing narrower fits.
- The subject is 40 to 50 characters long, scope and type included, in the imperative ("add", not "added"), starts lowercase after the colon and has no trailing period.
- A body is optional. Separate it from the subject with a blank line, and use it to say why, not what: the diff already shows what.
- One commit holds one change. Do not mix a refactor with a behavior change.
