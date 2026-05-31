package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type barkParser struct {
	src []rune
	pos int
}

type barkNode struct {
	Tag        string
	Classes    []string
	IDs        []string
	Attrs      map[string]string
	StyleDecls [][2]string
	Content    []barkContent
}

type barkContent struct {
	Text string
	Node *barkNode
}

func ParseBark(input string) (string, error) {
	p := &barkParser{src: []rune(input)}
	var nodes []*barkNode
	var textParts []string

	for !p.eof() {
		if p.peek() == '[' {
			n, err := p.parseNode()
			if err != nil {
				return "", err
			}
			nodes = append(nodes, n)
			continue
		}

		text := p.readUntil('[')
		if strings.TrimSpace(text) != "" {
			textParts = append(textParts, text)
		}
	}

	var out strings.Builder
	hasExplicitHTMLRoot := len(nodes) > 0 && strings.EqualFold(nodes[0].Tag, "html")
	if !hasExplicitHTMLRoot {
		out.WriteString("<html>")
	}

	for _, text := range textParts {
		out.WriteString(barkEscapeHTML(text))
	}
	for _, n := range nodes {
		out.WriteString(n.HTML())
	}

	if !hasExplicitHTMLRoot {
		out.WriteString("</html>")
	}

	return out.String(), nil
}

func (n *barkNode) HTML() string {
	return n.renderHTML(0, false)
}

func (n *barkNode) renderHTML(indent int, ownLine bool) string {
	var b strings.Builder
	indentStr := strings.Repeat("  ", indent)
	isRawText := barkIsRawTextTag(n.Tag)
	hasElementChildren := false
	hasMeaningfulText := false
	for _, item := range n.Content {
		if item.Node != nil {
			hasElementChildren = true
			continue
		}
		if item.Text == "" {
			continue
		}
		if isRawText || strings.TrimSpace(item.Text) != "" {
			hasMeaningfulText = true
		}
	}
	blockLayout := hasElementChildren

	if ownLine || blockLayout {
		b.WriteString(indentStr)
	}
	b.WriteByte('<')
	b.WriteString(n.Tag)

	if len(n.IDs) > 0 {
		b.WriteString(` id="`)
		b.WriteString(barkEscapeHTML(strings.Join(n.IDs, " ")))
		b.WriteByte('"')
	}

	if len(n.Classes) > 0 {
		b.WriteString(` class="`)
		b.WriteString(barkEscapeHTML(strings.Join(n.Classes, " ")))
		b.WriteByte('"')
	}

	if len(n.StyleDecls) > 0 {
		parts := make([]string, 0, len(n.StyleDecls))
		for _, decl := range n.StyleDecls {
			parts = append(parts, decl[0]+": "+decl[1]+";")
		}
		b.WriteString(` style="`)
		b.WriteString(barkEscapeHTML(strings.Join(parts, " ")))
		b.WriteByte('"')
	}

	if len(n.Attrs) > 0 {
		keys := make([]string, 0, len(n.Attrs))
		for k := range n.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteByte(' ')
			b.WriteString(k)
			b.WriteString(`="`)
			b.WriteString(barkEscapeHTML(n.Attrs[k]))
			b.WriteByte('"')
		}
	}

	b.WriteByte('>')

	if barkVoidTags[strings.ToLower(n.Tag)] {
		return b.String()
	}

	if !blockLayout {
		for _, item := range n.Content {
			if item.Node != nil {
				b.WriteString(item.Node.renderHTML(indent+1, false))
			} else if isRawText {
				b.WriteString(item.Text)
			} else {
				b.WriteString(barkEscapeHTML(item.Text))
			}
		}
		b.WriteString("</")
		b.WriteString(n.Tag)
		b.WriteByte('>')
		return b.String()
	}

	if hasMeaningfulText {
		for _, item := range n.Content {
			if item.Node != nil {
				continue
			}
			text := item.Text
			if !isRawText {
				text = strings.TrimSpace(text)
				if text == "" {
					continue
				}
			}
			b.WriteByte('\n')
			b.WriteString(strings.Repeat("  ", indent+1))
			if isRawText {
				b.WriteString(text)
			} else {
				b.WriteString(barkEscapeHTML(text))
			}
		}
	}

	for _, item := range n.Content {
		if item.Node == nil {
			continue
		}
		b.WriteByte('\n')
		b.WriteString(item.Node.renderHTML(indent+1, true))
	}

	b.WriteByte('\n')
	b.WriteString(indentStr)
	b.WriteString("</")
	b.WriteString(n.Tag)
	b.WriteByte('>')
	return b.String()
}

func (p *barkParser) parseNode() (*barkNode, error) {
	if err := p.expect('['); err != nil {
		return nil, err
	}

	n := &barkNode{
		Tag:   "div",
		Attrs: map[string]string{},
	}

	if r := p.peek(); barkIsNameStart(r) {
		n.Tag = p.readName()
	}

	for {
		p.skipSpaces()
		switch p.peek() {
		case '@':
			id, err := p.parseIDShortcut()
			if err != nil {
				return nil, err
			}
			n.IDs = append(n.IDs, id)
		case ':':
			item, err := p.parseClassName()
			if err != nil {
				return nil, err
			}
			n.Classes = append(n.Classes, item)
		case '~':
			key, value, err := p.parseStyleDecl()
			if err != nil {
				return nil, err
			}
			if _, exists := n.Attrs["style"]; exists {
				return nil, fmt.Errorf("style cannot be defined both with ~property and style= at rune %d", p.pos)
			}
			for _, decl := range n.StyleDecls {
				if decl[0] == key {
					return nil, fmt.Errorf("style property %q is defined more than once at rune %d", key, p.pos)
				}
			}
			n.StyleDecls = append(n.StyleDecls, [2]string{key, value})
		case '<':
			return nil, fmt.Errorf("old class syntax `<:` is no longer supported at rune %d; use :class instead", p.pos)
		default:
			if p.peek() == '{' {
				return nil, fmt.Errorf("curly-brace attribute blocks are no longer supported at rune %d", p.pos)
			}
			if p.looksLikeBareAttr() {
				key, value, err := p.parseBareAttr()
				if err != nil {
					return nil, err
				}
				if _, exists := n.Attrs[key]; exists {
					return nil, fmt.Errorf("attribute %q is defined more than once at rune %d", key, p.pos)
				}
				if key == "class" && len(n.Classes) > 0 {
					return nil, fmt.Errorf("class cannot be defined both with :class and class= at rune %d", p.pos)
				}
				if key == "id" && len(n.IDs) > 0 {
					return nil, fmt.Errorf("id cannot be defined both with @ and id= at rune %d", p.pos)
				}
				if key == "style" && len(n.StyleDecls) > 0 {
					return nil, fmt.Errorf("style cannot be defined both with style= and ~property at rune %d", p.pos)
				}
				n.Attrs[key] = value
				continue
			}
			goto body
		}
	}

body:
	p.skipSpaces()
	if p.peek() == '|' {
		p.pos++
		if !p.eof() && unicode.IsSpace(p.peek()) {
			p.skipSpaces()
		}
	}

	if barkIsRawTextTag(n.Tag) {
		raw, err := p.parseRawTextBody(n.Tag)
		if err != nil {
			return nil, err
		}
		if raw != "" {
			n.Content = append(n.Content, barkContent{Text: raw})
		}
		return n, nil
	}

	var text bytes.Buffer
	flushText := func() {
		if text.Len() == 0 {
			return
		}
		n.Content = append(n.Content, barkContent{Text: text.String()})
		text.Reset()
	}

	for {
		if p.eof() {
			return nil, fmt.Errorf("unterminated element <%s>", n.Tag)
		}

		if escaped, ok := p.escapedBodyRune(); ok {
			p.pos += 2
			text.WriteRune(escaped)
			continue
		}

		switch p.peek() {
		case '[':
			flushText()
			child, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, barkContent{Node: child})
		case ']':
			p.pos++
			flushText()
			return n, nil
		default:
			text.WriteRune(p.next())
		}
	}
}

func (p *barkParser) parseClassName() (string, error) {
	if err := p.expect(':'); err != nil {
		return "", err
	}
	start := p.pos
	for !p.eof() && barkIsClassNamePart(p.peek()) {
		p.pos++
	}
	if start == p.pos {
		return "", fmt.Errorf("expected class name after : at rune %d", p.pos)
	}
	return string(p.src[start:p.pos]), nil
}

func (p *barkParser) looksLikeBareAttr() bool {
	if p.eof() {
		return false
	}
	i := p.pos
	if !barkIsAttrNamePart(p.src[i]) {
		return false
	}
	for i < len(p.src) && barkIsAttrNamePart(p.src[i]) {
		i++
	}
	return i < len(p.src) && p.src[i] == '='
}

func (p *barkParser) parseBareAttr() (string, string, error) {
	start := p.pos
	for !p.eof() && barkIsAttrNamePart(p.peek()) {
		p.pos++
	}
	if start == p.pos {
		return "", "", fmt.Errorf("expected attribute name at rune %d", p.pos)
	}
	key := string(p.src[start:p.pos])

	if err := p.expect('='); err != nil {
		return "", "", err
	}
	if p.eof() {
		return "", "", fmt.Errorf("expected attribute value for %q at rune %d", key, p.pos)
	}

	var value string
	switch p.peek() {
	case '"', '\'':
		quote := p.next()
		start = p.pos
		for !p.eof() && p.peek() != quote {
			p.pos++
		}
		if p.eof() {
			return "", "", fmt.Errorf("unterminated quoted value for %q", key)
		}
		value = string(p.src[start:p.pos])
		p.pos++
	default:
		start = p.pos
		for !p.eof() {
			r := p.peek()
			if unicode.IsSpace(r) || r == '|' || r == '[' || r == ']' {
				break
			}
			p.pos++
		}
		value = string(p.src[start:p.pos])
	}

	return key, value, nil
}

func (p *barkParser) parseStyleDecl() (string, string, error) {
	if err := p.expect('~'); err != nil {
		return "", "", err
	}
	start := p.pos
	for !p.eof() && barkIsStyleNamePart(p.peek()) {
		p.pos++
	}
	if start == p.pos {
		return "", "", fmt.Errorf("expected style property after ~ at rune %d", p.pos)
	}
	key := string(p.src[start:p.pos])

	if err := p.expect('='); err != nil {
		return "", "", err
	}
	if p.eof() {
		return "", "", fmt.Errorf("expected style value for %q at rune %d", key, p.pos)
	}

	var value string
	switch p.peek() {
	case '"', '\'':
		quote := p.next()
		start = p.pos
		for !p.eof() && p.peek() != quote {
			p.pos++
		}
		if p.eof() {
			return "", "", fmt.Errorf("unterminated quoted value for style %q", key)
		}
		value = string(p.src[start:p.pos])
		p.pos++
	default:
		start = p.pos
		for !p.eof() {
			r := p.peek()
			if unicode.IsSpace(r) || r == '|' || r == '[' || r == ']' {
				break
			}
			p.pos++
		}
		value = string(p.src[start:p.pos])
	}

	return key, value, nil
}

func (p *barkParser) parseRawTextBody(tag string) (string, error) {
	start := p.pos
	depth := 0
	for !p.eof() {
		switch p.peek() {
		case '[':
			depth++
			p.pos++
		case ']':
			if depth == 0 {
				raw := string(p.src[start:p.pos])
				p.pos++
				return raw, nil
			}
			depth--
			p.pos++
		default:
			p.pos++
		}
	}
	return "", fmt.Errorf("unterminated raw-text element <%s>", tag)
}

func (p *barkParser) escapedBodyRune() (rune, bool) {
	if p.pos+1 >= len(p.src) || p.src[p.pos] != '\\' {
		return 0, false
	}
	switch p.src[p.pos+1] {
	case '[', ']':
		return p.src[p.pos+1], true
	default:
		return 0, false
	}
}

func (p *barkParser) parseIDShortcut() (string, error) {
	if err := p.expect('@'); err != nil {
		return "", err
	}
	start := p.pos
	for !p.eof() {
		r := p.peek()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			p.pos++
			continue
		}
		break
	}
	if start == p.pos {
		return "", fmt.Errorf("expected id after @ at rune %d", p.pos)
	}
	return string(p.src[start:p.pos]), nil
}

func (p *barkParser) readUntil(stop rune) string {
	start := p.pos
	for !p.eof() && p.peek() != stop {
		p.pos++
	}
	return string(p.src[start:p.pos])
}

func (p *barkParser) readName() string {
	start := p.pos
	for !p.eof() && barkIsNamePart(p.peek()) {
		p.pos++
	}
	return string(p.src[start:p.pos])
}

func (p *barkParser) skipSpaces() {
	for !p.eof() && unicode.IsSpace(p.peek()) {
		p.pos++
	}
}

func (p *barkParser) expect(want rune) error {
	if p.eof() || p.peek() != want {
		return fmt.Errorf("expected %q at rune %d", want, p.pos)
	}
	p.pos++
	return nil
}

func (p *barkParser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *barkParser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *barkParser) next() rune {
	r := p.peek()
	p.pos++
	return r
}

func barkIsNameStart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func barkIsNamePart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_'
}

func barkIsClassNamePart(r rune) bool {
	return barkIsNamePart(r)
}

func barkIsAttrNamePart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ':'
}

func barkIsStyleNamePart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_'
}

func barkIsRawTextTag(tag string) bool {
	switch strings.ToLower(tag) {
	case "script", "style":
		return true
	default:
		return false
	}
}

func barkEscapeHTML(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
