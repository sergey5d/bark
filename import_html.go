package main

import (
	"fmt"
	"strings"
	"unicode"
)

// HTML -> Bark

type htmlNode interface {
	isHTMLNode()
}

type htmlTextNode struct {
	Text string
}

func (htmlTextNode) isHTMLNode() {}

type htmlElementNode struct {
	Tag      string
	Attrs    [][2]string
	Children []htmlNode
}

func (htmlElementNode) isHTMLNode() {}

type htmlParser struct {
	src string
	pos int
}

func ConvertHTMLToBarkGo(source string) (string, error) {
	p := &htmlParser{src: source}
	nodes, err := p.parseNodes("")
	if err != nil {
		return "", err
	}

	var lines []string
	for _, n := range nodes {
		lines = append(lines, formatHTMLNodeAsBark(n, 0)...)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func (p *htmlParser) parseNodes(stopTag string) ([]htmlNode, error) {
	var nodes []htmlNode
	for !p.eof() {
		if p.startsWith("<!--") {
			p.skipComment()
			continue
		}
		if p.startsWith("<!") {
			p.skipDeclaration()
			continue
		}
		if stopTag != "" && p.startsWith("</") {
			tag, err := p.parseClosingTag()
			if err != nil {
				return nil, err
			}
			if strings.EqualFold(tag, stopTag) {
				return nodes, nil
			}
			continue
		}
		if p.startsWith("<") {
			elem, selfClosing, err := p.parseStartTag()
			if err != nil {
				return nil, err
			}
			if !selfClosing && !barkVoidTags[strings.ToLower(elem.Tag)] {
				if strings.EqualFold(elem.Tag, "script") || strings.EqualFold(elem.Tag, "style") {
					raw, err := p.readRawUntilClosingTag(elem.Tag)
					if err != nil {
						return nil, err
					}
					if raw != "" {
						elem.Children = append(elem.Children, htmlTextNode{Text: barkUnescapeHTML(raw)})
					}
				} else {
					children, err := p.parseNodes(elem.Tag)
					if err != nil {
						return nil, err
					}
					elem.Children = children
				}
			}
			nodes = append(nodes, elem)
			continue
		}

		text := p.readUntil("<")
		if text != "" {
			nodes = append(nodes, htmlTextNode{Text: barkUnescapeHTML(text)})
		}
	}

	if stopTag != "" {
		return nil, fmt.Errorf("unterminated <%s>", stopTag)
	}
	return nodes, nil
}

func (p *htmlParser) parseStartTag() (htmlElementNode, bool, error) {
	if !p.consume("<") {
		return htmlElementNode{}, false, fmt.Errorf("expected '<' at %d", p.pos)
	}
	tag := p.readName()
	if tag == "" {
		return htmlElementNode{}, false, fmt.Errorf("expected tag name at %d", p.pos)
	}

	var attrs [][2]string
	selfClosing := false
	for !p.eof() {
		p.skipSpaces()
		switch {
		case p.startsWith("/>"):
			p.pos += 2
			selfClosing = true
			return htmlElementNode{Tag: tag, Attrs: attrs}, selfClosing, nil
		case p.startsWith(">"):
			p.pos++
			return htmlElementNode{Tag: tag, Attrs: attrs}, selfClosing, nil
		default:
			key := p.readAttrName()
			if key == "" {
				return htmlElementNode{}, false, fmt.Errorf("expected attribute name in <%s> at %d", tag, p.pos)
			}
			p.skipSpaces()
			value := ""
			if p.consume("=") {
				p.skipSpaces()
				value = p.readAttrValue()
			}
			attrs = append(attrs, [2]string{key, barkUnescapeHTML(value)})
		}
	}
	return htmlElementNode{}, false, fmt.Errorf("unterminated <%s>", tag)
}

func (p *htmlParser) parseClosingTag() (string, error) {
	if !p.consume("</") {
		return "", fmt.Errorf("expected closing tag at %d", p.pos)
	}
	tag := p.readName()
	p.skipSpaces()
	if !p.consume(">") {
		return "", fmt.Errorf("unterminated closing tag </%s>", tag)
	}
	return tag, nil
}

func (p *htmlParser) readUntil(token string) string {
	if idx := strings.Index(p.src[p.pos:], token); idx >= 0 {
		out := p.src[p.pos : p.pos+idx]
		p.pos += idx
		return out
	}
	out := p.src[p.pos:]
	p.pos = len(p.src)
	return out
}

func (p *htmlParser) readName() string {
	start := p.pos
	for !p.eof() {
		r := rune(p.src[p.pos])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			p.pos++
			continue
		}
		break
	}
	return p.src[start:p.pos]
}

func (p *htmlParser) readRawUntilClosingTag(tag string) (string, error) {
	lower := strings.ToLower(p.src[p.pos:])
	needle := "</" + strings.ToLower(tag) + ">"
	idx := strings.Index(lower, needle)
	if idx < 0 {
		return "", fmt.Errorf("unterminated raw <%s> block", tag)
	}
	raw := p.src[p.pos : p.pos+idx]
	p.pos += idx + len(needle)
	return raw, nil
}

func (p *htmlParser) skipComment() {
	if idx := strings.Index(p.src[p.pos:], "-->"); idx >= 0 {
		p.pos += idx + 3
		return
	}
	p.pos = len(p.src)
}

func (p *htmlParser) skipDeclaration() {
	if idx := strings.Index(p.src[p.pos:], ">"); idx >= 0 {
		p.pos += idx + 1
		return
	}
	p.pos = len(p.src)
}

func (p *htmlParser) readAttrName() string {
	start := p.pos
	for !p.eof() {
		r := rune(p.src[p.pos])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ':' {
			p.pos++
			continue
		}
		break
	}
	return p.src[start:p.pos]
}

func (p *htmlParser) readAttrValue() string {
	if p.eof() {
		return ""
	}
	switch p.src[p.pos] {
	case '"', '\'':
		quote := p.src[p.pos]
		p.pos++
		start := p.pos
		for !p.eof() && p.src[p.pos] != quote {
			p.pos++
		}
		value := p.src[start:p.pos]
		if !p.eof() {
			p.pos++
		}
		return value
	default:
		start := p.pos
		for !p.eof() {
			r := rune(p.src[p.pos])
			if unicode.IsSpace(r) || r == '>' || (r == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '>') {
				break
			}
			p.pos++
		}
		return p.src[start:p.pos]
	}
}

func (p *htmlParser) skipSpaces() {
	for !p.eof() && unicode.IsSpace(rune(p.src[p.pos])) {
		p.pos++
	}
}

func (p *htmlParser) startsWith(prefix string) bool {
	return strings.HasPrefix(p.src[p.pos:], prefix)
}

func (p *htmlParser) consume(prefix string) bool {
	if p.startsWith(prefix) {
		p.pos += len(prefix)
		return true
	}
	return false
}

func (p *htmlParser) eof() bool {
	return p.pos >= len(p.src)
}

func normalizeBarkText(text string) string {
	return strings.TrimSpace(barkWhitespaceRE.ReplaceAllString(text, " "))
}

func formatBarkAttrValue(value string) string {
	if value != "" && barkAttrValueRE.MatchString(value) {
		return value
	}
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func barkSplitFields(value string) []string {
	return strings.Fields(value)
}

func formatHTMLNodeAsBark(n htmlNode, indent int) []string {
	switch typed := n.(type) {
	case htmlTextNode:
		text := normalizeBarkText(typed.Text)
		if text == "" {
			return nil
		}
		return []string{strings.Repeat(" ", indent) + text}
	case htmlElementNode:
		return formatHTMLElementAsBark(typed, indent)
	default:
		return nil
	}
}

func formatHTMLElementAsBark(elem htmlElementNode, indent int) []string {
	indentStr := strings.Repeat(" ", indent)
	tag := elem.Tag
	if tag == "div" {
		tag = ""
	}

	var ids []string
	var classes []string
	var otherAttrs [][2]string
	for _, attr := range elem.Attrs {
		switch attr[0] {
		case "id":
			ids = barkSplitFields(attr[1])
		case "class":
			classes = barkSplitFields(attr[1])
		default:
			otherAttrs = append(otherAttrs, attr)
		}
	}

	head := "["
	if tag != "" {
		head += tag
	}
	if len(ids) > 0 {
		if head == "[" {
			head += "@" + ids[0]
			ids = ids[1:]
		}
		for _, id := range ids {
			head += " @" + id
		}
	}
	if len(classes) > 0 {
		for idx, className := range classes {
			if head == "[" && idx == 0 {
				head += ":" + className
				continue
			}
			head += " :" + className
		}
	}
	if len(otherAttrs) > 0 {
		parts := make([]string, 0, len(otherAttrs))
		for _, attr := range otherAttrs {
			parts = append(parts, attr[0]+"="+formatBarkAttrValue(attr[1]))
		}
		if head == "[" {
			head += strings.Join(parts, " ")
		} else {
			head += " " + strings.Join(parts, " ")
		}
	}
	hasMetadata := len(ids) > 0 || len(classes) > 0 || len(otherAttrs) > 0

	var children []htmlNode
	for _, child := range elem.Children {
		if t, ok := child.(htmlTextNode); ok {
			if normalizeBarkText(t.Text) == "" {
				continue
			}
		}
		children = append(children, child)
	}

	if len(children) == 0 {
		return []string{indentStr + head + "]"}
	}

	onlyText := true
	var textParts []string
	for _, child := range children {
		t, ok := child.(htmlTextNode)
		if !ok {
			onlyText = false
			break
		}
		textParts = append(textParts, normalizeBarkText(t.Text))
	}
	if onlyText {
		body := strings.Join(textParts, " ")
		if hasMetadata {
			return []string{fmt.Sprintf("%s%s | %s]", indentStr, head, body)}
		}
		if tag != "" {
			return []string{fmt.Sprintf("%s%s %s]", indentStr, head, body)}
		}
		return []string{fmt.Sprintf("%s%s%s]", indentStr, head, body)}
	}

	lines := []string{indentStr + head}
	for _, child := range children {
		lines = append(lines, formatHTMLNodeAsBark(child, indent+2)...)
	}
	lines = append(lines, indentStr+"]")
	return lines
}

func barkUnescapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&nbsp;", " ",
	)
	return replacer.Replace(s)
}
