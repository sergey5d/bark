import { transformBarkx } from "./transform.js";

function stripQuery(id) {
  const queryIndex = id.indexOf("?");
  return queryIndex >= 0 ? id.slice(0, queryIndex) : id;
}

export function barkx() {
  return {
    name: "barkx",
    enforce: "pre",
    transform(source, id) {
      if (!stripQuery(id).endsWith(".barkx")) {
        return null;
      }

      return {
        code: transformBarkx(source, id),
        map: null,
      };
    },
  };
}

export default barkx;
