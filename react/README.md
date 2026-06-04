# Bark React

`barkx` is an experimental React mode for Bark.

It keeps normal JavaScript module semantics and uses `#[ ... ]` to enter Bark markup:

```jsx
import MyButton from "./MyButton";

export default function Page({ onSave, title }) {
  return #[main @page :shell
    [h1 {title}]
    [MyButton :primary onClick={onSave} Save]
    [ :card
      [p Bark keeps `:` classes, `@` ids, and bare `[` as `div`.]
    ]
  ];
}
```

The Vite plugin rewrites `.barkx` files into JSX and immediately lowers that JSX to normal JavaScript before Vite import analysis runs.

## Usage

```js
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import barkx from "./react/index.js";

export default defineConfig({
  plugins: [
    barkx(),
    react({ include: /\.(?:[jt]sx?|barkx)$/ }),
  ],
});
```

## Current behavior

- `#[ ... ]` enters Bark mode from JS/TS.
- Nested Bark tags continue to use `[tag ...]`.
- Bare `[` defaults to `div`.
- `:class-name` compiles to `className`.
- `:{expr}` appends dynamic class names into `className`.
- `@id` compiles to `id`.
- `@{expr}` compiles to a dynamic `id`.
- `key=value` and `key={expr}` compile to JSX props.
- `{expr}` inside Bark compiles to JSX child expressions.

## Current limits

- The transformer currently skips strings, template literals, and comments when looking for `#[ ... ]`, but it does not try to fully parse JavaScript regex literals.
- `.barkx` is intentionally focused on React-style modules, not the existing HTML-only Bark CLI.
