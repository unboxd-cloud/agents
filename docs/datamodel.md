# Data model — exposed to other repos and users

The platform's data model (the ADL metamodels, blueprints, and `platform.agent`)
is exposed four ways, so any repository, tool, or user can consume it:

1. **Source of truth — `.agent` files.** `platform.agent`, `metamodels/*.agent`,
   and `blueprints/*.agent` are the authored, human-readable model. Validate them
   with `make agents`.

2. **Go API — `pkg/adl`.** Import `github.com/unboxd-cloud/platform/pkg/adl` to
   parse, validate, and traverse the model from Go (`adl.Load`, `adl.Compile`,
   the `Agent` view).

3. **WebAssembly — `web/adl-runtime`.** The same runtime compiled to WASM, with
   TypeScript types, for the editor tooling and any JS/TS consumer.

4. **Exported JSON — `docs/datamodel.json`.** The whole model compiled into one
   machine-readable document (every entity, relation, brain, policy, function,
   etc., per file, with validity), regenerated with:

   ```sh
   make model     # -> docs/datamodel.json
   ```

   Other repos can fetch the raw JSON directly, or run
   `platform agent export <files…>` to produce their own slice.
