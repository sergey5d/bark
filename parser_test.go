package main

import "testing"

func TestParseBarkPreservesMixedContentOrder(t *testing.T) {
	input := `[p :experience-env
  [strong Environment:]
  Scala, Go, Python
]`

	got, err := ParseBark(input)
	if err != nil {
		t.Fatalf("ParseBark returned error: %v", err)
	}

	want := `<html><p class="experience-env">
  <strong>Environment:</strong>
  Scala, Go, Python
</p></html>`

	if got != want {
		t.Fatalf("unexpected HTML output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestParseBarkPreservesInterleavedChildrenAndText(t *testing.T) {
	input := `[p
  before
  [strong middle]
  after
]`

	got, err := ParseBark(input)
	if err != nil {
		t.Fatalf("ParseBark returned error: %v", err)
	}

	want := `<html><p>
  before
  <strong>middle</strong>
  after
</p></html>`

	if got != want {
		t.Fatalf("unexpected HTML output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}
