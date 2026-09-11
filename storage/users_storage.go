package storage

import (
	"strings"

	"ohurlshortener/core"
	"ohurlshortener/utils"

	"github.com/btcsuite/btcd/btcutil/base58"
)

// FindAllUsers 获取所有用户
func FindAllUsers() ([]core.User, error) {
	var found []core.User
	query := `SELECT * FROM public.users u`
	return found, DbSelect(query, &found)
}

func FindPagedUsers(page, size int) ([]core.User, error) {
	var found []core.User

	if page > 0 && size > 0 {
		offset := (page - 1) * size
		query := `SELECT * FROM public.users u ORDER BY u.id DESC LIMIT $1 OFFSET $2`
		return found, DbSelect(query, &found, size, offset)
	}

	return found, nil
}

// NewUser 新建用户
func NewUser(account string, password string, isAdmin bool) error {
	query := `INSERT INTO public.users (account, "password", is_admin) VALUES(:account,:password,:is_admin)`
	data, err := PasswordBase58Hash(password)
	if err != nil {
		return err
	}
	return DbNamedExec(query, core.User{Account: account, Password: data, IsAdmin: isAdmin})
}

// FindUserByPassword 按密码哈希查找用户（API token 即密码哈希）
func FindUserByPassword(password string) (core.User, error) {
	var user core.User
	query := `SELECT * FROM public.users u WHERE u.password = $1`
	err := DbGet(query, &user, password)
	return user, err
}

// UpdateUser 更新用户
func UpdateUser(user core.User) error {
	query := `UPDATE public.users SET account = :account , "password" = :password WHERE id = :id`
	return DbNamedExec(query, user)
}

// UpdateUserEnable 启用/停用用户（按账号，大小写不敏感）
func UpdateUserEnable(account string, enabled bool) error {
	query := `UPDATE public.users SET is_enable = :is_enable WHERE lower(account) = :account`
	return DbNamedExec(query, map[string]interface{}{
		"is_enable": enabled,
		"account":   strings.ToLower(strings.TrimSpace(account)),
	})
}

// FindUserByAccount 根据账号查找用户
func FindUserByAccount(account string) (core.User, error) {
	var user core.User
	query := `SELECT * FROM public.users u WHERE lower(u.account) = $1`
	return user, DbGet(query, &user, strings.ToLower(account))
}

// PasswordBase58Hash 密码加密
func PasswordBase58Hash(password string) (string, error) {
	data, err := utils.Sha256Of(password)
	if err != nil {
		return "", err
	}
	return base58.Encode(data), nil
}
