import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

import barkx from "../../react/index.js";

export default defineConfig({
  plugins: [
    barkx(),
    react({ include: /\.(?:[jt]sx?|barkx)$/ }),
  ],
});
