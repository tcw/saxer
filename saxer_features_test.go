package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitCodes(t *testing.T) {
	run := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		cmd := newRootCmd()
		cmd.SetArgs(args)
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		code := execute(cmd, &errOut)
		return code, out.String(), errOut.String()
	}
	code, out, _ := run("Fuel", exampleFile)
	assert.Equal(t, exitMatch, code)
	assert.NotEmpty(t, out)

	code, out, stderr := run("nomatch", exampleFile)
	assert.Equal(t, exitNoMatch, code)
	assert.Empty(t, out)
	assert.Empty(t, stderr)

	code, out, _ = run("-n", "nomatch", exampleFile)
	assert.Equal(t, exitNoMatch, code)
	assert.Equal(t, "0\n", out)

	code, _, stderr = run("car", "testdata/does-not-exist.xml")
	assert.Equal(t, exitError, code)
	assert.Contains(t, stderr, "does-not-exist.xml")

	code, _, stderr = run("", exampleFile)
	assert.Equal(t, exitError, code)
	assert.Contains(t, stderr, "empty query")

	code, _, _ = run()
	assert.Equal(t, exitError, code)
}

func TestNoMatchIsReported(t *testing.T) {
	_, err := runSaxer(t, nil, "nomatch", exampleFile)
	assert.ErrorIs(t, err, errNoMatch)
}

func TestEmptyQueryIsAnError(t *testing.T) {
	for _, q := range []string{"", "/", " "} {
		_, err := runSaxer(t, nil, q, exampleFile)
		require.Error(t, err)
		assert.NotErrorIs(t, err, errNoMatch)
	}
}

func gzipped(t *testing.T, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	_, err := zw.Write(data)
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return b.Bytes()
}

func TestGzipInput(t *testing.T) {
	data, err := os.ReadFile(exampleFile)
	require.NoError(t, err)
	gz := gzipped(t, data)

	got, err := runSaxer(t, bytes.NewReader(gz), "car")
	require.NoError(t, err)
	assert.Equal(t, lines(car1, car2, car3), got)

	path := filepath.Join(t.TempDir(), "example.xml.gz")
	require.NoError(t, os.WriteFile(path, gz, 0o644))
	got, err = runSaxer(t, nil, "-n", "Fuel", path)
	require.NoError(t, err)
	assert.Equal(t, "3\n", got)

	_, err = runSaxer(t, bytes.NewReader(gz[:len(gz)/2]), "car")
	assert.Error(t, err)
}

func TestShortInput(t *testing.T) {
	for _, in := range []string{"", "<", "\x1f"} {
		_, err := runSaxer(t, strings.NewReader(in), "car")
		assert.Error(t, err, in)
	}
	got, err := runSaxer(t, strings.NewReader("<a/>"), "a")
	require.NoError(t, err)
	assert.Equal(t, "<a/>\n", got)
}

type jsonMatch struct {
	Content    string            `json:"content"`
	LineStart  int               `json:"line_start"`
	LineEnd    int               `json:"line_end"`
	Path       string            `json:"path"`
	Attributes map[string]string `json:"attributes"`
}

func decodeJSONLines(t *testing.T, out string) []jsonMatch {
	t.Helper()
	var got []jsonMatch
	for _, line := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
		var m jsonMatch
		require.NoError(t, json.Unmarshal([]byte(line), &m), line)
		got = append(got, m)
	}
	return got
}

func TestJSONLines(t *testing.T) {
	out, err := runSaxer(t, nil, "-j", "engine", exampleFile)
	require.NoError(t, err)
	assert.Equal(t, `{"content":"<engine nr=\"001\">\n\t\t\t<Fuel>Gasoline</Fuel>\n\t\t</engine>","line_start":5,"line_end":7,"path":"cars/car/engine","attributes":{"nr":"001"}}`,
		strings.Split(out, "\n")[0])
	got := decodeJSONLines(t, out)
	require.Len(t, got, 3)
	assert.Equal(t, jsonMatch{engine3, 19, 21, "cars/car/engine", map[string]string{"nr": "003"}}, got[2])

	out, err = runSaxer(t, nil, "--json", "-i", "car?man=Saab", exampleFile)
	require.NoError(t, err)
	got = decodeJSONLines(t, out)
	require.Len(t, got, 1)
	assert.Equal(t, map[string]string{"vin": "wp033", "man": "Saab"}, got[0].Attributes)
	assert.True(t, strings.HasPrefix(got[0].Content, "\n\t\t<color>yellow</color>"))
}

func TestJSONEscaping(t *testing.T) {
	doc := "<a><b k=\"&quot;q&quot; &amp; &#x263A;\" e=\"\">\"\\\t\x01 é \xff</b></a>"
	out, err := runSaxer(t, strings.NewReader(doc), "-j", "b")
	require.NoError(t, err)
	got := decodeJSONLines(t, out)
	require.Len(t, got, 1)
	assert.Equal(t, "<b k=\"&quot;q&quot; &amp; &#x263A;\" e=\"\">\"\\\t\x01 é �</b>", got[0].Content)
	assert.Equal(t, map[string]string{"k": `"q" & ☺`, "e": ""}, got[0].Attributes)
}

func TestJSONConflicts(t *testing.T) {
	for _, flag := range []string{"-m", "-w"} {
		_, err := runSaxer(t, nil, "-j", flag, "car", exampleFile)
		assert.ErrorContains(t, err, "--json", flag)
	}
}

func TestSingleLineAndSquash(t *testing.T) {
	doc := "<a><b>x\r\ny\rz\n  w</b></a>"
	got, err := runSaxer(t, strings.NewReader(doc), "-l", "b")
	require.NoError(t, err)
	assert.Equal(t, "<b>x y z   w</b>\n", got)

	got, err = runSaxer(t, strings.NewReader("<a><b>  x \r\n\t y </b></a>"), "--squash", "b")
	require.NoError(t, err)
	assert.Equal(t, "<b> x y </b>\n", got)

	got, err = runSaxer(t, nil, "--squash", "-i", "engine", exampleFile)
	require.NoError(t, err)
	assert.Equal(t, lines("<Fuel>Gasoline</Fuel>", "<Fuel>Diesel</Fuel>", "<Fuel>Diesel</Fuel>"), got)
}

func TestQueryFeatures(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"wildcard", []string{"car/*/Fuel"}, lines(fuel1, fuel2, fuel3)},
		{"wildcard level", []string{"-n", "cars/*"}, lines("4")},
		{"text", []string{"color~red"}, lines("<color>red</color>")},
		{"text with path", []string{"-i", "engine/Fuel~Diesel"}, lines("Diesel", "Diesel")},
		{"text of child", []string{"-m", "car/color~blue"}, lines("3-3\tcars/car/color")},
		{"text of element contains", []string{"-m", "-c", "cars/car~yellow"}, lines("16-22\tcars/car")},
		{"text contains", []string{"-c", "-s", "-n", "Fuel~DIES"}, lines("2")},
		{"unescaped text", []string{`info~<some-xml>data<\/some-xml>`}, lines("<info>&lt;some-xml>data&lt;/some-xml></info>")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runSaxer(t, nil, append(tt.args, exampleFile)...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLiteralAmpersandInQuery(t *testing.T) {
	doc := `<a><b name="Owner &amp; Co">1</b><b name="Owner">2</b></a>`
	got, err := runSaxer(t, strings.NewReader(doc), `b?name=Owner \& Co`)
	require.NoError(t, err)
	assert.Equal(t, "<b name=\"Owner &amp; Co\">1</b>\n", got)
}

func TestUnescapeCharacterReferences(t *testing.T) {
	got, err := runSaxer(t, strings.NewReader(`<a><b>&#60;x&#x3E; &amp;lt;</b></a>`), "-u", "-i", "b")
	require.NoError(t, err)
	assert.Equal(t, "<x> &lt;\n", got)
}

func TestVersion(t *testing.T) {
	assert.NotEmpty(t, buildVersion())
	old := version
	defer func() { version = old }()
	version = "v9.9.9"
	assert.Equal(t, "v9.9.9", buildVersion())
}

func TestVersionFlag(t *testing.T) {
	oldV, oldC, oldD := version, commit, date
	defer func() { version, commit, date = oldV, oldC, oldD }()
	version, commit, date = "1.2.3", "abc123", "2026-11-10T00:00:00Z"
	got, err := runSaxer(t, strings.NewReader(""), "--version")
	require.NoError(t, err)
	assert.Equal(t, "saxer 1.2.3\ncommit: abc123\nbuilt: 2026-11-10T00:00:00Z\n", got)
}

func TestSquashFunctions(t *testing.T) {
	assert.Equal(t, "a b c", string(squash(nil, []byte(" \n a \t\r\n b c\n"))))
	assert.Equal(t, "", string(squash(nil, []byte(" \n "))))
	assert.Equal(t, "a b c d ", string(singleLine(nil, []byte("a\nb\r\nc\rd\n"))))
}
