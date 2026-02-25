package def

const (
    Dsn         = `%s:%s@tcp(%s:%d)/%s?timeout=10s&parseTime=true&charset=%s`
    HostPattern = `^(.*)\:(.*)\@(.*)\:(\d+)$`
    DbPattern   = `^([A-Za-z0-9_]+)$`
)
