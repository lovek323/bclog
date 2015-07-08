package events

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"text/tabwriter"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	settings "github.com/lovek323/bclog/settings"
)

type PhpLogEvent struct {
	SyslogTime       time.Time
	LogLevel         string
	Content          string
	File             string
	Line             int
	StackTraceEvents []PhpStackTraceLogEvent
}

func (e *PhpLogEvent) AddStackTraceEvent(stackTraceEvent *PhpStackTraceLogEvent) {
	e.StackTraceEvents = append(e.StackTraceEvents, *stackTraceEvent)
}

func (e *PhpLogEvent) PrintLine(index int) {
	background := ct.None

	switch e.LogLevel {
	case "Notice":
	case "Warning":
		background = ct.None
		break
	case "Fatal error":
		background = ct.Red
		break
	case "Catchable fatal error":
		background = ct.Red
		break
	case "Parse error":
		background = ct.Red
		break
	case "SQL Error":
		background = ct.Red
		break
	case "Strict standards":
		background = ct.None
		break
	default:
		log.Fatalf(e.LogLevel)
	}

	fmt.Printf("[%d]  ", index)
	fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05") + "  ")
	ct.ChangeColor(ct.Yellow, false, background, false)
	fmt.Print("php  ")
	ct.ChangeColor(ct.Cyan, false, background, false)
	fmt.Printf("%s-%s-%d  ", e.LogLevel, e.File, e.Line)
	ct.ChangeColor(ct.None, false, background, false)
	fmt.Printf("%s\n", e.Content)
	ct.ResetColor()
}

func (e *PhpLogEvent) PrintFull() {
	fmt.Printf("\n---------- PHP LOG EVENT ----------\n")

	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprintf(
		writer,
		"SyslogTime:\t%s\n",
		e.SyslogTime.Format("2006-01-02 15:04:05"),
	)

	fmt.Fprintf(writer, "LogLevel:\t%s\n", e.LogLevel)
	fmt.Fprintf(writer, "Content:\t%s\n", e.Content)
	fmt.Fprintf(writer, "File:\t%s\n", e.File)
	fmt.Fprintf(writer, "Line:\t%d\n", e.Line)

	writer.Flush()

	ct.ChangeColor(ct.White, true, ct.None, false)
	fmt.Print("\nStack trace\n")
	ct.ResetColor()

	for _, phpStackTraceLogEvent := range e.StackTraceEvents {
		fmt.Fprintf(
			writer,
			"%d.\t%s\t%s\t%d\n",
			phpStackTraceLogEvent.Number,
			phpStackTraceLogEvent.Method,
			phpStackTraceLogEvent.File,
			phpStackTraceLogEvent.Line,
		)
	}

	writer.Flush()

	fmt.Printf("\n---------- PHP LOG EVENT ----------\n")
}

func (e *PhpLogEvent) Summary() string {
	return "php-" + e.LogLevel
}

func (e *PhpLogEvent) GetSyslogTime() time.Time {
	return e.SyslogTime
}

func (e *PhpLogEvent) Suppress(settings_ settings.SettingsInterface) bool {
	contentPatterns := settings_.GetPhpSuppressContentRegexes()

	for _, pattern := range contentPatterns {
		matched, err := regexp.MatchString(pattern, e.Content)

		if err != nil {
			log.Fatalf(
				"Error while matching ignore filters: %s (%s)\n",
				pattern,
				err,
			)
		}

		if matched {
			return true
		}
	}

	return false
}

type PhpStackTraceLogEvent struct {
	SyslogTime time.Time
	Number     int
	Method     string
	Parameters string
	File       string
	Line       int
}

func (e *PhpStackTraceLogEvent) GetSyslogTime() time.Time {
	return e.SyslogTime
}

func (e *PhpStackTraceLogEvent) PrintLine(index int) {
	fmt.Printf("[%d]  ", index)
	fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05") + "  ")
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Print("php-stack-trace  ")
	ct.ChangeColor(ct.Cyan, false, ct.None, false)
	fmt.Printf("%d-%s-%d  ", e.Number, e.File, e.Line)
	ct.ResetColor()
	fmt.Printf("%s\n", e.Method)
}

func (e *PhpStackTraceLogEvent) PrintFull() {
}

func (e *PhpStackTraceLogEvent) Summary() string {
	return "php-stack-trace"
}

func (e *PhpStackTraceLogEvent) Suppress(
	settings_ settings.SettingsInterface,
) bool {
	return settings_.GetPhpSuppressStackTraces()
}

func NewPhpLogEvent(
	syslogTime time.Time,
	source string,
	message string,
) LogEventInterface {
	re := regexp.MustCompile("^(?P<source>.*?): PHP Stack trace:")

	matches := re.FindStringSubmatch(message)

	if matches != nil {
		return &PhpStackTraceLogEvent{
			SyslogTime: syslogTime,
			Number:     0,
			Method:     "",
			File:       "",
			Line:       0,
		}
	}

	re = regexp.MustCompile(
		"^php: SQL Error \\(store_(?P<storeId>[0-9]{1,})\\): " +
			"(?P<content>.{1,}) in (?P<file>[^ ]{1,}) " +
			"on line (?P<line>[0-9]{1,})",
	)

	matches = re.FindStringSubmatch(message)

	if matches != nil {
		line, err := strconv.ParseInt(matches[4], 10, 32)

		if err != nil {
			log.Fatalf("Could not parse line: %s (%s)\n", matches[4], err)
		}

		return &PhpLogEvent{
			SyslogTime: syslogTime,
			LogLevel:   "SQL Error",
			Content:    matches[2] + " (store ID: " + matches[1] + ")",
			File:       matches[3],
			Line:       int(line),
		}
	}

	re = regexp.MustCompile(
		"^(?P<source>.*?): PHP (?P<level>.*?):  (?P<content>.{1,}) " +
			"in (?P<file>[^ ]{1,}) " +
			"on line (?P<line>[0-9]{1,})",
	)

	matches = re.FindStringSubmatch(message)

	if matches != nil {
		line, err := strconv.ParseInt(matches[5], 10, 32)

		if err != nil {
			log.Fatalf("Could not parse line: %s (%s)\n", matches[5], err)
		}

		logLevel := matches[2]
		content := matches[3]
		file := matches[4]

		event := PhpLogEvent{
			SyslogTime: syslogTime,
			LogLevel:   logLevel,
			Content:    content,
			File:       file,
			Line:       int(line),
		}

		re = regexp.MustCompile(
			"^Uncaught exception '(?P<type>.*?)' with message " +
				"'(?P<message>.*?)' in (?P<firstFile>[^ ]{1,}):" +
				"(?P<firstLine>[0-9]{1,})(?P<stackTrace>.*)",
		)
		matches = re.FindStringSubmatch(content)

		if matches != nil {
			event.Content = matches[1] + ": " + matches[2]

			re = regexp.MustCompile(
				"#(?P<number>[0-9]{1,})#(?P<index>[0-9]{1,}) " +
					"(?P<file>[^ ]{1,})\\((?P<line>[0-9]{1,})\\): (?P<details>.*?)",
			)

			stackTrace := matches[5]
			matches_ := re.FindAllStringSubmatch(stackTrace, -1)

			for _, match := range matches_ {
				// number := match[1]
				index, err := strconv.ParseInt(match[2], 10, 32)

				if err != nil {
					log.Fatalf(
						"Could not parse number: %s (%s)\n",
						matches[2],
						err,
					)
				}

				file := match[3]

				line, err := strconv.ParseInt(match[4], 10, 32)

				if err != nil {
					log.Fatalf(
						"Could not parse number: %s (%s)\n",
						matches[4],
						err,
					)
				}

				details := match[5]

				stackTraceEvent := &PhpStackTraceLogEvent{
					SyslogTime: event.SyslogTime,
					Number:     int(index),
					Method:     details,
					Parameters: "",
					File:       file,
					Line:       int(line),
				}

				event.AddStackTraceEvent(stackTraceEvent)
			}
		}

		return &event
	}

	// Search for a stack trace.
	re = regexp.MustCompile(
		"^(?P<source>.*?): PHP[ ]{1,}(?P<number>[0-9]{1,})\\. " +
			"(?P<method>.*)\\((?P<parameters>.{0,})\\) " +
			"(?P<file>[^ ]*):(?P<line>[0-9]{1,})$",
	)

	matches = re.FindStringSubmatch(message)

	if matches != nil {
		// source := matches[1]

		number, err := strconv.ParseInt(matches[2], 10, 32)

		if err != nil {
			log.Fatalf("Could not parse number: %s (%s)\n", matches[2], err)
		}

		method := matches[3]
		parameters := matches[4]
		file := matches[5]

		line, err := strconv.ParseInt(matches[6], 10, 32)

		if err != nil {
			log.Fatalf("Could not parse line: %s (%s)\n", matches[6], err)
		}

		return &PhpStackTraceLogEvent{
			SyslogTime: syslogTime,
			Number:     int(number),
			Method:     method,
			Parameters: parameters,
			File:       file,
			Line:       int(line),
		}
	}

	// Search for a stack trace with eval()'d code.
	re = regexp.MustCompile(
		"^(?P<source>.*?): PHP[ ]{1,}(?P<number>[0-9]{1,})\\. " +
			"(?P<method>[^ ]*) (?P<file>[^ ]*)\\((?P<line>[0-9]{1,})\\) " +
			": eval\\(\\)'d code:(?P<evalLine>[0-9]{1,})$",
	)

	matches = re.FindStringSubmatch(message)

	if matches != nil {
		number, err := strconv.ParseInt(matches[2], 10, 32)

		if err != nil {
			log.Fatalf("Could not parse number: %s (%s)\n", matches[2], err)
		}

		line, err := strconv.ParseInt(matches[5], 10, 32)

		if err != nil {
			log.Fatalf("Could not parse line: %s (%s)\n", matches[5], err)
		}

		return &PhpStackTraceLogEvent{
			SyslogTime: syslogTime,
			Number:     int(number),
			Method:     matches[3],
			File:       matches[4],
			Line:       int(line),
		}
	}

	return nil
}
