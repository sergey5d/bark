package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		barkPrintUsage(os.Stderr, os.Args[0])
		os.Exit(2)
	}
	args := os.Args[1:]
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		barkPrintUsage(os.Stdout, os.Args[0])
		return
	}

	barkFormatSpacedDiv = false
	var positional []string
	for _, arg := range args {
		switch arg {
		case "--compact":
			barkFormatSpacedDiv = false
		case "--spaced":
			barkFormatSpacedDiv = true
		case "-h", "--help":
			barkPrintUsage(os.Stdout, os.Args[0])
			return
		default:
			positional = append(positional, arg)
		}
	}

	if len(positional) == 0 {
		barkPrintUsage(os.Stderr, os.Args[0])
		os.Exit(2)
	}

	mode := "gen"
	inputs := []string{}
	if barkIsModeArg(positional[0]) {
		if len(positional) < 2 {
			barkPrintUsage(os.Stderr, os.Args[0])
			os.Exit(2)
		}
		mode = positional[0]
		inputs = positional[1:]
	} else {
		inputs = positional
	}

	var err error
	switch mode {
	case "gen", "-g":
		err = barkGenerateHTML(inputs)
	case "format", "fmt", "-f":
		err = barkFormatFiles(inputs)
	case "import", "degen", "-i":
		err = barkReverseGenerate(inputs)
	case "-d":
		err = barkReverseGenerate(inputs)
	default:
		err = fmt.Errorf("unknown mode %q, expected gen, format, fmt, import, degen, -g, -f, -i, or -d", mode)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "bark: %v\n", err)
		os.Exit(1)
	}
}

func barkPrintUsage(w *os.File, argv0 string) {
	name := filepath.Base(argv0)
	fmt.Fprintf(w, "usage: %s [--compact|--spaced] [gen|format|fmt|import|degen|-g|-f|-i|-d] <file-or-pattern> [more-files-or-patterns...]\n", name)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Formatting style:")
	fmt.Fprintln(w, "  --compact     Emit compact bare div metadata, e.g. [:shell (default)")
	fmt.Fprintln(w, "  --spaced      Emit spaced bare div metadata, e.g. [ :shell")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Modes:")
	fmt.Fprintln(w, "  gen, -g      Generate HTML from .bark inputs (default)")
	fmt.Fprintln(w, "  format, fmt  Format .bark inputs in place")
	fmt.Fprintln(w, "  -f           Alias for format")
	fmt.Fprintln(w, "  import       Convert HTML inputs to .bark")
	fmt.Fprintln(w, "  degen, -i    Alias for import")
	fmt.Fprintln(w, "  -d           Alias for import")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s \"*.bark\"\n", name)
	fmt.Fprintf(w, "  %s *.bark\n", name)
	fmt.Fprintf(w, "  %s format *.bark\n", name)
	fmt.Fprintf(w, "  %s --spaced format *.bark\n", name)
	fmt.Fprintf(w, "  %s import \"*.html\"\n", name)
	fmt.Fprintf(w, "  %s -h\n", name)
}

func barkIsModeArg(arg string) bool {
	switch arg {
	case "gen", "format", "fmt", "import", "degen", "-g", "-f", "-i", "-d":
		return true
	default:
		return false
	}
}

func barkCollectInputs(inputs []string) ([]string, error) {
	var files []string
	for _, input := range inputs {
		matches, err := filepath.Glob(input)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			files = append(files, input)
			continue
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no input files matched")
	}
	return files, nil
}

func barkGenerateHTML(inputs []string) error {
	files, err := barkCollectInputs(inputs)
	if err != nil {
		return err
	}

	for _, src := range files {
		input, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		htmlOut, err := ParseBark(string(input))
		if err != nil {
			return fmt.Errorf("parse %s: %w", src, err)
		}
		out := strings.TrimSuffix(src, filepath.Ext(src)) + ".html"
		if err := os.WriteFile(out, []byte(htmlOut), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
	}

	return nil
}

func barkReverseGenerate(inputs []string) error {
	files, err := barkCollectInputs(inputs)
	if err != nil {
		return err
	}

	for _, src := range files {
		input, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		barkOut, err := ConvertHTMLToBarkGo(string(input))
		if err != nil {
			return fmt.Errorf("convert %s: %w", src, err)
		}
		out := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)) + ".bark"
		if err := os.WriteFile(out, []byte(barkOut), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
	}

	return nil
}

func barkFormatFiles(inputs []string) error {
	files, err := barkCollectInputs(inputs)
	if err != nil {
		return err
	}

	for _, src := range files {
		input, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		formatted, err := FormatBark(string(input))
		if err != nil {
			return fmt.Errorf("format %s: %w", src, err)
		}
		if err := os.WriteFile(src, []byte(formatted), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", src, err)
		}
	}

	return nil
}
