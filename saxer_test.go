package main

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The example file used throughout README.md.
const exampleFile = "testdata/example.xml"

// Elements of testdata/example.xml exactly as saxer emits them: the raw bytes
// of the matched element, keeping the original tab indentation.
const (
	car1 = "<car vin=\"wp031\" man=\"Volvo\">\n" +
		"\t\t<color>blue</color>\n" +
		"\t\t<xs:doors>4</xs:doors>\n" +
		"\t\t<engine nr=\"001\">\n" +
		"\t\t\t<Fuel>Gasoline</Fuel>\n" +
		"\t\t</engine>\n" +
		"\t</car>"
	car2 = "<car vin=\"wp032\" man=\"Volvo\">\n" +
		"\t\t<color>red</color>\n" +
		"\t\t<xs:doors>2</xs:doors>\n" +
		"\t\t<engine nr=\"002\">\n" +
		"\t\t\t<Fuel>Diesel</Fuel>\n" +
		"\t\t</engine>\n" +
		"\t</car>"
	car3 = "<car vin=\"wp033\" man=\"Saab\">\n" +
		"\t\t<color>yellow</color>\n" +
		"\t\t<xs:doors>4</xs:doors>\n" +
		"\t\t<engine nr=\"003\">\n" +
		"\t\t\t<Fuel>Diesel</Fuel>\n" +
		"\t\t</engine>\n" +
		"\t</car>"
	engine1 = "<engine nr=\"001\">\n\t\t\t<Fuel>Gasoline</Fuel>\n\t\t</engine>"
	engine2 = "<engine nr=\"002\">\n\t\t\t<Fuel>Diesel</Fuel>\n\t\t</engine>"
	engine3 = "<engine nr=\"003\">\n\t\t\t<Fuel>Diesel</Fuel>\n\t\t</engine>"
	fuel1   = "<Fuel>Gasoline</Fuel>"
	fuel2   = "<Fuel>Diesel</Fuel>"
	fuel3   = "<Fuel>Diesel</Fuel>"
)

// lines joins output lines the way saxer prints them, one per line.
func lines(l ...string) string {
	return strings.Join(l, "\n") + "\n"
}

// runSaxer executes the saxer command with args and stdin, returning stdout.
func runSaxer(t *testing.T, stdin io.Reader, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetIn(stdin)
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	return out.String(), err
}

// Examples from the "Queries" section of README.md.
func TestReadmeQueries(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"all car nodes", "car", lines(car1, car2, car3)},
		{"all color nodes", "color", lines("<color>blue</color>", "<color>red</color>", "<color>yellow</color>")},
		{"all engine nodes", "engine", lines(engine1, engine2, engine3)},
		{"engine nodes with nr attribute", "engine?nr", lines(engine1, engine2, engine3)},
		{"any node with nr attribute", "?nr", lines(engine1, engine2, engine3)},
		{"engine nodes with nr=003", "engine?nr=003", lines(engine3)},
		{"any node with nr=003", "?nr=003", lines(engine3)},
		{"car nodes with vin=wp031 and man=Volvo", "car?vin=wp031&man=Volvo", lines(car1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runSaxer(t, nil, tt.query, exampleFile)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Examples from the "Example usage" section of README.md.
func TestReadmeExampleUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"query by tag", []string{"engine"}, lines(engine1, engine2, engine3)},
		{"single line", []string{"-l", "engine"}, lines(
			"<engine nr=\"001\"> \t\t\t<Fuel>Gasoline</Fuel> \t\t</engine>",
			"<engine nr=\"002\"> \t\t\t<Fuel>Diesel</Fuel> \t\t</engine>",
			"<engine nr=\"003\"> \t\t\t<Fuel>Diesel</Fuel> \t\t</engine>")},
		{"tag with attribute value", []string{"engine?nr=001"}, lines(engine1)},
		{"attribute only", []string{"?man=Volvo"}, lines(car1, car2)},
		{"multiple attributes", []string{"?vin=wp031&man=Volvo"}, lines(car1)},
		{"inner xml", []string{"-i", "Fuel"}, lines("Gasoline", "Diesel", "Diesel")},
		{"count", []string{"-n", "Fuel"}, lines("3")},
		{"meta", []string{"-m", "engine"}, lines(
			"5-7    cars/car/engine",
			"12-14    cars/car/engine",
			"19-21    cars/car/engine")},
		{"first n", []string{"-f", "2", "Fuel"}, lines(fuel1, fuel2)},
		{"unescape", []string{"-u", "info"}, lines("<info><some-xml>data</some-xml></info>")},
		{"case insensitive", []string{"-s", "fuel"}, lines(fuel1, fuel2, fuel3)},
		{"omit namespace", []string{"-o", "doors"}, lines(
			"<xs:doors>4</xs:doors>", "<xs:doors>2</xs:doors>", "<xs:doors>4</xs:doors>")},
		{"contains", []string{"-c", "or"}, lines(
			"<color>blue</color>", "<xs:doors>4</xs:doors>",
			"<color>red</color>", "<xs:doors>2</xs:doors>",
			"<color>yellow</color>", "<xs:doors>4</xs:doors>")},
		{"wrap", []string{"-w", "Fuel"}, lines("<saxer-result>", fuel1, fuel2, fuel3, "</saxer-result>")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runSaxer(t, nil, append(tt.args, exampleFile)...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// README.md: "cat example.xml | saxer car"
func TestReadmeStdin(t *testing.T) {
	f, err := os.Open(exampleFile)
	require.NoError(t, err)
	defer f.Close()

	got, err := runSaxer(t, f, "car")
	require.NoError(t, err)
	assert.Equal(t, lines(car1, car2, car3), got)
}

// Keeps README.md honest: the example file it shows must be testdata/example.xml,
// and every "Command: / Returns:" block must match what saxer actually prints.
func TestReadmeIsUpToDate(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	require.NoError(t, err)
	example, err := os.ReadFile(exampleFile)
	require.NoError(t, err)

	t.Run("example file", func(t *testing.T) {
		assert.Contains(t, string(readme), indent(string(example)))
	})

	blocks := regexp.MustCompile(`Command:\n\n    saxer ([^\n]*)\n\n    Returns:\n((?:    [^\n]*\n)+)`).
		FindAllStringSubmatch(string(readme), -1)
	require.Len(t, blocks, 14, "number of Command/Returns examples in README.md")

	for _, b := range blocks {
		t.Run(b[1], func(t *testing.T) {
			args := strings.Fields(strings.ReplaceAll(b[1], `"`, ""))
			args[len(args)-1] = exampleFile
			got, err := runSaxer(t, nil, args...)
			require.NoError(t, err)
			assert.Equal(t, b[2], indent(got))
		})
	}
}

// indent prefixes every line with four spaces, as in a markdown code block.
func indent(s string) string {
	var b strings.Builder
	for _, l := range strings.SplitAfter(s, "\n") {
		if l != "" {
			b.WriteString("    " + l)
		}
	}
	return b.String()
}
