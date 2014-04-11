package events

import (
   "time"

   settings "github.com/lovek323/bclog/settings"
)

type LogEventInterface interface {
    PrintLine(int)
    PrintFull()

    GetSyslogTime()                      time.Time
    Summary()                            string
    Suppress(settings.SettingsInterface) bool
}
