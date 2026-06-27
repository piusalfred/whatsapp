# Code Documentation Skill for AI Agents

**Core Directive**  
Your role is to provide documentation that adds essential context without cluttering the codebase. Balance the ideal of self-documenting code with the reality that developers need structural, behavioral, and conceptual information raw syntax cannot convey. This skill translates that human‑centric trade‑off into strict rules to prevent redundant or high‑maintenance comments.

---

## 1. When NOT to Comment (The Restraint Rule)

Exercise strict restraint—skip documentation entirely in these cases:

- **Simple, idiomatic code** – Do not translate readable syntax into English. If a competent developer can understand it at a glance, adding a comment is noise. (Never write `// loops through the array` or `// sets the user ID`.)
- **ASCII art and section separators** – Never use decorative banners like `// ======`, `// ------`, `// ******`, or similar to visually group code. The file structure, function names, and blank lines provide grouping. These rot immediately when code moves and add zero semantic value.
- **Non-standard, non-godoc comments** – Only two kinds of comments are acceptable: (a) godoc comments on exported declarations (`// FuncName does X.`), and (b) sparse inline comments inside function bodies explaining *why*, not *what*. Everything else — label comments (`// --- Groups ---`), ownership markers, change logs in code, closing-brace annotations (`} // end if`), and decorative filler — is noise. Delete it.
- **Highly volatile or prototyping code** – When the codebase is in constant flux, comments become stale almost instantly, eroding trust. Keep them to an absolute minimum.
- **Granular maintenance burden** – If a comment would require updating with every minor logic or variable tweak, omit it. Instead, elevate the explanation to a higher‑level concept elsewhere.

---

## 2. Where to Place Documentation (The Location Strategy)

Separate documentation based on whether the reader is a **consumer** or a **maintainer**.

- **API boundaries & declarations (for consumers)** – Place formal, comprehensive documentation at the class, interface, module, type definition, or docstring level. Consumers find safe‑usage information here without reading implementation details.
- **Implementation bodies (for maintainers)** – Use sparse, inline comments inside function/method bodies solely to explain *why* the internal logic does something non‑obvious.
- **External project documentation** – Reserve READMEs, wikis, and dedicated architecture files for cross‑module guides, environment setup, and macro‑level design patterns.

---

## 3. What Must Be Documented (The Content Taxonomy)

Source code shows *how*; documentation must provide the missing *what*, *why*, and *where*. Address these elements explicitly:

- **Inputs & environmental state** – Define what parameters, flags, or configuration objects mean. Document the “state of the world” required before calling: how inputs behave when empty, null, undefined, or omitted.
- **Task orientation (“How do I…?”)** – Provide a clear path to usage. Explain how a developer uses this element to accomplish a concrete task. (e.g., “Format raw database timestamps into localized user‑facing strings”.)
- **Return values & failure modes** – State exactly what a return value represents. Document all failure modes—error codes, thrown exceptions, null returns—and what triggers them.
- **Hidden contracts & invisible obligations** – Detail requirements not enforced by the compiler: “Must call `init()` first”, “Must be awaited”, “Not thread‑safe”, “Requires explicit cleanup to avoid memory leaks”.
- **Side effects** – List actions beyond the immediate return value: writing logs, emitting events, updating a cache, mutating an external database.
- **Relational context** – Clarify how this file, class, or function fits into the broader project architecture. Highlight its role among related components.
- **Complex signatures & auto‑generated code** – Thoroughly document heavily overloaded parameters, abstract method signatures, or auto‑generated boundaries that humans cannot parse easily.
- **Extension points & subclassing contracts** – For object‑oriented classes, interfaces, or traits, explicitly state whether a developer is expected, allowed, or forbidden to override specific methods.
- **Design decisions & the “Why”** – At the statement or block level, explain non‑obvious logic. Reveal historical baggage, business rules, performance trade‑offs, or external constraints (e.g., `// Workaround for API rate limit; retry after 50ms`). If code intentionally diverges from a specification, explain why.

---

## 4. Quality Standards & Audience Awareness

Write clean, structured documentation that respects both human readers and automated tooling.

- **Use standard doc syntax** – Format API‑boundary documentation with JSDoc, Pydoc, Javadoc, or equivalent standard tags so generators can parse it reliably.
- **Focus on what parsers miss** – Tools extract names, parameters, and types. Your documentation must supply intent, context, nuance, and algorithmic complexity.
- **Eliminate ambiguity** – Use precise, standard terminology. Avoid vague descriptions or insider jargon that could lead to systemic misunderstandings.
- **Assume code is less intuitive than it seems** – Authors (including AI) routinely overestimate clarity for outsiders. Never assume the reader shares your immediate execution context.
- **Bridge skill gaps** – Advanced developers may verify logic by reading code, but juniors depend on documentation for orientation. Favor high‑level conceptual overviews over line‑by‑line handholding.

---

## 5. Workflow Integration (The Timing Rule)

Documentation is an active part of development, not an afterthought.

- **Document during authoring** – Generate documentation alongside code creation or review. This is when operational context is freshest.
- **Refactor comments with code** – When modifying code, audit all nearby comments. If a refactor, rename, or structural change makes a comment redundant or inaccurate, update or delete it immediately.