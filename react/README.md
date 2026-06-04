# Bark React

`barkx` is an experimental React mode for Bark.

It supports two authoring modes.

Module mode keeps normal JavaScript module semantics and uses `#[ ... ]` to enter Bark markup:

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

Template-only mode lets the file itself be the component:

```bark
[button
  :counter-button
  :bark-button
  :{props.className}
  onClick={props.onReset}
  type=button
  [span :button-label Reset in Bark]
  [span :button-detail Current count: {props.count}]
]
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

- A `.barkx` file can be either a normal JS/TS module with `#[ ... ]` snippets, or a template-only component with a top-level Bark root.
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
- Template-only mode currently expects a single top-level Bark root.
- `.barkx` remains separate from the existing HTML-only Bark CLI.
