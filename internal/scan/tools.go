package scan

import (
	"fmt"
	"sheettracer/internal/db"
	"sheettracer/internal/sheets"
	"strings"

	"github.com/xuri/efp"
)

// ImportRangeTarget holds extracted arguments for a single IMPORTRANGE call.
type ImportRangeTarget struct {
	SpreadsheetID string // Raw expression for parameter 1 (URL / ID)
	Range         string // Raw expression for parameter 2 (Range string)
}

// ExtractImportRanges parses a Google Sheets formula and returns all IMPORTRANGE calls found.
func ExtractImportRanges(formula string) ([]ImportRangeTarget, error) {
	ps := efp.ExcelParser()
	tokens := ps.Parse(formula)

	var results []ImportRangeTarget

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		// Look for an IMPORTRANGE function start token
		if tok.TType == efp.TokenTypeFunction &&
			tok.TSubType == efp.TokenSubTypeStart &&
			strings.EqualFold(tok.TValue, "IMPORTRANGE") {

			call, nextIdx := parseImportRangeArguments(tokens, i)
			results = append(results, call)
			i = nextIdx // Jump parser index past this function call
		}
	}

	return results, nil
}

// parseImportRangeArguments extracts parameters by tracking function nesting depth and commas.
func parseImportRangeArguments(tokens []efp.Token, startIdx int) (ImportRangeTarget, int) {
	var call ImportRangeTarget
	var currentArg strings.Builder

	args := make([]string, 0, 2)
	depth := 0

	i := startIdx + 1
	for ; i < len(tokens); i++ {
		tok := tokens[i]

		// Track nested function depth
		if tok.TSubType == efp.TokenSubTypeStart {
			depth++
		} else if tok.TSubType == efp.TokenSubTypeStop {
			depth--
			if depth < 0 {
				// Reached the closing parenthesis of the outer IMPORTRANGE call
				if currentArg.Len() > 0 {
					args = append(args, strings.TrimSpace(currentArg.String()))
				}
				break
			}
		}

		// Top-level commas (depth == 0) mark parameter separators
		if depth == 0 && tok.TType == efp.TokenTypeArgument {
			args = append(args, strings.TrimSpace(currentArg.String()))
			currentArg.Reset()
			continue
		}

		// Reconstruct raw argument expression text
		currentArg.WriteString(tok.TValue)
	}

	if len(args) > 0 {
		c := cleanArgument(args[0])
		id, err := sheets.ParseID(c)
		if err != nil {
			fmt.Errorf("[ParseID]:%v", err)
		}
		call.SpreadsheetID = id
	}
	if len(args) > 1 {
		call.Range = cleanArgument(args[1])
	}

	return call, i
}

// cleanArgument strips wrapping quotes from literal string arguments while leaving dynamic expressions intact.
func cleanArgument(arg string) string {
	if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") && len(arg) >= 2 {
		return arg[1 : len(arg)-1]
	}
	return arg
}

// ToA1Notation converts 0-based row and col indices to A1 notation (e.g., row 0, col 0 -> "A1").
func toA1Notation(row, col int) string {
	var colStr strings.Builder

	// Convert 0-based column index to Base-26 letters (A-Z, AA-ZZ, etc.)
	colNum := col + 1
	for colNum > 0 {
		colNum-- // Adjust for 1-based indexing
		colStr.WriteByte(byte('A' + (colNum % 26)))
		colNum /= 26
	}

	// Reverse the column letters since they were extracted right-to-left
	runes := []rune(colStr.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return fmt.Sprintf("%s%d", string(runes), row+1)
}

func isFormulaStrict(val string) bool {
	trimmed := strings.TrimSpace(val)
	return strings.HasPrefix(trimmed, "=") && !strings.HasPrefix(trimmed, "'=")
}

func ScanValues(tab db.Tab, val [][]interface{}) []db.Edge {
	fmt.Printf("[TabIDScan]:%v\n", tab.TabID)

	var e []db.Edge
	for irow, row := range val {
		for icol, col := range row {
			s, ok := col.(string)
			if !ok {
				continue
			}
			if !isFormulaStrict(s) {
				continue
			}
			extracts, err := ExtractImportRanges(s)
			if err != nil {
				fmt.Errorf("[Formula parse]:%v\n", err)
			}
			src_cell := toA1Notation(irow, icol)
			// fmt.Printf("[Cell]:%v\n", src_cell)
			for _, target := range extracts {
				// fmt.Printf("	[Target]=> [ID]:%v;[Range]:%v\n", target.SpreadsheetID, target.Range)
				e = append(e, db.Edge{
					SourceSpreadsheet: tab.SpreadsheetID,
					SourceTabID:       tab.TabID,
					SourceCell:        src_cell,
					TargetGoogleID:    target.SpreadsheetID,
					TargetRange:       &target.Range,
				})
			}

			// fmt.Printf("[extracts]: %v\n", extracts)
			// fmt.Println("")
			_ = src_cell
		}
	}

	// fmt.Printf("\n[e]: %v\n", e)
	return e
}
