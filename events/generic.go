package events

import (
	"fmt"
	"regexp"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	settings "github.com/lovek323/bclog/settings"
)

type GenericLogEvent struct {
	SyslogTime time.Time
	Name       string
	Content    string
}

func (e *GenericLogEvent) PrintLine(index int) {
	fmt.Printf("[%d]  ", index)
	fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05") + "  ")
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Print("generic  ")
	ct.ChangeColor(ct.Cyan, false, ct.None, false)
	fmt.Printf("%s  ", e.Name)
	ct.ResetColor()
	fmt.Printf("%s\n", e.Content)
}

func (e *GenericLogEvent) PrintFull() {
}

func (e *GenericLogEvent) Summary() string {
	return "generic-" + e.Name
}

func (e *GenericLogEvent) Suppress(settings_ settings.SettingsInterface) bool {
	for _, name := range settings_.GetGenericSuppressNames() {
		if e.Name == name {
			return true
		}
	}

	return false
}

func (e *GenericLogEvent) GetSyslogTime() time.Time {
	return e.SyslogTime
}

func NewGenericLogEvent(
	syslogTime time.Time,
	source string,
	message string,
) LogEventInterface {
	re := regexp.MustCompile(
		"^(?P<name>.*?): (?P<message>.*)$",
	)

	matches := re.FindStringSubmatch(message)

	if matches == nil {
		return nil
	}

	name := matches[1]
	content := matches[2]

	return &GenericLogEvent{
		SyslogTime: syslogTime,
		Name:       name,
		Content:    content,
	}
}
