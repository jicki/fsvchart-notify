package service

import (
	"crypto/tls"
	"fmt"
	"log"

	"fsvchart-notify/internal/config"

	"github.com/go-ldap/ldap/v3"
)

// LDAPUser 表示从 LDAP 获取的用户信息
type LDAPUser struct {
	Username    string
	DisplayName string
	Email       string
}

// AuthenticateLDAP 使用 LDAP 验证用户
func AuthenticateLDAP(cfg config.LDAPConfig, username, password string) (*LDAPUser, error) {
	var scheme string
	if cfg.UseTLS {
		scheme = "ldaps"
	} else {
		scheme = "ldap"
	}
	url := fmt.Sprintf("%s://%s:%d", scheme, cfg.Host, cfg.Port)

	conn, err := ldap.DialURL(url, ldap.DialWithTLSConfig(&tls.Config{
		InsecureSkipVerify: false,
	}))
	if err != nil {
		return nil, fmt.Errorf("LDAP 连接失败: %w", err)
	}
	defer conn.Close()

	// 使用管理员账号绑定搜索
	if cfg.BindDN != "" {
		err = conn.Bind(cfg.BindDN, cfg.BindPassword)
		if err != nil {
			return nil, fmt.Errorf("LDAP 管理员绑定失败: %w", err)
		}
	}

	// 搜索用户
	filter := fmt.Sprintf(cfg.UserFilter, ldap.EscapeFilter(username))
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter,
		[]string{"dn", cfg.DisplayAttr, cfg.EmailAttr},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP 搜索失败: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("LDAP 用户未找到: %s", username)
	}
	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("LDAP 搜索到多个用户: %s", username)
	}

	entry := sr.Entries[0]

	// 使用用户 DN 和密码进行绑定验证
	err = conn.Bind(entry.DN, password)
	if err != nil {
		return nil, fmt.Errorf("LDAP 密码验证失败")
	}

	log.Printf("LDAP 认证成功: %s", username)

	return &LDAPUser{
		Username:    username,
		DisplayName: entry.GetAttributeValue(cfg.DisplayAttr),
		Email:       entry.GetAttributeValue(cfg.EmailAttr),
	}, nil
}
