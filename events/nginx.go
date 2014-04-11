package events

import (
    "fmt"
    "log"
    "regexp"
    "strconv"
    "time"

    ct "github.com/daviddengcn/go-colortext"
)

type NginxLogEventInterface interface {
    Println()
}

type NginxAccessLogEvent struct {
    SyslogTime time.Time
    Source     string
    Hostname   string
    IpAddress  string
    Time       time.Time
    Request    NginxLogEventRequest
}

type NginxLogEventRequest struct {
    Method          string
    Uri             string
    ProtocolVersion string
    StatusCode      int
    ContentLength   int
}

func (e *NginxAccessLogEvent) Println() {
    // Ignore 200-399 events manually for now
    if e.Request.StatusCode >= 200 && e.Request.StatusCode < 400 {
        return
    }

    background := ct.None
    bold       := false

    if e.Request.StatusCode >= 500 {
        background = ct.Red
        bold       = false
    } else {
        background = ct.None
        bold       = true
    }

    fmt.Print(e.Time.Format("2006-01-02 15:04:05")+"  ")
    ct.ChangeColor(ct.Yellow, bold, background, false)
    fmt.Print("nginx-access  ")
    ct.ChangeColor(ct.Cyan, bold, background, false)
    fmt.Printf("%s-%d  ", e.Request.Method, e.Request.StatusCode)
    fmt.Printf("%s\n", e.Request.Uri)
    ct.ResetColor()
}

type NginxErrorLogEvent struct {
    SyslogTime time.Time
    LogLevel   string
    Content    string
    Client     string
    Server     string
    Request    NginxLogEventRequest
    Host       string
    Referrer   string
}

func (e *NginxErrorLogEvent) Println() {
    if e.LogLevel == "error" {
        fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05")+"  ")
        ct.ChangeColor(ct.Yellow, false, ct.Red, false)
        fmt.Print("nginx-error  ")
        ct.ChangeColor(ct.Cyan, false, ct.Red, false)
        fmt.Printf("%s  ", e.LogLevel)
        ct.ChangeColor(ct.None, false, ct.Red, false)
        fmt.Printf("%s %s\n", e.Request.Uri, e.Content)
        ct.ResetColor()
    } else {
        fmt.Print(e.SyslogTime.Format("2006-01-02 15:04:05")+"  ")
        ct.ChangeColor(ct.Yellow, false, ct.None, false)
        fmt.Print("nginx-error  ")
        ct.ChangeColor(ct.Cyan, false, ct.None, false)
        fmt.Printf("%s  ", e.LogLevel)
        ct.ResetColor()
        fmt.Printf("%s %s\n", e.Request.Uri, e.Content)
    }
}

func NewNginxLogEvent(syslogTime time.Time, source string, message string) NginxLogEventInterface {
    re := regexp.MustCompile(
        "^nginx: (?P<hostname>.*?) (?P<ipAddress>[0-9\\.]*) (?:.*?) (?:.*?) "+
        "\\[(?P<time>[0-9]{2}/(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)/[0-9]{4}:[0-9]{2}:[0-9]{2}:[0-9]{2} \\+[0-9]{4})\\]  "+
        "\"(?P<method>.*?) (?P<uri>.*?) (?P<protocolVersion>.*?)\" "+
        "(?P<statusCode>[0-9]{1,}) (?P<contentLength>[0-9]{1,}) (?:.*?) "+
        "(?:.*?) (?:.*?)$",
    )

    matches := re.FindStringSubmatch(message)

    if matches == nil {
        // Search for other log messages.
        re = regexp.MustCompile(
            "^nginx:  \\[(?P<level>.*?)\\] (?P<content>.*?), "+
            "client: (?P<clientIp>.*), server: (?P<server>.*), "+
            "request: \"(?P<method>.*?) (?P<uri>.*?) "+
            "(?P<protocolVersion>.*?)\", host: (?P<host>.*?), "+
            "referrer: (?P<referrer>.*?)$",
        )

        matches = re.FindStringSubmatch(message)

        if matches == nil {
            return nil
        }

        request := NginxLogEventRequest{
            Method:          matches[5],
            Uri:             matches[6],
            ProtocolVersion: matches[7],
            StatusCode:      0,
            ContentLength:   0,
        }

        return &NginxErrorLogEvent{
            SyslogTime: syslogTime,
            LogLevel:   matches[1],
            Content:    matches[2],
            Client:     matches[3],
            Server:     matches[4],
            Request:    request,
            Host:       matches[8],
            Referrer:   matches[9],
        }
    }

    time_, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[3])

    if err != nil {
        log.Fatalf("Could not parse time: %s (%s)\n", matches[3])
    }

    statusCode, err := strconv.ParseInt(matches[7], 10, 32)

    if err != nil {
        log.Fatalf("Could not parse status code: %s (%s)\n", matches[7], err)
    }

    contentLength, err := strconv.ParseInt(matches[8], 10, 32)

    if err != nil {
        log.Fatalf(
            "Could not parse content length: %s (%s)\n",
            matches[8],
            err,
        )
    }

    request := NginxLogEventRequest{
        Method:          matches[4],
        Uri:             matches[5],
        ProtocolVersion: matches[6],
        StatusCode:      int(statusCode),
        ContentLength:   int(contentLength),
    }

    return &NginxAccessLogEvent{
        SyslogTime: syslogTime,
        Source:     source,
        Hostname:   matches[1],
        IpAddress:  matches[2],
        Time:       time_,
        Request:    request,
    }
}
