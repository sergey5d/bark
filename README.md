# Bark

Bark is a tiny bracket-based markup language for writing HTML with less tag noise.
It is just another fun attempt to make something more concise than HTML while staying fully compatible with it.
It is still very rough. I built it for my personal website, and there are probably still bugs or sharp edges. If you run into any, feel free to reach out and I will try to fix them.

It is intentionally small:

- tags are written as `[tag ...]`
- bare `[` means `div`
- ids use `@id`
- classes use `<: class1, class2`
- attributes use `key=value`
- `|` separates metadata from body text when needed

Bark can go both ways:

- `bark -> html`
- `html -> bark`

## Example

```bark
[html lang=en
  [head
    [meta charset=utf-8]
    [title Example page]
    [link rel=stylesheet href=site.css]
  ]
  [body
    [header <: site-header
      [<: shell
        [a <: wordmark href=landing-page.html aria-label=Home
          [span <: dot]
          [span Example Person]
        ]
      ]
    ]
    [main
      [<: page-stack
        [<: page-head
          [h1 <: title | Hello]
          [p <: lede | This is Bark.]
        ]
      ]
    ]
  ]
]
```

Transcribes to:

```html
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Example page</title>
    <link rel="stylesheet" href="site.css">
  </head>
  <body>
    <header class="site-header">
      <div class="shell">
        <a class="wordmark" href="landing-page.html" aria-label="Home">
          <span class="dot"></span>
          <span>Example Person</span>
        </a>
      </div>
    </header>
    <main>
      <div class="page-stack">
        <div class="page-head">
          <h1 class="title">Hello</h1>
          <p class="lede">This is Bark.</p>
        </div>
      </div>
    </main>
  </body>
</html>
```

## Core syntax

### Tags

```bark
[p Hello]
[section
  [p Nested]
]
```

Transcribes to:

```html
<p>Hello</p>
<section>
  <p>Nested</p>
</section>
```

Unnamed blocks default to `div`:

```bark
[<: shell
  [p Inside a div.shell]
]
```

Transcribes to:

```html
<div class="shell">
  <p>Inside a div.shell</p>
</div>
```

### Ids

Use `@` for ids:

```bark
[@root]
[section @hero]
```

Transcribes to:

```html
<div id="root"></div>
<section id="hero"></section>
```

Multiple ids are allowed and are joined into the final `id` attribute:

```bark
[@root @primary]
```

Transcribes to:

```html
<div id="root primary"></div>
```

### Classes

Use `<:` for classes:

```bark
[p <: hello-title Hello]
[<: shell, page-stack]
```

Transcribes to:

```html
<p class="hello-title">Hello</p>
<div class="shell page-stack"></div>
```

Classes are comma-separated.

### Attributes

Attributes use `key=value`:

```bark
[a href=about.html | About]
[meta name=viewport content="width=device-width, initial-scale=1"]
```

Transcribes to:

```html
<a href="about.html">About</a>
<meta name="viewport" content="width=device-width, initial-scale=1">
```

Quote a value when it contains spaces or special characters:

```bark
[meta content="width=device-width, initial-scale=1"]
```

Transcribes to:

```html
<meta content="width=device-width, initial-scale=1">
```

### Body separator

Use `|` optionally when a block has ids, classes, or attributes and you want to start body text more clearly:

```bark
[h1 <: title | Hello]
[a href=contact-page.html | Contact]
[p @hero <: lede | Intro text]
```

Transcribes to:

```html
<h1 class="title">Hello</h1>
<a href="contact-page.html">Contact</a>
<p id="hero" class="lede">Intro text</p>
```

Here is an example where `|` is not used:

```bark
[span Hello]
[p This is plain text.]
```

Transcribes to:

```html
<span>Hello</span>
<p>This is plain text.</p>
```

### Escaping

Escapes are mainly useful while the parser is still in the metadata zone. Any escape sequence there ends metadata parsing and starts body text.

Supported escape sequences are:

- `\[` always means a literal `[`
- `\<`
- `\|`
- `\=`

For example:

```bark
[p a\=b]
[p href=x\?y=z]
```

Transcribes to:

```html
<p>a=b</p>
<p>href=x?y=z</p>
```

Once body parsing has started, only `\[` is treated specially:

```bark
[p literal \[ bracket]
```

Transcribes to:

```html
<p>literal [ bracket</p>
```

Quoted attribute values are not affected by these escapes.

## Commands

Default mode generates HTML from Bark:

```bash
go run bark.go "*.bark"
go run bark.go gen "*.bark"
go run bark.go -g "*.bark"
```

Import mode converts HTML to Bark:

```bash
go run bark.go import "*.html"
go run bark.go degen "*.html"
go run bark.go -i "*.html"
go run bark.go -d "*.html"
```

## Example files

See [examples/landing-page.bark](/Users/sergeyd/Projects/bark/examples/landing-page.bark) and [examples/contact-page.bark](/Users/sergeyd/Projects/bark/examples/contact-page.bark) for larger examples.
