import test from "node:test";
import assert from "node:assert/strict";

import { transformBarkx } from "./transform.js";

test("transforms Barkx snippets into JSX with Bark sugar", () => {
  const source = `import Button from "./Button";

export default function App({ title, onSave }) {
  return #[main @page :shell
    [h1 {title}]
    [Button :primary onClick={onSave} Save]
  ];
}
`;

  const output = transformBarkx(source, "App.barkx");

  assert.match(output, /return\s+\(<main id=\{"page"\} className=\{"shell"\}>/);
  assert.match(output, /<h1>\{title\}<\/h1>/);
  assert.match(output, /<Button className=\{"primary"\} onClick=\{onSave\}>\{"Save"\}<\/Button>/);
});

test("preserves bare [ syntax as div inside Barkx", () => {
  const source = `export default function Panel() {
  return #[ @root :shell
    [span Hello]
  ];
}
`;

  const output = transformBarkx(source, "Panel.barkx");

  assert.match(output, /<div id=\{"root"\} className=\{"shell"\}>/);
  assert.match(output, /<span>\{"Hello"\}<\/span>/);
});

test("transforms nested Barkx snippets inside JavaScript expressions", () => {
  const source = `export default function List({ items }) {
  return #[ul
    {items.map(item => #[li key={item.id} {item.label}])}
  ];
}
`;

  const output = transformBarkx(source, "List.barkx");

  assert.match(output, /\{items\.map\(item => \(\<li key=\{item\.id\}>\{item\.label\}<\/li>\)\)\}/);
});

test("rewrites inline Bark style declarations into a React style prop", () => {
  const source = `export default function Button() {
  return #[button ~background-color=red ~padding={spacing} Click];
}
`;

  const output = transformBarkx(source, "Button.barkx");

  assert.match(output, /style=\{\{ backgroundColor: "red", padding: spacing \}\}/);
});

test("supports dynamic class sugar with :{expr}", () => {
  const source = `export default function Button({ classes }) {
  return #[button :primary :{classes} Click];
}
`;

  const output = transformBarkx(source, "DynamicClass.barkx");

  assert.match(output, /className=\{\["primary", classes\]\.filter\(Boolean\)\.join\(" "\)\}/);
});

test("supports dynamic id sugar with @{expr}", () => {
  const source = `export default function Section({ sectionId }) {
  return #[section @{sectionId} Hello];
}
`;

  const output = transformBarkx(source, "DynamicID.barkx");

  assert.match(output, /<section id=\{sectionId\}>/);
});
