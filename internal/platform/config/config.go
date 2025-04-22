package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"

	"hippo/internal/platform/consts"
)

type Config struct {
	Env        string          `mapstructure:"env" validate:"required,oneof=local dev prod"`
	App        App             `mapstructure:"app" validate:"required"`
	HttpServer HttpServer      `mapstructure:"http_server" validate:"required"`
	GrpcAudit  GrpcAuditClient `mapstructure:"grpc_audit_client" validate:"required"`
	DBConn     DBConn          `mapstructure:"db_conn" validate:"required"`
}

type App struct {
	HandlerTimeout   time.Duration `mapstructure:"handler_timeout" validate:"required,gt=0"`
	RefreshTokenLife time.Duration `mapstructure:"refresh_token_life" validate:"required,gt=0"`
	AccessTokenLife  time.Duration `mapstructure:"access_token_life" validate:"required,gt=0"`
}

type HttpServer struct {
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,gt=0"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,gt=0"`
	Idle         time.Duration `mapstructure:"idle_timeout" validate:"required,gt=0"`
}

type GrpcAuditClient struct {
	Host         string        `mapstructure:"host" validate:"required,ip4_addr"`
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Timeout      time.Duration `mapstructure:"timeout" validate:"required,gt=0"`
	CertFilePath string        `mapstructure:"cert_path" validate:"file_if_provided"`
}

type DBConn struct {
	Host            string        `mapstructure:"host" validate:"required"`
	Port            int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Name            string        `mapstructure:"name" validate:"required"`
	User            string        `mapstructure:"user" validate:"required"`
	Password        string        `mapstructure:"password" validate:"required"`
	SSLMode         string        `mapstructure:"sslmode" validate:"oneof=disable require verify-ca verify-full"`
	SSLRootCert     string        `mapstructure:"sslrootcert,omitempty" validate:"file_if_provided"`
	MaxOpenConns    int           `mapstructure:"max_open_conns,omitempty"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns,omitempty"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime,omitempty"`
}

const (
	defaultConfigDir  = "configs"
	defaultConfigName = "config"
)

func NewConfig(configDir, configName string) (*Config, error) {
	v := viper.New()

	if configDir == "" {
		configDir = defaultConfigDir
	}
	if configName == "" {
		configName = defaultConfigName
	}

	setDefaults(v)

	v.AddConfigPath(configDir)
	v.SetConfigName(configName)
	v.SetConfigType(consts.CfgType)

	v.SetEnvPrefix(consts.EnvVarPrefix)
	v.AutomaticEnv()
	v.BindEnv("db_conn.user", consts.EnvVarPrefix+"_"+consts.EnvVarDbUser)
	v.BindEnv("db_conn.password", consts.EnvVarPrefix+"_"+consts.EnvVarDbPwd)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	validate := validator.New()
	validate.RegisterValidation("file_if_provided", fileIfProvided)

	if err := validate.Struct(cfg); err != nil {
		if strings.Contains(err.Error(), "DBConn.User") {
			return nil, fmt.Errorf("database username must be set via %s_%s env var",
				consts.EnvVarPrefix, consts.EnvVarDbUser)
		}
		if strings.Contains(err.Error(), "DBConn.Password") {
			return nil, fmt.Errorf("database password must be set via %s_%s env var",
				consts.EnvVarPrefix, consts.EnvVarDbPwd)
		}
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	go runConfigWatcher(cfg)

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.handler_timeout", 3*time.Second)
	v.SetDefault("app.refresh_token_life", 720*time.Minute)
	v.SetDefault("app.access_token_life", 3*time.Minute)

	v.SetDefault("http_server.port", 8080)
	v.SetDefault("http_server.read_timeout", 5*time.Second)
	v.SetDefault("http_server.write_timeout", 5*time.Second)
	v.SetDefault("http_server.idle_timeout", 60*time.Second)

	v.SetDefault("grpc_audit_client.host", "localhost")
	v.SetDefault("grpc_audit_client.port", 9000)
	v.SetDefault("grpc_audit_client.timeout", 5*time.Second)

	v.SetDefault("db_conn.host", "localhost")
	v.SetDefault("db_conn.port", 5432)
	v.SetDefault("db_conn.name", "med_service")
	v.SetDefault("db_conn.sslmode", "verify-ca")
	v.SetDefault("db_conn.sslrootcert", "/etc/ssl/postgres/ca.crt")
}

func fileIfProvided(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	_, err := os.Stat(fl.Field().String())
	return err == nil
}

// startConfigWatcher watches & updates config data on the fly
func runConfigWatcher(cfg Config) {
	//TODO
}
