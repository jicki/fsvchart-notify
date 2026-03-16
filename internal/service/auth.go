package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"fsvchart-notify/internal/config"
	"fsvchart-notify/internal/database"
	"fsvchart-notify/internal/models"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// 密钥用于签名JWT令牌
var jwtSecret = []byte("fsvchart-notify-secret-key")

var authConfig *config.AuthConfig

// InitAuth 初始化认证配置
func InitAuth(cfg *config.AuthConfig) {
	authConfig = cfg
	jwtSecret = []byte(cfg.JWTSecret)
}

// JWTClaims 定义JWT令牌的声明
type JWTClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(username, role string) (string, error) {
	nowTime := time.Now()

	hours := 24
	if authConfig != nil && authConfig.TokenExpiry > 0 {
		hours = authConfig.TokenExpiry
	}
	expireTime := nowTime.Add(time.Duration(hours) * time.Hour)

	claims := JWTClaims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  nowTime.Unix(),
			Issuer:    "fsvchart-notify",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

// ParseToken 解析JWT令牌
func ParseToken(token string) (*JWTClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*JWTClaims); ok && tokenClaims.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// HashPassword 对密码进行哈希处理
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码哈希
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// AuthenticateUser 验证用户凭据（本地优先，LDAP 回退）
func AuthenticateUser(username, password string) (*models.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database connection failed")
	}

	var user models.User
	var passwordHash string
	var createdAt, updatedAt string

	err := db.QueryRow(`
		SELECT id, username, password, COALESCE(display_name, '') as display_name,
		       COALESCE(email, '') as email, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.DisplayName,
		&user.Email, &user.Role, &createdAt, &updatedAt,
	)

	if err == nil {
		// 本地用户存在，验证密码
		user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		if CheckPasswordHash(password, passwordHash) {
			return &user, nil
		}

		// 本地密码不匹配，尝试 LDAP
		if authConfig != nil && authConfig.LDAP.Enabled {
			ldapUser, ldapErr := AuthenticateLDAP(authConfig.LDAP, username, password)
			if ldapErr == nil {
				// LDAP 验证成功，更新本地用户信息
				if ldapUser.DisplayName != "" && user.DisplayName == "" {
					db.Exec(`UPDATE users SET display_name = ? WHERE id = ?`, ldapUser.DisplayName, user.ID)
					user.DisplayName = ldapUser.DisplayName
				}
				if ldapUser.Email != "" && user.Email == "" {
					db.Exec(`UPDATE users SET email = ? WHERE id = ?`, ldapUser.Email, user.ID)
					user.Email = ldapUser.Email
				}
				// 同步 LDAP admin 组角色
				newRole := authConfig.LDAP.DefaultRole
				if ldapUser.IsAdmin {
					newRole = "admin"
				}
				if user.Role != newRole {
					db.Exec(`UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, newRole, user.ID)
					user.Role = newRole
				}
				return &user, nil
			}
		}

		return nil, errors.New("invalid password")
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// 本地用户不存在，尝试 LDAP 认证
	if authConfig != nil && authConfig.LDAP.Enabled {
		ldapUser, ldapErr := AuthenticateLDAP(authConfig.LDAP, username, password)
		if ldapErr != nil {
			return nil, fmt.Errorf("认证失败: %w", ldapErr)
		}

		// LDAP 认证成功，自动创建本地用户
		hashedPassword, hashErr := HashPassword(password)
		if hashErr != nil {
			return nil, fmt.Errorf("密码加密失败: %w", hashErr)
		}

		displayName := ldapUser.DisplayName
		if displayName == "" {
			displayName = username
		}

		// 根据 LDAP admin 组决定角色
		role := authConfig.LDAP.DefaultRole
		if ldapUser.IsAdmin {
			role = "admin"
		}

		result, insertErr := db.Exec(`
			INSERT INTO users (username, password, display_name, email, role)
			VALUES (?, ?, ?, ?, ?)
		`, username, hashedPassword, displayName, ldapUser.Email, role)
		if insertErr != nil {
			return nil, fmt.Errorf("创建本地用户失败: %w", insertErr)
		}

		id, _ := result.LastInsertId()
		log.Printf("LDAP 用户首次登录，已创建本地账号: %s (ID: %d)", username, id)

		return &models.User{
			ID:          id,
			Username:    username,
			DisplayName: displayName,
			Email:       ldapUser.Email,
			Role:        role,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	}

	return nil, errors.New("user not found")
}

// GetUserByUsername 通过用户名获取用户信息
func GetUserByUsername(username string) (*models.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database connection failed")
	}

	var user models.User
	var createdAt, updatedAt string

	err := db.QueryRow(`
		SELECT id, username, password, COALESCE(display_name, '') as display_name,
		       COALESCE(email, '') as email, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.DisplayName,
		&user.Email, &user.Role, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 解析时间
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &user, nil
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(username, oldPassword, newPassword string) error {
	// 先验证用户和旧密码
	user, err := AuthenticateUser(username, oldPassword)
	if err != nil {
		return err
	}

	// 哈希新密码
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	db := database.GetDB()
	_, err = db.Exec(`
		UPDATE users
		SET password = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, hashedPassword, user.ID)

	return err
}

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(username string, info models.UpdateUserRequest) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database connection failed")
	}

	_, err := db.Exec(`
		UPDATE users
		SET display_name = ?, email = ?, updated_at = CURRENT_TIMESTAMP
		WHERE username = ?
	`, info.DisplayName, info.Email, username)

	return err
}
