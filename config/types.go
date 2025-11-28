package config

type server struct {
	Secret  []byte
	Version string
	Name    string
}

type snowflake struct {
	WorkerID      int64 `mapstructure:"worker-id"`
	DatancenterID int64 `mapstructure:"datancenter-id"`
}

type service struct {
	Name     string
	AddrList []string
	LB       bool `mapstructure:"load-balance"`
}

type mySQL struct {
	Addr     string
	Database string
	Username string
	Password string
	Charset  string
}

type etcd struct {
	Addr string
}

type kafka struct {
	Brokers         []string
	Topic           string
	GroupID         string `mapstructure:"group-id"`
	DeadLetterTopic string `mapstructure:"dead-letter-topic"`
}

type redis struct {
	Addr     string
	Password string
}

type oss struct {
	Endpoint        string
	AccessKeyID     string `mapstructure:"accessKey-id"`
	AccessKeySecret string `mapstructure:"accessKey-secret"`
	BucketName      string
	MainDirectory   string `mapstructure:"main-directory"`
}

type config struct {
	Server    server
	Snowflake snowflake
	MySQL     mySQL
	Etcd      etcd
	Kafka     kafka
	Redis     redis
	OSS       oss
}
