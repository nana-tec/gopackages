package ntlogger

type Category string
type SubCategory string
type ExtraKey string

const (
	AppName      ExtraKey = "AppName"
	LoggerName   ExtraKey = "Logger"
	ClientIp     ExtraKey = "ClientIp"
	HostIp       ExtraKey = "HostIp"
	Method       ExtraKey = "Method"
	StatusCode   ExtraKey = "StatusCode"
	BodySize     ExtraKey = "BodySize"
	Path         ExtraKey = "Path"
	Latency      ExtraKey = "Latency"
	RequestBody  ExtraKey = "RequestBody"
	ResponseBody ExtraKey = "ResponseBody"
	ErrorMessage ExtraKey = "ErrorMessage"
)

type LogConfig struct {
	FilePath            string `mapstructure:"LOG_FILE_PATH"`
	Encoding            string `mapstructure:"LOG_ENCODING"`
	Level               string `mapstructure:"LOG_LEVEL"`
	TelemetryEnabled    string `mapstructure:"TELEMETRY_ENABLED"`
	TelemetryEndpoint   string `mapstructure:"TELEMETRY_ENDPOINT"`
	TelemetryProjectDsn string `mapstructure:"TELEMETRY_PROJECT_DSN"`
	TelemetryIsSecured  string `mapstructure:"TELEMETRY_IS_SECURED"`
	AppName             string `mapstructure:"APP_NAME"`
	AppServiceName      string `mapstructure:"APP_SERVICE_NAME"`
	AppNameSpace        string `mapstructure:"APP_NAMESAPCE"`
	AppVersion          string `mapstructure:"APP_VERSION"`
	Environment         string `mapstructure:"ENVIRONMENT"`
}
