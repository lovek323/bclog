package events

import (
    "encoding/json"
    "fmt"
    "log"
    "regexp"
    "time"

    ct "github.com/daviddengcn/go-colortext"
)

type BigcommerceAppLogEvent struct {
    SyslogTime   time.Time
    ProcessId    int
    LogLevel     string
    Content      string
    Args         string
    StoreContext BigcommerceAppStoreContext
}

type BigcommerceAppStoreContext struct {
    StoreId   int    `json:"store_id,string"`
    StoreHash string `json:"store_hash"`
    Domain    string
}

func (e *BigcommerceAppLogEvent) Println() {
    if (e.LogLevel == "DEBUG") {
        return;
    }

    fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05")+"  ")
    ct.ChangeColor(ct.Yellow, false, ct.None, false)
    fmt.Print("bigcommerce-app  ")
    ct.ChangeColor(ct.Cyan, false, ct.None, false)
    fmt.Printf("%s-%d  ", e.LogLevel, e.StoreContext.StoreId)
    ct.ResetColor()
    fmt.Printf("%s\n", e.Content)
}

func NewBigcommerceAppLogEvent(
    syslogTime time.Time,
    source string,
    processId int,
    message string,
) *BigcommerceAppLogEvent {
    re := regexp.MustCompile(
        "^BigcommerceApp\\.(?P<logLevel>.*?): (?P<content>.*?) "+
        "(?P<args>\\[\\]|\\{.*?\\}) (?P<storeContext>\\{.*?\\})$",
    )

    matches := re.FindStringSubmatch(message)

    if matches == nil {
        return nil
    }

    logLevel         := matches[1]
    content          := matches[2]
    args             := matches[3]
    storeContextJson := matches[4]

    var storeContext BigcommerceAppStoreContext

    err := json.Unmarshal([]byte(storeContextJson), &storeContext)

    if err != nil {
        log.Fatalf(
            "Could not parse store context: %s (%s)\n",
            storeContextJson,
            err,
        )
    }

    return &BigcommerceAppLogEvent{
        SyslogTime:   syslogTime,
        ProcessId:    processId,
        LogLevel:     logLevel,
        Content:      content,
        Args:         args,
        StoreContext: storeContext,
    }
}
