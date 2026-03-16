package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	} `yaml:"server"`
	Auth AuthConfig `yaml:"auth"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret   string     `yaml:"jwt_secret"`
	TokenExpiry int        `yaml:"token_expiry_hours"`
	LDAP        LDAPConfig `yaml:"ldap"`
}

// LDAPConfig LDAP 认证配置
type LDAPConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	UseTLS       bool   `yaml:"use_tls"`
	BindDN       string `yaml:"bind_dn"`
	BindPassword string `yaml:"bind_password"`
	BaseDN       string `yaml:"base_dn"`
	UserFilter   string `yaml:"user_filter"`
	DisplayAttr  string `yaml:"display_name_attr"`
	EmailAttr    string `yaml:"email_attr"`
	DefaultRole  string `yaml:"default_role"`
	AdminGroupDN string `yaml:"admin_group_dn"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &Config{}
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "fsvchart-notify-secret-key"
	}
	if cfg.Auth.TokenExpiry == 0 {
		cfg.Auth.TokenExpiry = 24
	}
	if cfg.Auth.LDAP.Port == 0 {
		cfg.Auth.LDAP.Port = 389
	}
	if cfg.Auth.LDAP.UserFilter == "" {
		cfg.Auth.LDAP.UserFilter = "(uid=%s)"
	}
	if cfg.Auth.LDAP.DisplayAttr == "" {
		cfg.Auth.LDAP.DisplayAttr = "cn"
	}
	if cfg.Auth.LDAP.EmailAttr == "" {
		cfg.Auth.LDAP.EmailAttr = "mail"
	}
	if cfg.Auth.LDAP.DefaultRole == "" {
		cfg.Auth.LDAP.DefaultRole = "user"
	}

	return cfg, nil
}
