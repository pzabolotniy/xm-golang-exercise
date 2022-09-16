package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type App struct {
	DB          *DB          `mapstructure:"db"`
	WebAPI      *WebAPI      `mapstructure:"web_api"`
	GeoIP       *GeoIP       `mapstructure:"geoip"` //nolint:tagliatelle // need to discuss
	ClientToken *ClientToken `mapstructure:"client_token"`
}

type DB struct {
	ConnString      string        `mapstructure:"conn_string"`
	MigrationDir    string        `mapstructure:"migration_dir"`
	MigrationTable  string        `mapstructure:"migration_table"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type WebAPI struct {
	Listen string `mapstructure:"listen"`
}

type GeoIP struct {
	AllowedCountryName string        `mapstructure:"allowed_country_name"`
	EndPoint           string        `mapstructure:"endpoint"` //nolint:tagliatelle // need to discuss
	Timeout            time.Duration `mapstructure:"timeout"`
}

type ClientToken struct {
	Issuer string        `mapstructure:"issuer"`
	Secret string        `mapstructure:"secret"`
	TTL    time.Duration `mapstructure:"ttl"`
}

func LoadConfig() (*App, error) {
	viper.SetConfigName("config") // hardcoded config name
	viper.SetConfigType("yaml")   // hardcoded extension
	viper.AddConfigPath(".")      // hardcoded configfile path
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("unable to read config from file: %w", err)
	}
	viper.AutomaticEnv()

	config := new(App)
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}

	return config, nil
}
