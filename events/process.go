package events

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	settings "github.com/lovek323/bclog/settings"
)

type ProcessLogEvent struct {
	SyslogTime time.Time
	Name       string
	ProcessId  int
	Content    string
}

func (e *ProcessLogEvent) PrintLine(index int) {
	fmt.Printf("[%d]  ", index)
	fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05") + "  ")
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Print("process  ")
	ct.ChangeColor(ct.Cyan, false, ct.None, false)
	fmt.Printf("%s-%d  ", e.Name, e.ProcessId)
	ct.ResetColor()
	fmt.Printf("%s\n", e.Content)
}

func (e *ProcessLogEvent) PrintFull() {
}

func (e *ProcessLogEvent) Summary() string {
	return "process"
}

func (e *ProcessLogEvent) Suppress(settings_ settings.SettingsInterface) bool {
	for _, name := range settings_.GetProcessSuppressNames() {
		if e.Name == name {
			return true
		}
		matched, _ := regexp.MatchString(name, e.Name)
		if matched {
			return true
		}
	}

	return false
}

func (e *ProcessLogEvent) GetSyslogTime() time.Time {
	return e.SyslogTime
}

func NewProcessLogEvent(
	syslogTime time.Time,
	source string,
	message string,
) LogEventInterface {
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

	switch name {
	case "bigcommerce_app", "ool bigcommerce_app":
		event := NewBigcommerceAppLogEvent(
			syslogTime,
			source,
			int(processId),
			content,
		)

		// This could be a PHP error as well.
		if event != (*BigcommerceAppLogEvent)(nil) {
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
