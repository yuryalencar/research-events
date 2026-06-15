import nextConfig from "eslint-config-next"

// NOTE: eslint is pinned to ^9.x (see package.json) instead of the latest
// v10 because eslint-plugin-react@7.37.5 (pulled in by eslint-config-next)
// still calls the removed `context.getFilename()` API and crashes under
// ESLint 10. eslint-config-next@16.2.7's own peerDependencies require
// `eslint >=9.0.0`, so v9 is the actually-supported version today.
// Once eslint-plugin-react ships a fix and eslint-config-next raises its
// peer range, remove this note and bump eslint back to ^10.
const config = [
  ...nextConfig,
  {
    ignores: [".next/**", "node_modules/**"],
  },
]

export default config
