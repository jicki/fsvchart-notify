package service

import (
	"database/sql"
	"errors"
	"time"

	"fsvchart-notify/internal/database"
	"fsvchart-notify/internal/models"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// 密钥用于签名JWT令牌
var jwtSecret = []byte("fsvchart-notify-secret-key")

// JWTClaims 定义JWT令牌的声明
type JWTClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(username, role string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour) // 令牌有效期24小时

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

// AuthenticateUser 验证用户凭据
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

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 解析时间
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	// 验证密码
	if !CheckPasswordHash(password, passwordHash) {
		return nil, errors.New("invalid password")
	}

	return &user, nil
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
