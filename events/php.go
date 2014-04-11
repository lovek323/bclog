package events

import (
    "fmt"
    "log"
    "regexp"
    "strconv"
    "time"

    ct       "github.com/daviddengcn/go-colortext"
    settings "github.com/lovek323/bclog/settings"
)

type PhpLogEvent struct {
    SyslogTime time.Time
    LogLevel   string
    Content    string
    File       string
    Line       int
}

func (e *PhpLogEvent) PrintLine(index int) {
    background := ct.None

    switch (e.LogLevel) {
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
    case "SQL Error":
        background = ct.Red
        break
    default:
        log.Fatalf(e.LogLevel)
    }

    fmt.Printf("[%d]  ", index)
    fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05")+"  ")
    ct.ChangeColor(ct.Yellow, false, background, false)
    fmt.Print("php  ")
    ct.ChangeColor(ct.Cyan, false, background, false)
    fmt.Printf("%s-%s-%d  ", e.LogLevel, e.File, e.Line)
    ct.ChangeColor(ct.None, false, background, false)
    fmt.Printf("%s\n", e.Content)
    ct.ResetColor()
}

func (e *PhpLogEvent) PrintFull() {
}

func (e *PhpLogEvent) Summary() string {
    return "php-"+e.LogLevel
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
            return true;
        }
    }

    return false
}

type PhpStackTraceEvent struct {
    SyslogTime time.Time
    Number     int
    Method     string
    File       string
    Line       int
}

func (e *PhpStackTraceEvent) PrintLine(index int) {
    // Don't show stack traces.
    return;

    fmt.Printf("[%d]  ", index)
    fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05")+"  ")
    ct.ChangeColor(ct.Yellow, false, ct.None, false)
    fmt.Print("php-stack-trace  ")
    ct.ChangeColor(ct.Cyan, false, ct.None, false)
    fmt.Printf("%d-%s-%d  ", e.Number, e.File, e.Line)
    ct.ResetColor()
    fmt.Printf("%s\n", e.Method)
}

func (e *PhpStackTraceEvent) PrintFull() {
}

func (e *PhpStackTraceEvent) Summary() string {
    return "php-stack-trace"
}

func (e *PhpStackTraceEvent) Suppress(
    settings_ settings.SettingsInterface,
) bool {
    return settings_.GetPhpSuppressStackTraces()
}

func (e *PhpStackTraceEvent) GetSyslogTime() time.Time {
    return e.SyslogTime
}

func NewPhpLogEvent(
    syslogTime time.Time,
    source string,
    message string,
) LogEventInterface {
    re := regexp.MustCompile("^(?P<source>.*?): PHP Stack trace:")

    matches := re.FindStringSubmatch(message)

    if matches != nil {
        return &PhpStackTraceEvent{
            SyslogTime: syslogTime,
            Number:     0,
            Method:     "",
            File:       "",
            Line:       0,
        }
    }

    re = regexp.MustCompile(
        "^php: SQL Error \\(store_(?P<storeId>[0-9]{1,})\\): "+
        "(?P<content>.{1,}) in (?P<file>[^ ]{1,}) "+
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
            Content:    matches[2]+ " (store ID: "+matches[1]+")",
            File:       matches[3],
            Line:       int(line),
        }
    }

    re = regexp.MustCompile(
        "^(?P<source>.*?): PHP (?P<level>.*?):  (?P<content>.{1,}) "+
        "in (?P<file>[^ ]{1,}) "+
        "on line (?P<line>[0-9]{1,})",
    )

    matches = re.FindStringSubmatch(message)

    if matches != nil {
        line, err := strconv.ParseInt(matches[5], 10, 32)

        if err != nil {
            log.Fatalf("Could not parse line: %s (%s)\n", matches[5], err)
        }

        return &PhpLogEvent{
            SyslogTime: syslogTime,
            LogLevel:   matches[2],
            Content:    matches[3],
            File:       matches[4],
            Line:       int(line),
        }
    }

    // Search for a stack trace.
    re = regexp.MustCompile(
        "^(?P<source>.*?): PHP[ ]{1,}(?P<number>[0-9]{1,})\\. "+
        "(?P<method>[^ ]*) (?P<file>[^ ]*):(?P<line>[0-9]{1,})$",
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

        return &PhpStackTraceEvent{
            SyslogTime: syslogTime,
            Number:     int(number),
            Method:     matches[3],
            File:       matches[4],
            Line:       int(line),
        }
    }

    // Search for a stack trace with eval()'d code.
    re = regexp.MustCompile(
        "^(?P<source>.*?): PHP[ ]{1,}(?P<number>[0-9]{1,})\\. "+
        "(?P<method>[^ ]*) (?P<file>[^ ]*)\\((?P<line>[0-9]{1,})\\) "+
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

        return &PhpStackTraceEvent{
            SyslogTime: syslogTime,
            Number:     int(number),
            Method:     matches[3],
            File:       matches[4],
            Line:       int(line),
        }
    }

    fmt.Printf("\rNot a PHP log message: %s\n", message)

    return nil
}
