# Engineering Rules

- Preserve and evolve existing shared solutions. Do not split them into
  caller-specific implementations unless strictly necessary; discuss any such
  split with the user before implementing it.
- Do not change established technical decisions or contracts without explicit
  user approval. This includes alphabets, encodings, formats, lengths, and
  interfaces. Approval to fix an implementation detail does not authorize
  changing those contracts.
- Keep existing descriptions and documentation unchanged when the public
  contract and expected behavior are unchanged. Do not rewrite them merely to
  reflect internal implementation changes, unless the user explicitly requests
  a documentation update.
