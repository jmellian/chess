package chess

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

// GamesFromPGN returns all PGN decoding games from the
// reader.  It is designed to be used decoding multiple PGNs
// in the same file.  An error is returned if there is an
// issue parsing the PGNs.
func GamesFromPGN(r io.Reader, verbose ...string) ([]*Game, error) {
	if len(verbose) == 0 {
		verbose = append(verbose, "true")
	}
	games := []*Game{}
	current := ""
	count := 0
	totalCount := 0
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "" {
			count++
		} else {
			current += line
		}
		if count == 2 {
			game, err := decodePGN(current)
			if err != nil {
				return nil, err
			}
			games = append(games, game)
			count = 0
			current = ""
			totalCount++
			if verbose[0] == "true" {
				log.Println("Processed game", totalCount)
			}
		}
	}
	return games, nil
}

// LimitGamesFromPGN returns <nrOfGames PGN decoding games from the
// reader.  It is designed to be used decoding multiple PGNs
// in the same file.  An error is returned if there is an
// issue parsing the PGNs.
// If verbose is set to true then log information is printed
// br is the pointer to the next game to read from the pgn file
func LimitedGamesFromPGN(br *bufio.Reader, nrOfGames int, verbose bool) ([]*Game, *bufio.Reader, int, error) {
	games := []*Game{}
	current := ""
	count := 0
	totalCount := 0
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, 0, err
		}
		if strings.TrimSpace(line) == "" {
			count++
		} else {
			current += line
		}
		if count == 2 {
			game, err := decodePGN(current)
			if err != nil {
				return nil, nil, 0, err
			}
			games = append(games, game)
			count = 0
			current = ""
			totalCount++
			if verbose {
				log.Println("Processed game", totalCount)
			}
			if totalCount < nrOfGames {
			} else {
				return games, br, totalCount, nil
			}
		}
	}
	return games, nil, totalCount, nil
}

// Same result as LimitedGamesFromPGN but no logging information is printed
func LimitedGamesFromPGNReader(br *bufio.Reader, nrOfGames int) ([]*Game, *bufio.Reader, int, error) {
	return LimitedGamesFromPGN(br, nrOfGames, false)
}

func decodePGN(pgn string) (*Game, error) {
	tagPairs := getTagPairs(pgn)
	moveStrs, outcome := moveList(pgn)
	g := NewGame(TagPairs(tagPairs))
	g.ignoreAutomaticDraws = true
	var notation Notation = AlgebraicNotation{}
	if len(moveStrs) > 0 {
		_, err := LongAlgebraicNotation{}.Decode(g.Position(), moveStrs[0])
		if err == nil {
			notation = LongAlgebraicNotation{}
		}
	}
	for _, alg := range moveStrs {
		m, err := notation.Decode(g.Position(), alg)
		if err != nil {
			return nil, fmt.Errorf("chess: pgn decode error %s on move %d", err.Error(), g.Position().moveCount)
		}
		if err := g.Move(m); err != nil {
			return nil, fmt.Errorf("chess: pgn invalid move error %s on move %d", err.Error(), g.Position().moveCount)
		}
	}
	g.outcome = outcome
	return g, nil
}

func encodePGN(g *Game) string {
	s := ""
	for _, tag := range g.tagPairs {
		s += fmt.Sprintf("[%s \"%s\"]\n", tag.Key, tag.Value)
	}
	s += "\n"
	for i, move := range g.moves {
		pos := g.positions[i]
		txt := g.notation.Encode(pos, move)
		if i%2 == 0 {
			s += fmt.Sprintf("%d.%s", (i/2)+1, txt)
		} else {
			s += fmt.Sprintf(" %s ", txt)
		}
	}
	s += " " + string(g.outcome)
	return s
}

var (
	tagPairRegex = regexp.MustCompile(`\[(.*)\s\"(.*)\"\]`)
)

func getTagPairs(pgn string) []*TagPair {
	tagPairs := []*TagPair{}
	matches := tagPairRegex.FindAllString(pgn, -1)
	for _, m := range matches {
		results := tagPairRegex.FindStringSubmatch(m)
		if len(results) == 3 {
			pair := &TagPair{
				Key:   results[1],
				Value: results[2],
			}
			tagPairs = append(tagPairs, pair)
		}
	}
	return tagPairs
}

var (
	moveNumRegex = regexp.MustCompile(`(?:\d+\.+)?(.*)`)
)

func moveList(pgn string) ([]string, Outcome) {
	// remove comments
	text := removeSection("{", "}", pgn)
	// remove variations
	text = removeSection(`\(`, `\)`, text)
	// remove tag pairs
	text = removeSection(`\[`, `\]`, text)
	// remove line breaks
	text = strings.Replace(text, "\n", " ", -1)

	list := strings.Split(text, " ")
	filtered := []string{}
	var outcome Outcome
	for _, move := range list {
		move = strings.TrimSpace(move)
		switch move {
		case string(NoOutcome), string(WhiteWon), string(BlackWon), string(Draw):
			outcome = Outcome(move)
		case "":
		default:
			results := moveNumRegex.FindStringSubmatch(move)
			if len(results) == 2 && results[1] != "" {
				filtered = append(filtered, results[1])
			}
		}
	}
	return filtered, outcome
}

func removeSection(leftChar, rightChar, s string) string {
	r := regexp.MustCompile(leftChar + ".*?" + rightChar)
	for {
		i := r.FindStringIndex(s)
		if i == nil {
			return s
		}
		s = s[0:i[0]] + s[i[1]:len(s)]
	}
}

func splitPGN(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// http://www.saremba.de/chessgml/standards/pgn/pgn-complete.htm
	if atEOF {
		return 0, nil, nil
	}

	if len(data) == 0 {
		return 0, nil, nil
	}

	// Read Tag pair section:  [Left bracket, tag name, tag value, and right bracket]
	pos := 0
	for data[pos] == '[' {
		if i := bytes.IndexByte(data[pos:], '\n'); i >= 0 {
			// We have a full newline-terminated line.
			pos = pos + i + 1
			//// Request more data.
			if pos >= len(data) {
				return 0, nil, nil
			}
		} else {
			// request more data
			return 0, nil, nil
		}
	}

	pos = pos + 1
	//endOfTagPairSection := pos

	// Read Move section
	continueReadingMoveSection := true
	for continueReadingMoveSection {
		if i := bytes.IndexByte(data[pos:], '\n'); i > 0 {
			// We have a full newline-terminated line.
			pos = pos + i + 1
		} else {
			if i < 0 {
				//// Request more data.
				return 0, nil, nil
			} else {
				continueReadingMoveSection = false
			}
		}
	}
	endOfMoveSection := pos
	return pos + 1, data[0:endOfMoveSection], nil
}
