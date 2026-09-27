# Design It Twice

Load for `think twice`, or when the shape of an interface, module, or data model is the crux of a `decide`. Your first design is unlikely to be the best (Ousterhout). Vocabulary: see the `depth-and-seam` lens in `lenses.md` (module, interface, seam, adapter, depth).

## 1. Frame the problem space

Before generating designs, write a short user-facing frame:

- Constraints any design must satisfy (callers, invariants, ordering, error modes, performance).
- Dependencies it relies on and how each is reached: in-process, local substitute (e.g. in-memory DB), owned remote service, or true external.
- A rough illustrative code sketch that makes the constraints concrete. It is not a proposal.

Show it, then continue immediately; the user reads while the designs are produced.

## 2. Generate 3+ radically different designs

Give each design a different constraint. When sub-agents are available and the user allowed them, run one per design in parallel with a technical brief (files, coupling, dependency categories, what sits behind the seam) and the project's domain terms; otherwise produce them sequentially yourself.

- Design A: minimize the interface, 1–3 entry points, maximum leverage per entry point.
- Design B: maximize flexibility, many use cases and extension points.
- Design C: optimize the most common caller; the default case is trivial.
- Design D (when cross-seam dependencies exist): ports and adapters.

Each design states: the interface (types, methods, params, invariants, errors), a caller usage example, what the implementation hides, the dependency strategy and adapters, and where leverage is high or thin.

## 3. Compare and recommend

Present the designs one after another, then compare them in prose on **depth** (leverage per unit of interface), **locality** (where future change concentrates), and **seam placement** (does something actually vary across it?). Give one recommendation. If parts of two designs combine well, propose the hybrid. Be opinionated: the user wants a strong read, not a menu.
