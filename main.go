package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "os/user"
    "regexp"
    "strconv"
    "strings"
    "text/tabwriter"
    "time"

    ct       "github.com/daviddengcn/go-colortext"
    events    "github.com/lovek323/bclog/events"
    settings  "github.com/lovek323/bclog/settings"
    linenoise "github.com/GeertJohan/go.linenoise"
)

var history    []events.LogEventInterface
var statistics map[string][]events.LogEventInterface
var settings_  settings.Settings
var lastPrompt time.Time

func main() {
    statistics = make(map[string][]events.LogEventInterface)

    loadConfig()

    linenoise.SetCompletionHandler(func (in string) []string {
        availableCommands := []string{"clear", "help", "reload", "show", "quit", "summary"}
        matchedCommands   := []string{}

        for _, command := range availableCommands {
            if len(in) <= len(command) && strings.Index(command, in) == 0 {
                matchedCommands = append(matchedCommands, command)
            } else if len(in) > len("show") && in[0:len("show")] == "show" {
                for summary, _ := range statistics {
                    if len(in[len("show "):]) <= len(summary) && strings.Index(summary, in[len("show "):]) == 0 {
                        matchedCommands = append(matchedCommands, "show "+summary)
                    }
                }
            }
        }

        return matchedCommands
    })

    go func () {
        for {
            lastPrompt = time.Now()

            line, err := linenoise.Line("> ")

            if err != nil {
                if err == linenoise.KillSignalError {
                    quit()
                } else {
                    log.Fatalf("Could not read line: %s\n", err)
                }
            }

            err = linenoise.AddHistory(line)

            if err != nil {
                log.Printf("Failed to add %s to history (%s)\n", line, err)
            }

            args := strings.Split(line, " ")

            if len(args) == 0 {
                continue
            }

            switch (args[0]) {
            case "": summary([]string{"last-prompt"}); break
            case "clear": linenoise.Clear(); break
            case "quit": quit(); break
            case "reload": loadConfig(); break
            case "show": show(args[1:]); break
            case "summary": summary(args[1:]); break

            default:
                index, err := strconv.ParseInt(line, 10, 32)

                if err == nil {
                    event := history[index]
                    event.PrintFull()
                } else {
                    fmt.Printf("Unrecognised command: %s\n\n", line)
                }
            }
        }
    }()

    readLog()
}

func loadConfig() {
    user, err := user.Current()
    if err != nil {
        log.Fatalf("Could not determine home directory: %s\n", err)
    }
    configJson, err := ioutil.ReadFile(fmt.Sprintf("%s/.config/bclog/config.json", user.HomeDir))

    if err != nil {
        log.Fatalf("Error reading ~/.config/bclog/config.json: %s", err)
    }
    err = json.Unmarshal(configJson, &settings_)

    if err != nil {
        log.Fatalf("Error reading ~/.config/bclog/config.json: %s", err)
    }
    fmt.Println("Loaded config\n")
}

func quit() {
    os.Exit(0)
}

func summary(args []string) {
    var duration time.Duration
    var err error

    if len(args) > 0 {
        if args[0] == "last-prompt" {
            duration = time.Now().Sub(lastPrompt)
        } else {
            duration, err = time.ParseDuration(args[0])

            if err != nil {
                fmt.Println(
                    "Invalid syntax: first argument to summary must be a valid "+
                    "duration or empty",
                );
                fmt.Println("summary [duration]\n");
            }
        }
    } else {
        duration, _ = time.ParseDuration("24h")
    }

    ct.ChangeColor(ct.Yellow, true, ct.None, false)
    fmt.Printf("\nSUMMARY (LAST %s)\n", duration)
    ct.ResetColor()

    writer := new(tabwriter.Writer)
    writer.Init(os.Stdout, 0, 8, 2, ' ', 0)

    for summary, events := range statistics {
        lastDuration := history[len(history)-1].GetSyslogTime().Sub(
            events[len(events)-1].GetSyslogTime(),
        )

        count := 0

        for _, event := range events {
            if event == nil {
                continue
            }
            if history[len(history)-1].GetSyslogTime().Sub(event.GetSyslogTime()) <= duration {
                count++
            }
        }

        if count == 0 {
            continue
        }

        fmt.Fprintf(
            writer,
            "%s\t%d event(s)\tLast %s ago\n",
            summary,
            count,
            lastDuration,
        )
    }

    writer.Flush()

    fmt.Print("\n")
}

func show(args []string) {
    if len(args) < 2 {
        fmt.Println("Invalid syntax: show requires two arguments")
        fmt.Println("show <type> <duration>\n")

        return
    }

    summary       := args[0]
    duration, err := time.ParseDuration(args[1])

    if err != nil {
        fmt.Println(
            "Invalid syntax: second argument to show must be a valid duration",
        );
        fmt.Println("show <type> <duration>\n");
    }

    fmt.Println("\n---------- SHOW ----------");
    fmt.Printf("Showing %s events from the last %s\n", summary, duration)

    for index, event := range history {
        if (summary == "*" || event.Summary() == summary) &&
            history[len(history)-1].GetSyslogTime().Sub(event.GetSyslogTime()) <= duration {
            event.PrintLine(index)
        }
    }
    fmt.Println("--------------------------\n");
}

func readLog() {
    command := exec.Command(
        "ssh",
        "vagrant@localhost",
        "-p2200",
        "-i",
        settings_.PrimaryKeyFile,
        "--",
        "sudo tail -n 10000 -f /var/log/syslog",
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
                log.Printf("Could not read from log tail command stdout: %s\n", err)
            } else {
                log.Printf("EOF\n")
            }
            break
        }

        event := getEvent(line)

        if event == nil {
            log.Printf("\rCould not parse: %s", line)
            fmt.Print("\r> ")
        } else {
            if (!event.Suppress(&settings_)) {
                fmt.Print("\r")
                event.PrintLine(len(history))
                fmt.Print("\r> ")
            }

            history = append(history, event)
            summary := event.Summary()

            if _, exists := statistics[summary]; !exists {
                statistics[summary] = make([]events.LogEventInterface, 1)
            }

            statistics[summary] = append(statistics[summary], event)
        }
    }

    command.Wait()
}

func getEvent(text string) events.LogEventInterface {
    re := regexp.MustCompile(
        "^(?P<date>(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Oct|Nov|Dec)(?:[ ]{1,})"+
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

    var event events.LogEventInterface

    event = events.NewNginxLogEvent(syslogTime, source, message)

    if event != (events.LogEventInterface)(nil) {
        return event
    }

    event = events.NewProcessLogEvent(syslogTime, source, message)

    if event != (events.LogEventInterface)(nil) {
        return event
    }

    event = events.NewPhpLogEvent(syslogTime, source, message)

    if phpStackTraceLogEvent, ok := event.(*events.PhpStackTraceLogEvent); ok {
        for i := len(history)-1; i >= 0; i-- {
            if phpLogEvent, ok := (history[i]).(*events.PhpLogEvent); ok {
                phpLogEvent.AddStackTraceEvent(phpStackTraceLogEvent)
                break
            }
        }
    }

    if event != (events.LogEventInterface)(nil) {
        return event
    }

    event = events.NewGenericLogEvent(syslogTime, source, message)

    if event != (events.LogEventInterface)(nil) {
        return event
    }

    return nil
}
