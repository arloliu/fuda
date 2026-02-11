package pager

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/colors"
)

// searchState tracks the current search query, matches, and cursor position.
type searchState struct {
	query   string // current search query (plain text)
	matches []match
	current int // index into matches (the "active" match)
	active  bool
}

// match records a single search hit by line number and byte offsets within the
// stripped (ANSI-free) version of that line.
type match struct {
	line       int // 0-based line index in the content
	startByte  int // byte offset in the stripped line
	lengthByte int // byte length of the matched text in the stripped line
}

// buildMatches scans every line of content for query (case-insensitive) and
// populates s.matches. Matching is performed on ANSI-stripped text so that
// escape sequences don't interfere with search.
func (s *searchState) buildMatches(content string) {
	s.matches = nil
	s.current = 0

	if s.query == "" {
		return
	}

	lower := strings.ToLower(s.query)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		stripped := ansi.Strip(line)
		strippedLower := strings.ToLower(stripped)
		offset := 0

		for {
			idx := strings.Index(strippedLower[offset:], lower)
			if idx < 0 {
				break
			}

			s.matches = append(s.matches, match{
				line:       i,
				startByte:  offset + idx,
				lengthByte: len(lower),
			})

			offset += idx + len(lower)
		}
	}
}

// highlightContent returns a copy of content with all matches highlighted.
// The current match uses a distinct style so the user can see which match is
// active.
func (s *searchState) highlightContent(content string) string {
	if len(s.matches) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	highlightStyle := colors.SearchHighlightStyle
	currentStyle := colors.SearchCurrentStyle

	// Group matches by line for efficient processing.
	byLine := make(map[int][]int) // line -> indices into s.matches
	for i, m := range s.matches {
		byLine[m.line] = append(byLine[m.line], i)
	}

	for lineIdx, matchIndices := range byLine {
		if lineIdx >= len(lines) {
			continue
		}

		lines[lineIdx] = highlightLine(lines[lineIdx], s.matches, matchIndices, s.current, highlightStyle, currentStyle)
	}

	return strings.Join(lines, "\n")
}

// highlightLine applies search highlights to a single line.
//
// The tricky part: the original line contains ANSI escape sequences but our
// match offsets refer to the *stripped* text. We walk through the original
// line, tracking our position in the stripped text so we can insert highlight
// markers at the right places.
func highlightLine(
	line string,
	allMatches []match,
	indices []int,
	currentIdx int,
	hlStyle, curStyle styles,
) string {
	stripped := ansi.Strip(line)

	// Build a set of byte ranges (in stripped coords) that need highlighting
	// and whether each is the current match.
	type hlRange struct {
		start, end int
		isCurrent  bool
	}

	ranges := make([]hlRange, 0, len(indices))
	for _, mi := range indices {
		m := allMatches[mi]
		ranges = append(ranges, hlRange{
			start:     m.startByte,
			end:       m.startByte + m.lengthByte,
			isCurrent: mi == currentIdx,
		})
	}

	// Walk the original line, tracking the position within the stripped text.
	// ansiRe matches ANSI escape sequences.
	ansiRe := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	locs := ansiRe.FindAllStringIndex(line, -1)

	var out strings.Builder
	out.Grow(len(line) * 2) //nolint:mnd // rough estimate

	origIdx := 0   // cursor in `line`
	stripIdx := 0  // cursor in `stripped`
	locCursor := 0 // cursor in `locs`

	// rangeAt returns the hlRange that starts at the given stripped position,
	// or nil.
	rangeAt := func(pos int) *hlRange {
		for i := range ranges {
			if ranges[i].start == pos {
				return &ranges[i]
			}
		}

		return nil
	}

	for origIdx < len(line) {
		// If we're at an ANSI escape, copy it verbatim.
		if locCursor < len(locs) && origIdx == locs[locCursor][0] {
			end := locs[locCursor][1]
			out.WriteString(line[origIdx:end])
			origIdx = end
			locCursor++

			continue
		}

		// Check if a highlight range starts here.
		if r := rangeAt(stripIdx); r != nil {
			matchText := stripped[r.start:r.end]

			style := hlStyle
			if r.isCurrent {
				style = curStyle
			}

			out.WriteString(style.Render(matchText))

			// Advance origIdx past the matched content (skipping any
			// embedded ANSI escapes).
			consumed := 0
			for consumed < len(matchText) && origIdx < len(line) {
				if locCursor < len(locs) && origIdx == locs[locCursor][0] {
					end := locs[locCursor][1]
					origIdx = end
					locCursor++

					continue
				}

				origIdx++
				consumed++
			}

			stripIdx = r.end

			continue
		}

		// Normal character — copy as-is.
		out.WriteByte(line[origIdx])
		origIdx++
		stripIdx++
	}

	return out.String()
}

// nextMatch advances to the next match, wrapping around.
func (s *searchState) nextMatch() {
	if len(s.matches) == 0 {
		return
	}

	s.current = (s.current + 1) % len(s.matches)
}

// prevMatch moves to the previous match, wrapping around.
func (s *searchState) prevMatch() {
	if len(s.matches) == 0 {
		return
	}

	s.current = (s.current - 1 + len(s.matches)) % len(s.matches)
}

// currentLine returns the line number of the current match, or -1.
func (s *searchState) currentLine() int {
	if len(s.matches) == 0 {
		return -1
	}

	return s.matches[s.current].line
}

// styles is a local alias to keep function signatures short.
type styles = lipglossStyle

type lipglossStyle = lipgloss.Style
