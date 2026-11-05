# Contributor agent instructions

- Read README.md, CONTRIBUTING.md and docs/publisher/STATUS.md first.
- Keep this application local and lightweight: Go, Wails, Vue and strict TypeScript.
- Keep Go business logic independent of Wails. Use narrow, typed desktop adapters.
- No generic shell, arbitrary URL or unrestricted filesystem bridge methods.
- Treat manifests and bridge inputs as untrusted. Preserve path and integrity checks.
- Previewing/importing never creates a review or approval. Never publish the demo.
- No paid services or publishing to external accounts without explicit authorization.
- Keep runtime packages pinned. Add a dependency only for a demonstrated need.
- Use the design tokens and shared controls; open sections, no decorative cards or
  colored vertical navigation markers. Preserve the supplied paw artwork.
- Keep credentials, local data, media, tools and environments out of Git.
- Run appropriate checks; describe native UI checks separately from browser fixtures.
- Record completed work and one next action in docs/publisher/STATUS.md.
- Public documentation, code and comments are English. Use neutral Spanish with
  the project owner; do not use regional slang or voseo.
