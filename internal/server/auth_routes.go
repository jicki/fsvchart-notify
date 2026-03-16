package server

import (
	"net/http"
	"strconv"

	"fsvchart-notify/internal/models"
	"fsvchart-notify/internal/service"

	"github.com/gin-gonic/gin"
)

// 登录处理
func login(c *gin.Context) {
	var loginReq models.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// 验证用户凭据
	user, err := service.AuthenticateUser(loginReq.Username, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Authentication failed",
			"error":   err.Error(),
		})
		return
	}

	// 生成JWT令牌
	token, err := service.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate token",
			"error":   err.Error(),
		})
		return
	}

	// 返回登录成功响应
	c.JSON(http.StatusOK, models.LoginResponse{
		Token:       token,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	})
}

// 获取当前用户信息
func getCurrentUser(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Not authenticated",
		})
		return
	}

	// 获取用户信息
	user, err := service.GetUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get user information",
			"error":   err.Error(),
		})
		return
	}

	// 返回用户信息（不包含密码）
	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"email":        user.Email,
		"role":         user.Role,
		"auth_source":  user.AuthSource,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	})
}

// 修改密码
func changePassword(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Not authenticated",
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// 更新密码
	err := service.UpdateUserPassword(username.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Failed to change password",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Password changed successfully",
	})
}

// 更新用户信息
func updateUserInfo(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Not authenticated",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// 更新用户信息
	err := service.UpdateUserInfo(username.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update user information",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "User information updated successfully",
	})
}

// 获取所有用户列表（仅管理员）
func listUsers(c *gin.Context) {
	users, err := service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败",
			"error":   err.Error(),
		})
		return
	}
	// 确保返回空数组而非 null
	if users == nil {
		users = []models.User{}
	}
	c.JSON(http.StatusOK, users)
}

// 修改用户角色（仅管理员）
func updateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	var req models.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 校验 role 值
	if req.Role != "admin" && req.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色只能为 admin 或 user",
		})
		return
	}

	if err := service.UpdateUserRole(userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "修改角色失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "角色修改成功",
	})
}

// 管理员重置用户密码
func adminResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	var req models.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求格式错误",
			"error":   err.Error(),
		})
		return
	}

	if err := service.AdminResetPassword(userID, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "重置密码失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码重置成功",
	})
}
