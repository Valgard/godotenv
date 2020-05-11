// Manages .env files.
package godotenv

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Valgard/go-pcre"
	"github.com/Valgard/godotenv/internal"
)

const (
	stateVarname int = iota
	stateValue
)

const (
	regexVarname    string = `(?i:[A-Z][A-Z0-9_]*+)`
	regexEmptyLines string = `(?:\s*+(?:#[^\n]*+)?+)++`
)

type dotEnv struct {
	path       string
	cursor     int
	lineno     int
	data       string
	end        int
	values     map[string]string
	loadedVars map[string]bool
	envKey     string
	defaultEnv string
	debugKey   string
	prodEnvs   []string
	testEnvs   []string
}

// Creates new dotEnv instance
func New(opts ...option) *dotEnv {
	env := &dotEnv{
		path:       "",
		cursor:     0,
		lineno:     0,
		data:       "",
		end:        0,
		values:     map[string]string{},
		loadedVars: map[string]bool{},
		envKey:     "APP_ENV",
		debugKey:   "APP_DEBUG",
		prodEnvs:   []string{"prod"},
		defaultEnv: "dev",
	}

	env.Option(opts...)

	return env
}

func init() {
	dotenv = New()
}

func (d *dotEnv) Option(opts ...option) (previous []option) {
	for _, opt := range opts {
		previous = append(previous, opt(d))
	}

	return previous
}

func (d *dotEnv) Load(path string, extraPaths ...string) error {
	return d.doLoad(false, append([]string{path}, extraPaths...)...)
}

func (d *dotEnv) LoadEnv(path string, opts ...option) error {
	previous := d.Option(opts...)
	defer d.Option(previous...)

	env := os.Getenv(d.envKey)

	if p := fmt.Sprintf("%s.%s", path, "dist"); internal.IsFile(path) || !internal.IsFile(p) {
		if err := d.Load(path); err != nil {
			return err
		}
	} else {
		if err := d.Load(p); err != nil {
			return err
		}
	}

	if env == "" {
		env = d.defaultEnv
		if err := d.Populate(map[string]string{d.envKey: env}, false); err != nil {
			return nil
		}
	}

	if p := fmt.Sprintf("%s.%s", path, "local"); !internal.Contains(d.testEnvs, env) && internal.IsFile(p) {
		if err := d.Load(p); err != nil {
			return err
		}
		if e := os.Getenv(d.envKey); e != "" {
			env = e
		}
	}

	if env == "local" {
		return nil
	}

	if p := fmt.Sprintf("%s.%s", path, env); internal.IsFile(p) {
		if err := d.Load(p); err != nil {
			return err
		}
	}

	if p := fmt.Sprintf("%s.%s.%s", path, env, "local"); internal.IsFile(p) {
		if err := d.Load(p); err != nil {
			return err
		}
	}

	return nil
}

func (d *dotEnv) Overload(path string, extraPaths ...string) error {
	return d.doLoad(true, append([]string{path}, extraPaths...)...)
}

func (d *dotEnv) Populate(values map[string]string, overrideExistingVars bool) error {
	var (
		currentEnv = map[string]string{}
	)

	for _, line := range os.Environ() {
		key := strings.Split(line, "=")[0]
		value := strings.Split(line, "=")[1]
		currentEnv[key] = value
	}

	for name, value := range values {
		_, varLoaded := d.loadedVars[name]
		_, envExists := currentEnv[name]
		if !varLoaded && (!overrideExistingVars && envExists) {
			continue
		}

		_ = os.Setenv(name, value)

		if !varLoaded {
			d.loadedVars[name] = true
		}
	}

	return nil
}

func (d *dotEnv) Parse(data string, path string) (map[string]string, error) {
	var (
		err   error
		state int    = stateVarname
		name  string = ""
		value string = ""
	)
	defer func() {
		d.values = map[string]string{}
		d.data = ""
		d.path = ""
	}()

	d.path = path
	d.data = strings.ReplaceAll(strings.ReplaceAll(data, "\r\n", "\n"), "\r", "\n")
	d.lineno = 1
	d.cursor = 0
	d.end = len(d.data)
	d.values = map[string]string{}

	d.skipEmptyLines()

	for d.cursor < d.end {
		switch state {
		case stateVarname:
			name, err = d.lexVarname()
			if err != nil {
				return d.values, err
			}
			state = stateValue
		case stateValue:
			value, err = d.lexValue()
			d.values[name] = value
			if err != nil {
				return d.values, err
			}
			state = stateVarname
		}
	}

	if state == stateValue {
		d.values[name] = ""
	}

	return d.values, nil
}

func (d *dotEnv) lexVarname() (string, error) {
	match, ok := internal.Match(`(export[ \t]++)?(`+regexVarname+`)`, d.data, pcre.ANCHORED, d.cursor)
	if !ok {
		return "", d.error("invalid character in variable name")
	}
	d.moveCursor(match.Finding)

	if d.cursor == d.end || '\n' == d.data[d.cursor] || '#' == d.data[d.cursor] {
		if match.Groups[1].Finding != "" {
			return "", d.error("unable to unset an environment variable")
		}

		return "", d.error("missing = in the environment variable declaration")
	}

	if ' ' == d.data[d.cursor] || '\t' == d.data[d.cursor] {
		return "", d.error("whitespace characters are not supported after the variable name")
	}

	if '=' != d.data[d.cursor] {
		return "", d.error("missing = in the environment variable declaration")
	}

	d.cursor++

	return match.Groups[2].Finding, nil
}

func (d *dotEnv) lexValue() (string, error) {
	match, ok := internal.Match(`[ \t]*+(?:#.*)?$`, d.data, pcre.ANCHORED + pcre.MULTILINE, d.cursor)
	if ok {
		d.moveCursor(match.Finding)
		d.skipEmptyLines()

		return "", nil
	}

	if ' ' == d.data[d.cursor] || '\t' == d.data[d.cursor] {
		return "", d.error("whitespace are not supported before the value")
	}

	v := ""

	for {
		switch {
		case '\'' == d.data[d.cursor]:
			length := 0

			for {
				length++
				if d.cursor +length == d.end {
					d.cursor+= length

					return "", d.error("missing quote to end the value")
				}

				if !('\'' == d.data[d.cursor +length]) {
					break
				}
			}

			v = d.data[1 + d.cursor : length- 1]
			d.cursor+= 1 + length
		case '"' == d.data[d.cursor]:
			value := ""

			d.cursor++
			if d.cursor == d.end {
				return "", d.error("missing quote to end the value")
			}

			for ok := true; ok; ok = '"' != d.data[d.cursor] || ('\\' == d.data[d.cursor - 1] && '\\' != d.data[d.cursor - 2]) {
				value+= string(d.data[d.cursor])
				d.cursor++

				if d.cursor == d.end {
					return "", d.error("missing quote to end the value")
				}
			}

			d.cursor++

			var err error
			value = strings.ReplaceAll(value, "\\\"", "\"")
			value = strings.ReplaceAll(value, "\\r", "\r")
			value = strings.ReplaceAll(value, "\\n", "\n")

			resolvedValue := value

			resolvedValue, err = d.resolveVariables(resolvedValue)
			if err != nil {
				return "", err
			}

			resolvedValue, err = d.resolveCommands(resolvedValue)
			if err != nil {
				return "", err
			}

			resolvedValue = strings.ReplaceAll(resolvedValue, "\\\\", "\\")

			v+= resolvedValue
		default:
			value := ""
			prevChr := d.data[d.cursor - 1]
			for ok := true; ok; ok = d.cursor < d.end && !internal.Contains([]string{"\n", "\"", "'"}, string(d.data[d.cursor])) && !((' ' == prevChr || '\t' == prevChr) && '#' == d.data[d.cursor]) {
				if '\\' == d.data[d.cursor] && len(d.data) > d.cursor + 1 && ('"' == d.data[d.cursor + 1] || '\'' == d.data[d.cursor + 1]) {
					d.cursor++
				}

				prevChr = d.data[d.cursor]
				value+= string(d.data[d.cursor])

				if '$' == d.data[d.cursor] && len(d.data) > d.cursor + 1 && '(' == d.data[d.cursor + 1] {
					d.cursor++
					nestedValue, err := d.lexNestedExpression()
					if err != nil {
						return "", err
					}
					value+= "(" + nestedValue + ")"
				}

				d.cursor++
			}

			var err error
			value = strings.TrimRight(value, " \t\n\r\x00\x0B")

			resolvedValue := value

			resolvedValue, err = d.resolveVariables(resolvedValue)
			if err != nil {
				return "", err
			}

			resolvedValue, err = d.resolveCommands(resolvedValue)
			if err != nil {
				return "", err
			}

			resolvedValue = strings.ReplaceAll(resolvedValue, "\\\\", "\\")

			_, matchOk := internal.Match("\\s+", value, 0, 0)
			if resolvedValue == value && matchOk {
				return "", d.error("a value containing spaces must be surrounded by quotes")
			}

			v+= resolvedValue

			if d.cursor < d.end && '#' == d.data[d.cursor] {
				break
			}
		}

		if !(d.cursor < d.end && '\n' != d.data[d.cursor]) {
			break
		}
	}

	d.skipEmptyLines()

	return v, nil
}

func (d *dotEnv) lexNestedExpression() (string, error) {
	panic("lexNestedExpression: implement me")
}

func (d *dotEnv) skipEmptyLines() {
	match, ok := internal.Match(regexEmptyLines, d.data, pcre.ANCHORED, d.cursor)
	if ok {
		d.moveCursor(match.Finding)
	}
}

func (d *dotEnv) resolveCommands(value string) (string, error) {
	if ! strings.Contains(value, "$") {
		return value, nil
	}

	return value, nil
}

func (d *dotEnv) resolveVariables(value string) (string, error) {
	if ! strings.Contains(value, "$") {
		return value, nil
	}

	regex := `
		(?<!\\)
		(?P<backslashes>\\*)               # escaped with a backslash?
		\$
		(?!\()                             # no opening parenthesis
		(?P<opening_brace>\{)?             # optional brace
		(?P<name>` + regexVarname + `)?   # var name
		(?P<default_value>:[-=][^\}]++)?   # optional default value
		(?P<closing_brace>\})?             # optional closing brace
	`

	callback := func(matches pcre.Match) (string, error) {
		// odd number of backslashes means the $ character is escaped
		if 1 == len(matches.Groups[matches.NamedGroups["backslashes"]].Finding) % 2 {
			return matches.Finding[1:], nil
		}

		// unescaped $ not followed by variable name
		if 0 == len(matches.Groups[matches.NamedGroups["name"]].Finding) {
			return matches.Finding, nil
		}

		if "{" == matches.Groups[matches.NamedGroups["opening_brace"]].Finding && 0 == len(matches.Groups[matches.NamedGroups["closing_brace"]].Finding) {
			return "", d.error("unclosed braces on variable expansion")
		}

		name := matches.Groups[matches.NamedGroups["name"]].Finding

		value, ok := d.values[name]
		if ! ok {
			value = os.Getenv(name)
		}

		if "" == value && "" != matches.Groups[matches.NamedGroups["default_value"]].Finding {
			unsupportedChars := "'\"{$"
			if strings.ContainsAny(value, unsupportedChars) {
				return "", d.error(fmt.Sprintf("Unsupported character %q found in the default value of variable \"$%s\"", value[strings.IndexAny(value, unsupportedChars)], name))
			}

			value = matches.Groups[matches.NamedGroups["default_value"]].Finding[2:]

			if '=' == matches.Groups[matches.NamedGroups["default_value"]].Finding[1] {
				d.values[name] = value
			}
		}

		if 0 == len(matches.Groups[matches.NamedGroups["opening_brace"]].Finding)  && 0 != len(matches.Groups[matches.NamedGroups["closing_brace"]].Finding) {
			value+= "}"
		}

		return matches.Groups[matches.NamedGroups["backslashes"]].Finding + value, nil
	}

	var err error
	value, err = internal.ReplaceCallback(regex, value, pcre.EXTENDED, callback)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (d *dotEnv) moveCursor(text string) {
	d.cursor += len(text)
	d.lineno += strings.Count(text, "\n")
}

func (d *dotEnv) doLoad(overrideExistingVars bool, paths ...string) error {
	for _, path := range paths {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return &PathError{
				Path: path,
				Err:  err,
			}
		}

		data, err := d.Parse(string(content), path)
		if err != nil {
			return err
		}

		if err := d.Populate(data, overrideExistingVars); err != nil {
			return err
		}
	}

	return nil
}

func (d *dotEnv) error(message string) ParseError {
	return ParseError{
		Line:     d.lineno,
		Position: d.cursor,
		Message:  message,
	}
}
