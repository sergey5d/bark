package main

import (
	"regexp"
)

var barkWhitespaceRE = regexp.MustCompile(`\s+`)
var barkAttrValueRE = regexp.MustCompile(`^[^=\s\[\]"]+$`)

var barkVoidTags = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}
