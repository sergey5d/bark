# Bark

Bark is a tiny bracket-based markup language for writing HTML with less tag noise.
It is just another attempt to make something more concise than HTML while staying fully compatible with it.
I built it for my personal website, and there might still be sharp edges in it. If you run into any, feel free to reach out and I will try to fix them.

It is intentionally small:

- tags are written as `[tag ...]`
- bare `[` means `div`
- ids use `@id`
- classes use `:class-name`
- attributes use `key=value`
- `|` separates metadata from body text when needed

Bark can go both ways:

- `bark -> html`
- `html -> bark`

## Install

Install from the repo locally while you are working on it:

```bash
go install .
```

This uses your local working tree, including uncommitted changes.

Or install it from GitHub:

```bash
go install github.com/sergey5d/bark@latest
```

This uses the latest version available on GitHub, so if you have changed the repo locally, push first or use `go install .` instead.

If you installed `bark`, run it directly:

```bash
bark "*.bark"
bark import "*.html"
```

If you already have an older `bark` binary installed, run `go install .` again after pulling or changing the repo so the command on your `PATH` picks up the latest version.

## Example

```bark
[html lang=en
  [head
    [meta charset=utf-8]
    [title Example page]
    [link rel=stylesheet href=site.css]
  ]
  [body
    [header :site-header
      [:shell
        [a :wordmark href=landing-page.html aria-label=Home
          [span :dot]
          [span Example Person]
        ]
      ]
    ]
    [main
      [:page-stack
        [:page-head
          [h1 :title | Hello]
          [p :lede | This is Bark.]
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
[:shell
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

Use `:class-name` for classes:

```bark
[p :hello-title Hello]
[:shell :page-stack]
```

Transcribes to:

```html
<p class="hello-title">Hello</p>
<div class="shell page-stack"></div>
```

Repeat `:class-name` for multiple classes.

### Attributes

Attributes use `key=value`:

```bark
[a href=about.html | About]
[meta name=viewport content="width=device-width, initial-scale=1"]
[p attr=value Text]
```

Transcribes to:

```html
<a href="about.html">About</a>
<meta name="viewport" content="width=device-width, initial-scale=1">
<p attr="value">Text</p>
```

Quote a value when it contains spaces or special characters:

```bark
[p title="hello world" Hover text]
```

Transcribes to:

```html
<p title="hello world">Hover text</p>
```

### Body separator

Use `|` optionally when a block has ids, classes, or attributes and you want to start body text more clearly:

```bark
[h1 :title | Hello]
[a href=contact-page.html | Contact]
[p @hero :lede | Intro text]
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
[p @hero :lede title=hello Text without separator]
```

Transcribes to:

```html
<span>Hello</span>
<p>This is plain text.</p>
<p id="hero" class="lede" title="hello">Text without separator</p>
```

### Escaping

Escaping is intentionally minimal.

The only supported escape sequences are:

- `\[` which means a literal `[`
- `\]` which means a literal `]`

If you need to use characters that would otherwise look like metadata, end the metadata block first with `|`.

For example:

```bark
[p | a=b]
[p | :note]
[p literal \[ bracket and \] bracket]
```

Transcribes to:

```html
<p>a=b</p>
<p>:note</p>
<p>literal [ bracket and ] bracket</p>
```

Quoted attribute values are not affected by `\[` or `\]` escaping.

### Raw text tags

`script` and `style` are treated as raw-text tags.

Inside them, bare `[` and `]` are allowed as normal text as long as they balance before the final closing `]` of the Bark block:

```bark
[script const xs = [1, 2, 3];]
[style .note[data-kind="x"] { color: red; }]
```

Transcribes to:

```html
<script>const xs = [1, 2, 3];</script>
<style>.note[data-kind="x"] { color: red; }</style>
```

## Commands

If `bark` is installed, default mode generates HTML from Bark:

```bash
bark "*.bark"
bark gen "*.bark"
bark -g "*.bark"
```

Import mode converts HTML to Bark:

```bash
bark import "*.html"
bark degen "*.html"
bark -i "*.html"
bark -d "*.html"
```

If you want to run it without installing first:

```bash
go run bark.go "*.bark"
go run bark.go import "*.html"
```

## Example files

See [examples/landing-page.bark](examples/landing-page.bark) and [examples/contact-page.bark](/examples/contact-page.bark) for larger examples.
