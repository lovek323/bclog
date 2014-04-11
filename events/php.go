package events

import (
    "fmt"
    "log"
    "regexp"
    "strconv"
    "time"

    ct "github.com/daviddengcn/go-colortext"
)

type PhpLogEventInterface interface {
    Println(int)
    PrintFull()
}

type PhpLogEvent struct {
    SyslogTime time.Time
    LogLevel   string
    Content    string
    File       string
    Line       int
}

func (e *PhpLogEvent) Println(index int) {
    ignoreFilters := []string{
        "^Failed to write to Twig cache.*$",
        "^Undefined index: MBALoginToken$",
    }

    for _, ignoreFilter := range ignoreFilters {
        matched, err := regexp.MatchString(ignoreFilter, e.Content)

        if err != nil {
            log.Fatalf(
                "Error while matching ignore filters: %s (%s)\n",
                ignoreFilter,
                err,
            )
        }

        if matched {
            return;
        }
    }

    background := ct.None

    switch (e.LogLevel) {
    case "Notice":
    case "Warning":
        background = ct.None
        break
    case "Fatal error":
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

type PhpStackTraceEvent struct {
    SyslogTime time.Time
    Number     int
    Method     string
    File       string
    Line       int
}

func (e *PhpStackTraceEvent) Println(index int) {
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

func NewPhpLogEvent(
    syslogTime time.Time,
    source string,
    message string,
) PhpLogEventInterface {
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
        "^(?P<source>.*?): PHP (?P<level>.*?):  (?P<content>.{1,}) in (?P<file>[^ ]{1,}) "+
        "on line (?P<line>[0-9]{1,})$",
    )

    matches = re.FindStringSubmatch(message)

    if matches != nil {
        line, err := strconv.ParseInt(matches[5], 10, 32)

        if err != nil {
            log.Fatalf("Could not parse line: %s (%s)\n", matches[4], err)
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

    return nil
}
