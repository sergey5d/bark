export class BarkxSyntaxError extends SyntaxError {
  constructor(message, id, pos) {
    super(`${id}: ${message} at offset ${pos}`);
    this.name = "BarkxSyntaxError";
    this.id = id;
    this.pos = pos;
  }
}

export function transformBarkx(source, id = "<anonymous>") {
  let output = "";
  let last = 0;
  let pos = 0;

  while (pos < source.length) {
    const next = skipJavaScriptConstruct(source, pos, id);
    if (next !== pos) {
      pos = next;
      continue;
    }

    if (source[pos] === "#" && source[pos + 1] === "[") {
      output += source.slice(last, pos);
      const parser = new BarkxParser(source, pos + 1, id);
      const node = parser.parseNode();
      output += `(${renderNode(node)})`;
      pos = parser.pos;
      last = pos;
      continue;
    }

    pos++;
  }

  return output + source.slice(last);
}

class BarkxParser {
  constructor(source, pos, id) {
    this.source = source;
    this.pos = pos;
    this.id = id;
  }

  parseNode() {
    this.expect("[");

    const node = {
      tag: "div",
      id: null,
      classes: [],
      attrs: [],
      styles: [],
      children: [],
    };

    if (isTagNameStart(this.peek())) {
      node.tag = this.readTagName();
    }

    for (;;) {
      this.skipSpaces();

      switch (this.peek()) {
        case "@":
          if (node.id !== null) {
            this.error("multiple @id entries are not allowed in barkx");
          }
          node.id = this.parseID();
          continue;
        case ":":
          node.classes.push(this.parseClassName());
          continue;
        case "~":
          node.styles.push(this.parseStyleDecl());
          continue;
        default:
          if (this.looksLikeAttr()) {
            const attr = this.parseAttr();
            if (attr.name === "id" && node.id !== null) {
              this.error("id cannot be defined with both @id and id=");
            }
            if (attr.name === "class" || attr.name === "className") {
              if (node.classes.length > 0) {
                this.error("className cannot be defined with both :class and className=");
              }
            }
            if (attr.name === "style" && node.styles.length > 0) {
              this.error("style cannot be defined with both ~property and style=");
            }
            node.attrs.push(attr);
            continue;
          }
      }

      break;
    }

    this.skipSpaces();
    if (this.peek() === "|") {
      this.pos++;
      if (isSpace(this.peek())) {
        this.skipSpaces();
      }
    }

    let text = "";
    const flushText = () => {
      if (text === "") {
        return;
      }
      node.children.push({ kind: "text", value: text });
      text = "";
    };

    while (!this.eof()) {
      if (this.peek() === "\\") {
        const escaped = this.parseEscapedText();
        if (escaped !== null) {
          text += escaped;
          continue;
        }
      }

      switch (this.peek()) {
        case "[":
          flushText();
          node.children.push({ kind: "node", node: this.parseNode() });
          break;
        case "{":
          flushText();
          node.children.push({ kind: "expr", code: this.parseExpression() });
          break;
        case "]":
          this.pos++;
          flushText();
          return node;
        default:
          text += this.next();
          break;
      }
    }

    this.error(`unterminated barkx node <${node.tag}>`);
  }

  parseID() {
    this.expect("@");
    const start = this.pos;
    while (!this.eof() && isShortcutNamePart(this.peek())) {
      this.pos++;
    }
    if (start === this.pos) {
      this.error("expected an id after @");
    }
    return this.source.slice(start, this.pos);
  }

  parseClassName() {
    this.expect(":");
    const start = this.pos;
    while (!this.eof() && isShortcutNamePart(this.peek())) {
      this.pos++;
    }
    if (start === this.pos) {
      this.error("expected a class name after :");
    }
    return this.source.slice(start, this.pos);
  }

  parseStyleDecl() {
    this.expect("~");
    const start = this.pos;
    while (!this.eof() && isStyleNamePart(this.peek())) {
      this.pos++;
    }
    if (start === this.pos) {
      this.error("expected a style property after ~");
    }
    const name = this.source.slice(start, this.pos);

    this.expect("=");
    return {
      name,
      value: this.parsePropValue(),
    };
  }

  parseAttr() {
    const name = this.readAttrName();
    this.expect("=");
    return {
      name,
      value: this.parsePropValue(),
    };
  }

  parsePropValue() {
    if (this.eof()) {
      this.error("expected a value");
    }

    switch (this.peek()) {
      case '"':
      case "'":
        return {
          kind: "string",
          value: this.readQuotedString(),
        };
      case "{":
        return {
          kind: "expr",
          code: this.parseExpression(),
        };
      default:
        return {
          kind: "string",
          value: this.readBareValue(),
        };
    }
  }

  parseExpression() {
    const block = readBalancedJavaScript(this.source, this.pos, this.id);
    this.pos = block.end;
    return transformBarkx(block.content, this.id);
  }

  parseEscapedText() {
    if (this.pos + 1 >= this.source.length) {
      return null;
    }

    const next = this.source[this.pos + 1];
    if (next === "[" || next === "]" || next === "{" || next === "}" || next === "\\" || next === "#") {
      this.pos += 2;
      return next;
    }

    return null;
  }

  readTagName() {
    const start = this.pos;
    while (!this.eof() && isTagNamePart(this.peek())) {
      this.pos++;
    }
    return this.source.slice(start, this.pos);
  }

  readAttrName() {
    const start = this.pos;
    while (!this.eof() && isAttrNamePart(this.peek())) {
      this.pos++;
    }
    return this.source.slice(start, this.pos);
  }

  readQuotedString() {
    const quote = this.next();
    let value = "";

    while (!this.eof()) {
      const ch = this.next();
      if (ch === "\\") {
        if (this.eof()) {
          this.error("unterminated escape in quoted string");
        }
        value += this.next();
        continue;
      }
      if (ch === quote) {
        return value;
      }
      value += ch;
    }

    this.error("unterminated quoted string");
  }

  readBareValue() {
    const start = this.pos;
    while (!this.eof()) {
      const ch = this.peek();
      if (isSpace(ch) || ch === "[" || ch === "]" || ch === "{") {
        break;
      }
      this.pos++;
    }
    if (start === this.pos) {
      this.error("expected a bare value");
    }
    return this.source.slice(start, this.pos);
  }

  looksLikeAttr() {
    const start = this.pos;
    if (!isAttrNameStart(this.source[start])) {
      return false;
    }

    let pos = start;
    while (pos < this.source.length && isAttrNamePart(this.source[pos])) {
      pos++;
    }
    return this.source[pos] === "=";
  }

  skipSpaces() {
    while (!this.eof() && isSpace(this.peek())) {
      this.pos++;
    }
  }

  expect(ch) {
    if (this.peek() !== ch) {
      this.error(`expected ${JSON.stringify(ch)}`);
    }
    this.pos++;
  }

  error(message) {
    throw new BarkxSyntaxError(message, this.id, this.pos);
  }

  eof() {
    return this.pos >= this.source.length;
  }

  peek() {
    return this.eof() ? "" : this.source[this.pos];
  }

  next() {
    const ch = this.peek();
    this.pos++;
    return ch;
  }
}

function renderNode(node) {
  const props = [];

  if (node.id !== null) {
    props.push(`id=${renderPropValue({ kind: "string", value: node.id })}`);
  }
  if (node.classes.length > 0) {
    props.push(`className=${renderPropValue({ kind: "string", value: node.classes.join(" ") })}`);
  }
  if (node.styles.length > 0) {
    props.push(renderStyleProp(node.styles));
  }

  for (const attr of node.attrs) {
    props.push(`${attr.name}=${renderPropValue(attr.value)}`);
  }

  const children = renderChildren(node.children);
  const propSuffix = props.length > 0 ? ` ${props.join(" ")}` : "";

  if (children.length === 0) {
    return `<${node.tag}${propSuffix} />`;
  }

  return `<${node.tag}${propSuffix}>${children.join("")}</${node.tag}>`;
}

function renderChildren(children) {
  return children
    .filter((child) => child.kind !== "text" || child.value.trim() !== "")
    .map((child) => {
      switch (child.kind) {
        case "node":
          return renderNode(child.node);
        case "expr":
          return `{${child.code}}`;
        case "text":
          return `{${JSON.stringify(child.value)}}`;
        default:
          return "";
      }
    });
}

function renderStyleProp(styles) {
  const entries = styles.map((style) => {
    const key = renderStyleKey(style.name);
    const value = style.value.kind === "expr" ? style.value.code : JSON.stringify(style.value.value);
    return `${key}: ${value}`;
  });

  return `style={{ ${entries.join(", ")} }}`;
}

function renderStyleKey(name) {
  if (name.startsWith("--")) {
    return JSON.stringify(name);
  }

  const camel = name.replace(/-([a-zA-Z0-9])/g, (_, ch) => ch.toUpperCase());
  return /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(camel) ? camel : JSON.stringify(camel);
}

function renderPropValue(value) {
  if (value.kind === "expr") {
    return `{${value.code}}`;
  }
  return `{${JSON.stringify(value.value)}}`;
}

function readBalancedJavaScript(source, start, id) {
  if (source[start] !== "{") {
    throw new BarkxSyntaxError("expected '{' to start a JavaScript expression", id, start);
  }

  let depth = 0;
  let pos = start;

  while (pos < source.length) {
    const next = skipJavaScriptConstruct(source, pos, id);
    if (next !== pos) {
      pos = next;
      continue;
    }

    const ch = source[pos];
    if (ch === "{") {
      depth++;
      pos++;
      continue;
    }
    if (ch === "}") {
      depth--;
      pos++;
      if (depth === 0) {
        return {
          content: source.slice(start + 1, pos - 1),
          end: pos,
        };
      }
      continue;
    }

    pos++;
  }

  throw new BarkxSyntaxError("unterminated JavaScript expression", id, start);
}

function skipJavaScriptConstruct(source, pos, id) {
  const ch = source[pos];
  const next = source[pos + 1];

  if (ch === "'" || ch === '"') {
    return skipQuotedString(source, pos, ch, id);
  }
  if (ch === "`") {
    return skipTemplateLiteral(source, pos, id);
  }
  if (ch === "/" && next === "/") {
    return skipLineComment(source, pos);
  }
  if (ch === "/" && next === "*") {
    return skipBlockComment(source, pos, id);
  }

  return pos;
}

function skipQuotedString(source, pos, quote, id) {
  let cursor = pos + 1;

  while (cursor < source.length) {
    const ch = source[cursor];
    if (ch === "\\") {
      cursor += 2;
      continue;
    }
    if (ch === quote) {
      return cursor + 1;
    }
    cursor++;
  }

  throw new BarkxSyntaxError("unterminated JavaScript string literal", id, pos);
}

function skipTemplateLiteral(source, pos, id) {
  let cursor = pos + 1;

  while (cursor < source.length) {
    const ch = source[cursor];
    if (ch === "\\") {
      cursor += 2;
      continue;
    }
    if (ch === "`") {
      return cursor + 1;
    }
    if (ch === "$" && source[cursor + 1] === "{") {
      const block = readBalancedJavaScript(source, cursor + 1, id);
      cursor = block.end;
      continue;
    }
    cursor++;
  }

  throw new BarkxSyntaxError("unterminated JavaScript template literal", id, pos);
}

function skipLineComment(source, pos) {
  let cursor = pos + 2;
  while (cursor < source.length && source[cursor] !== "\n") {
    cursor++;
  }
  return cursor;
}

function skipBlockComment(source, pos, id) {
  const end = source.indexOf("*/", pos + 2);
  if (end < 0) {
    throw new BarkxSyntaxError("unterminated block comment", id, pos);
  }
  return end + 2;
}

function isSpace(ch) {
  return ch === " " || ch === "\n" || ch === "\r" || ch === "\t" || ch === "\f";
}

function isTagNameStart(ch) {
  return /[A-Za-z_]/.test(ch);
}

function isTagNamePart(ch) {
  return /[A-Za-z0-9_.-]/.test(ch);
}

function isShortcutNamePart(ch) {
  return /[A-Za-z0-9_-]/.test(ch);
}

function isStyleNamePart(ch) {
  return /[A-Za-z0-9_-]/.test(ch);
}

function isAttrNameStart(ch) {
  return /[A-Za-z_$]/.test(ch);
}

function isAttrNamePart(ch) {
  return /[A-Za-z0-9_$:-]/.test(ch);
}
