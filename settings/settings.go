package settings

type Settings struct {
    BigcommerceApp struct {
        SuppressLogLevels []string
    }

    NginxAccess struct {
        SuppressStatusCodes []int
    }

    Process struct {
        SuppressNames []string
    }

    Php struct {
        SuppressStackTraces    bool
        SuppressContentRegexes []string
    }
}

type SettingsInterface interface {
   GetBigcommerceAppSuppressLogLevels() []string
   GetNginxSuppressStatusCodes()        []int
   GetPhpSuppressStackTraces()          bool
   GetPhpSuppressContentRegexes()       []string
   GetProcessSuppressNames()            []string
}

func (s *Settings) GetBigcommerceAppSuppressLogLevels() []string {
    return s.BigcommerceApp.SuppressLogLevels
}

func (s *Settings) GetNginxSuppressStatusCodes() []int {
    return s.NginxAccess.SuppressStatusCodes
}

func (s *Settings) GetPhpSuppressStackTraces() bool {
    return s.Php.SuppressStackTraces
}

func (s *Settings) GetPhpSuppressContentRegexes() []string {
    return s.Php.SuppressContentRegexes
}

func (s *Settings) GetProcessSuppressNames() []string {
    return s.Process.SuppressNames
}
