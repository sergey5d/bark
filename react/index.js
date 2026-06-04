import { readFile } from "node:fs/promises";

import { transformWithEsbuild } from "vite";

import { transformBarkx } from "./transform.js";

function stripQuery(id) {
  const queryIndex = id.indexOf("?");
  return queryIndex >= 0 ? id.slice(0, queryIndex) : id;
}

export function barkx() {
  return {
    name: "barkx",
    enforce: "pre",
    async load(id) {
      const cleanID = stripQuery(id);
      if (!cleanID.endsWith(".barkx")) {
        return null;
      }

      const source = await readFile(cleanID, "utf8");
      const jsx = transformBarkx(source, cleanID);

      return transformWithEsbuild(jsx, cleanID, {
        loader: "jsx",
        jsx: "automatic",
        sourcemap: true,
        target: "esnext",
      });
    },
  };
}

export default barkx;
