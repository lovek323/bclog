package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "time"

    ct        "github.com/daviddengcn/go-colortext"
    events    "github.com/lovek323/bclog/events"
    linenoise "github.com/GeertJohan/go.linenoise"
)

type LogEventInterface interface {
    Println(int)
    PrintFull()
}

type ProcessLogEvent struct {
    SyslogTime time.Time
    Name       string
    ProcessId  int
    Content    string
}

func (e *ProcessLogEvent) Println(index int) {
    // TODO: Create a CronLogEvent and handle that separately (looking for
    // errors)
    /* Ignore cron and postfix log messages manually. */
    if (e.Name == "/USR/SBIN/CRON"  ||
        e.Name == "postfix/pickup"  ||
        e.Name == "postfix/cleanup" ||
        e.Name == "postfix/qmgr"    ||
        e.Name == "postfix/discard" ||
        e.Name == "postfix/smtp") {
        return
    }

    fmt.Printf("[%d]  ", index)
    fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05")+"  ")
    ct.ChangeColor(ct.Yellow, false, ct.None, false)
    fmt.Print("process  ")
    ct.ChangeColor(ct.Cyan, false, ct.None, false)
    fmt.Printf("%s-%d  ", e.Name, e.ProcessId)
    ct.ResetColor()
    fmt.Printf("%s\n", e.Content)
}

func (e *ProcessLogEvent) PrintFull() {
}

var history []LogEventInterface

func main() {
    go func () {
        for {
            line, err := linenoise.Line("")

            if err != nil {
                if err == linenoise.KillSignalError {
                    quit()
                } else {
                    log.Fatalf("Could not read line: %s\n", err)
                }
            }

            if len(line) > 0 {
                if (line == "quit") {
                    quit()
                } else {
                    index, err := strconv.ParseInt(line, 10, 32)

                    if err == nil {
                        event := history[index]
                        event.PrintFull()
                    } else {
                        fmt.Printf("Received %s\n", line)
                    }
                }
            }
        }
    }()

    readLog()
}

func quit() {
    os.Exit(0)
}

func readLog() {
    command := exec.Command(
        "ssh",
        "vagrant@localhost",
        "-p2222",
        "-i",
        "/Users/jason.oconal/.vagrant.d/insecure_private_key",
        "--",
        "sudo tail -n 500 -f /var/log/syslog",
    )

    stdout, err := command.StdoutPipe()

    if err != nil {
        log.Printf("Could not create log tail command: %s\n", err)
    }

    if err = command.Start(); err != nil {
        log.Printf("Could not start log tail command: %s\n", err)
    }

    reader  := bufio.NewReader(stdout)

    for {
        line, err := reader.ReadString('\n')

        if err != nil {
            if err != io.EOF {
                log.Printf("Could not read from log tail command stdout: %s", err)
            } else {
                log.Printf("EOF")
            }
            break
        }

        fmt.Print("\r")

        event := getEvent(line)

        if event == nil {
            log.Printf("Could not parse: %s", line)
        } else {
            event.Println(len(history))
            history = append(history, event)
        }
    }

    command.Wait()
}

func getEvent(text string) LogEventInterface {
    re := regexp.MustCompile(
        "^(?P<date>(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Oct|Nov|Dec) "+
        "(?:[0-9]{1,}) [0-9]{2}:[0-9]{2}:[0-9]{2}) "+
        "(?P<source>.*?) "+
        //"(?P<process>[A-Za-z])\\[(?P<pid>[0-9]{1,})\\]: "+
        "(?P<message>.*)\n$",
    )

    matches := re.FindStringSubmatch(text)

    if matches == nil {
        return nil
    }

    syslogTime, err := time.Parse("Jan 2 15:04:05", matches[1])

    if err != nil {
        log.Fatalf("Failed to parse %s (%s)", matches[1], err)
    }

    source  := matches[2]
    message := matches[3]

    var event LogEventInterface

    event = events.NewNginxLogEvent(syslogTime, source, message)

    if event != (LogEventInterface)(nil) {
        return event
    }

    event = getProcessEvent(syslogTime, source, message)

    if event != (LogEventInterface)(nil) {
        return event
    }

    event = events.NewPhpLogEvent(syslogTime, source, message)

    if event != (LogEventInterface)(nil) {
        return event
    }

    return nil
}

func getProcessEvent(syslogTime time.Time, source string, message string) LogEventInterface {
    re := regexp.MustCompile(
        "^(?P<name>.*?)\\[(?P<processId>[0-9]{1,})\\]: (?P<message>.*)$",
    )

    matches := re.FindStringSubmatch(message)

    if matches == nil {
        return nil
    }

    name := matches[1]

    processId, err := strconv.ParseInt(matches[2], 10, 32)

    if err != nil {
        log.Fatalf("Could not parse process ID: %s (%s)\n", matches[2], err)
    }

    content := matches[3]

    switch (name) {
    case "bigcommerce_app":
        event := events.NewBigcommerceAppLogEvent(
            syslogTime,
            source,
            int(processId),
            content,
        )

        // This could be a PHP error as well.
        if event != (*events.BigcommerceAppLogEvent)(nil) {
            return event
        }

        return nil
    }

    return &ProcessLogEvent{
        SyslogTime: syslogTime,
        Name:       name,
        ProcessId:  int(processId),
        Content:    content,
    }
}
