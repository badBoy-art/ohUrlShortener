package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"ohurlshortener/storage"
)

func TestChangeUserState(t *testing.T) {
	init4Test(t)

	account := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	password := "-2aDzm=0(ln_9^1"
	if err := NewUser(account, password, false); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = storage.RedisDelete(AdminUserPrefix + account)
		_ = storage.DbNamedExec(`DELETE FROM public.users WHERE account = :account`, map[string]interface{}{"account": account})
	}()

	target, err := GetUserByAccountFromRedis(account)
	if err != nil {
		t.Fatal(err)
	}
	if target.IsEmpty() {
		t.Fatal("newly created user not found in Redis")
	}
	admin := adminOperator(t)

	// 启用状态下可正常登录
	if _, err := Login(account, password); err != nil {
		t.Errorf("login before disable should succeed, got: %v", err)
	}

	// 停用后登录被拒绝
	if err := ChangeUserState(account, false, admin); err != nil {
		t.Fatal(err)
	}
	if _, err := Login(account, password); err == nil || !strings.Contains(err.Error(), "停用") {
		t.Errorf("login after disable should fail with 停用 message, got: %v", err)
	}

	// 重新启用后恢复登录
	if err := ChangeUserState(account, true, admin); err != nil {
		t.Fatal(err)
	}
	if _, err := Login(account, password); err != nil {
		t.Errorf("login after re-enable should succeed, got: %v", err)
	}

	// 不能停用自己的账号
	if err := ChangeUserState(account, false, target); err == nil || !strings.Contains(err.Error(), "自己") {
		t.Errorf("disable self should fail with 自己 message, got: %v", err)
	}

	// 不能停用管理员账号（用另一个 admin 操作）
	adminAccount := fmt.Sprintf("testadmin_%d", time.Now().UnixNano())
	if err := NewUser(adminAccount, password, true); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = storage.RedisDelete(AdminUserPrefix + adminAccount)
		_ = storage.DbNamedExec(`DELETE FROM public.users WHERE account = :account`, map[string]interface{}{"account": adminAccount})
	}()
	if err := ChangeUserState(adminAccount, false, admin); err == nil || !strings.Contains(err.Error(), "管理员") {
		t.Errorf("disable admin should fail with 管理员 message, got: %v", err)
	}

	// 不存在的账号
	if err := ChangeUserState("no_such_user_xyz", false, admin); err == nil || !strings.Contains(err.Error(), "不存在") {
		t.Errorf("disable nonexistent user should fail with 不存在 message, got: %v", err)
	}
}
