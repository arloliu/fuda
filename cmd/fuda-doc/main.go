package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/pager"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

var (
	targetStruct = flag.String("struct", "", "Struct name to generate docs for (required unless -tui)")
	targetPath   = flag.String("path", "", "Directory or file path containing the struct (required)")
	outputTarget = flag.String("output", "stdout", "Output target: file path or \"stdout\"")
	markdown     = flag.Bool("markdown", false, "Output in Markdown format")
	ascii        = flag.Bool("ascii", false, "Output in terminal-friendly format with ANSI colors")
	noPager      = flag.Bool("no-pager", false, "Disable built-in pager for ASCII output")
	forceColor   = flag.Bool("color", false, "Force ANSI color output even when stdout is not a TTY (useful with: | less -R)")
	tuiMode      = flag.Bool("tui", false, "Launch interactive TUI explorer (all structs if -struct is omitted)")
	showVersion  = flag.Bool("version", false, "Print version and exit")
	envSummary   = flag.Bool("env-summary", false, "Print a summary table of all env-tagged fields")
	envFile      = flag.Bool("env-file", false, "Generate a .env.example file from env-tagged fields")
	yamlDefault  = flag.Bool("yaml-default", false, "Generate a default YAML config with comments")
)

func init() {
	// Register short aliases — they share the same pointer as the long form.
	flag.StringVar(targetStruct, "s", "", "Short for -struct")
	flag.StringVar(targetPath, "p", "", "Short for -path")
	flag.StringVar(outputTarget, "o", "stdout", "Short for -output")
	flag.BoolVar(markdown, "m", false, "Short for -markdown")
	flag.BoolVar(ascii, "a", false, "Short for -ascii")
	flag.BoolVar(forceColor, "c", false, "Short for -color")
	flag.BoolVar(tuiMode, "t", false, "Short for -tui")
	flag.BoolVar(showVersion, "v", false, "Short for -version")

	flag.Usage = func() {
		_, _ = fmt.Fprint(os.Stderr, "Usage: fuda-doc [flags]\n\n")
		_, _ = fmt.Fprint(os.Stderr, "Flags:\n")
		_, _ = fmt.Fprint(os.Stderr, "  -s, --struct string    Struct name to generate docs for (required unless -tui)\n")
		_, _ = fmt.Fprint(os.Stderr, "  -p, --path string      Directory or file path containing the struct (required)\n")
		_, _ = fmt.Fprint(os.Stderr, "  -o, --output string    Output target: file path or \"stdout\" (default \"stdout\")\n")
		_, _ = fmt.Fprint(os.Stderr, "  -m, --markdown         Output in Markdown format\n")
		_, _ = fmt.Fprint(os.Stderr, "  -a, --ascii            Output in terminal-friendly format with ANSI colors\n")
		_, _ = fmt.Fprint(os.Stderr, "      --no-pager         Disable built-in pager for ASCII output\n")
		_, _ = fmt.Fprint(os.Stderr, "  -c, --color            Force ANSI color output (useful with: | less -R)\n")
		_, _ = fmt.Fprint(os.Stderr, "  -t, --tui              Launch interactive TUI explorer\n")
		_, _ = fmt.Fprint(os.Stderr, "  -v, --version          Print version and exit\n")
		_, _ = fmt.Fprint(os.Stderr, "      --env-summary      Print a summary table of all env-tagged fields\n")
		_, _ = fmt.Fprint(os.Stderr, "      --env-file         Generate a .env.example file from env-tagged fields\n")
		_, _ = fmt.Fprint(os.Stderr, "      --yaml-default     Generate a default YAML config with comments\n")
	}
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	if *showVersion {
		fmt.Println("fuda-doc " + version)

		return nil
	}

	// Utility modes: env-summary, env-file, yaml-default.
	if *envSummary || *envFile || *yamlDefault {
		return runUtility()
	}

	// In TUI mode, -struct is optional (discovers all structs); otherwise required.
	if *tuiMode {
		if *targetPath == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Error: -path flag is required")
			_, _ = fmt.Fprintln(os.Stderr)
			flag.Usage()

			return errors.New("-path flag is required")
		}

		return runTUI()
	}

	if *targetStruct == "" || *targetPath == "" {
		if *targetStruct == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Error: -struct flag is required")
		}

		if *targetPath == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Error: -path flag is required")
		}

		_, _ = fmt.Fprintln(os.Stderr)
		flag.Usage()

		return errors.New("required flags missing")
	}

	// Determine format — default to ascii if neither specified
	format := docgen.FormatASCII
	if *markdown {
		format = docgen.FormatMarkdown
	} else if *ascii {
		format = docgen.FormatASCII
	}

	// Determine if we should use the built-in pager:
	// pager is enabled when ASCII format + stdout + TTY + not disabled
	toStdout := *outputTarget == "" || *outputTarget == "stdout"
	isTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	usePager := format == docgen.FormatASCII && toStdout && isTTY && !*noPager

	if usePager {
		return runWithPager(format)
	}

	return runDirect(format, toStdout)
}

func runWithPager(format docgen.OutputFormat) error {
	// Force color output for the pager (lipgloss may disable colors for non-TTY writers)
	lipgloss.SetColorProfile(termenv.TrueColor)

	var buf bytes.Buffer

	if err := docgen.Generate(*targetStruct, *targetPath, &buf, format); err != nil {
		return err
	}

	return pager.Run(buf.String(), *targetStruct)
}

func runDirect(format docgen.OutputFormat, toStdout bool) error {
	if *forceColor {
		lipgloss.SetColorProfile(termenv.TrueColor)
	}

	var out *os.File
	var err error

	if !toStdout {
		out, err = os.Create(*outputTarget)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
	} else {
		out = os.Stdout
	}

	if genErr := docgen.Generate(*targetStruct, *targetPath, out, format); genErr != nil {
		if out != os.Stdout {
			_ = out.Close()
		}

		return genErr
	}

	if out != os.Stdout {
		_ = out.Close()
	}

	return nil
}

func runTUI() error {
	lipgloss.SetColorProfile(termenv.TrueColor)

	docs, err := docgen.ParseAll(*targetStruct, *targetPath)
	if err != nil {
		return err
	}

	return tui.Run(docs)
}

func runUtility() error {
	if *targetPath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Error: -path flag is required")
		_, _ = fmt.Fprintln(os.Stderr)
		flag.Usage()

		return errors.New("-path flag is required")
	}

	docs, err := docgen.ParseAll(*targetStruct, *targetPath)
	if err != nil {
		return err
	}

	if *envSummary {
		return docgen.PrintEnvSummary(docs, os.Stdout)
	}

	if *yamlDefault {
		return docgen.PrintDefaultYAML(docs, os.Stdout, true)
	}

	return docgen.PrintEnvFile(docs, os.Stdout)
}
