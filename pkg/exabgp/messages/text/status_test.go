package text

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var testSummaryDataFile = filepath.Join("testdata", "neighbor-summary.txt")

func TestParseSummaryTestData(t *testing.T) {
	file, err := os.ReadFile(testSummaryDataFile)
	require.NoError(t, err)

	totalLines := testGetTotalLinesInFile(t, testSummaryDataFile)
	parsedEvents, err := SummariesFromBytes(file)
	require.NoError(t, err)
	require.Equal(t, totalLines-1, len(parsedEvents))
}

func TestParseSummaryDown(t *testing.T) {
	file, err := os.ReadFile(testSummaryDataFile)
	require.NoError(t, err)

	totalLines := testGetTotalLinesInFile(t, testSummaryDataFile)
	parsedEvents, err := SummariesFromBytes(file)
	require.NoError(t, err)
	require.Equal(t, totalLines-1, len(parsedEvents))
	require.Equal(t, "down", parsedEvents[0].Status)
	require.Equal(t, "up", parsedEvents[1].Status)
	require.Equal(t, "down", parsedEvents[2].Status)
}

func TestSummaryEntryFromStringReturnsDetailedParseError(t *testing.T) {
	_, err := SummaryEntryFromString("127.0.0.1 broken")
	require.Error(t, err)

	var parseErr *ParseError
	require.True(t, errors.As(err, &parseErr))
	require.Equal(t, "summary parser", parseErr.Parser)
	require.Equal(t, "127.0.0.1 broken", parseErr.Input)
	require.Zero(t, parseErr.Line)
	require.Equal(t, "summary parser: unable to parse input: \"127.0.0.1 broken\"", err.Error())
}

func TestSummariesFromBytesReturnsLineNumberInParseError(t *testing.T) {
	data := strings.Join([]string{
		summaryHeaderLine,
		"127.0.0.1       64496        down idle                  0          0",
		"127.0.0.1 broken",
	}, "\n")

	_, err := SummariesFromBytes([]byte(data))
	require.Error(t, err)

	var parseErr *ParseError
	require.True(t, errors.As(err, &parseErr))
	require.Equal(t, 3, parseErr.Line)
	require.Contains(t, err.Error(), "at line 3")
	require.Contains(t, err.Error(), "127.0.0.1 broken")
}
