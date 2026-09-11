// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"ohurlshortener/core"
	"ohurlshortener/storage"
)

// TestShortUrlOwnership 校验短链接归属权限：仅创建者与 admin 可操作
func TestShortUrlOwnership(t *testing.T) {
	init4Test(t)

	admin := adminOperator(t)
	userA := newTestUser(t, "ownership_a")
	userB := newTestUser(t, "ownership_b")

	destUrl := fmt.Sprintf("https://owner.example.com/%d", time.Now().UnixNano())
	code, err := GenerateShortUrlWithDests(destUrl, "", 0, nil, userA)
	if err != nil {
		t.Fatalf("user A create failed: %v", err)
	}
	defer func() { _ = DeleteUrlAndAccessLogs(code, admin) }()

	// 落库归属校验
	saved, err := storage.FindShortUrl(code)
	if err != nil || saved.IsEmpty() {
		t.Fatalf("find saved short url failed: %v", err)
	}
	if saved.CreatedBy != userA.ID {
		t.Errorf("created_by = %d, want %d", saved.CreatedBy, userA.ID)
	}

	dests := []core.ShortUrlDest{{Label: "pc", DestUrl: "https://pc.example.com"}}

	// 用户 B 对 A 的短链操作全部无权限
	if _, err := ChangeState(code, false, userB); !errors.Is(err, ErrNoPermission) {
		t.Errorf("B change state: want ErrNoPermission, got %v", err)
	}
	if err := DeleteUrlAndAccessLogs(code, userB); !errors.Is(err, ErrNoPermission) {
		t.Errorf("B delete: want ErrNoPermission, got %v", err)
	}
	if _, err := AppendShortUrlDests(code, dests, userB); !errors.Is(err, ErrNoPermission) {
		t.Errorf("B append: want ErrNoPermission, got %v", err)
	}
	if _, err := GetShortUrlStats(code, userB); !errors.Is(err, ErrNoPermission) {
		t.Errorf("B stats: want ErrNoPermission, got %v", err)
	}

	// B 用与 A 相同的 destUrl 创建：报「已由其他用户创建」
	if _, err := GenerateShortUrlWithDests(destUrl, "", 0, nil, userB); err == nil {
		t.Error("B create same destUrl: want error, got nil")
	}

	// 创建者本人可以操作
	if _, err := ChangeState(code, true, userA); err != nil {
		t.Errorf("A change state own link failed: %v", err)
	}

	// admin 可以操作所有
	if _, err := ChangeState(code, false, admin); err != nil {
		t.Errorf("admin change state failed: %v", err)
	}
	if _, err := ChangeState(code, true, admin); err != nil {
		t.Errorf("admin re-enable failed: %v", err)
	}
	if _, err := GetShortUrlStats(code, admin); err != nil {
		t.Errorf("admin stats failed: %v", err)
	}
	if _, err := AppendShortUrlDests(code, dests, admin); err != nil {
		t.Errorf("admin append failed: %v", err)
	}
}

// TestPagedUrlsOwnership 校验分页列表按归属过滤：普通用户仅见自己名下短链，admin 见全部
func TestPagedUrlsOwnership(t *testing.T) {
	init4Test(t)

	admin := adminOperator(t)
	userA := newTestUser(t, "list_a")
	userB := newTestUser(t, "list_b")

	created := map[string]bool{}
	create := func(owner core.User, host string) string {
		t.Helper()
		destUrl := fmt.Sprintf("https://%s.example.com/%d", host, time.Now().UnixNano())
		code, err := GenerateShortUrlWithDests(destUrl, "", 0, nil, owner)
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		created[code] = true
		return code
	}
	codeA1 := create(userA, "list-a1")
	codeA2 := create(userA, "list-a2")
	codeB1 := create(userB, "list-b1")
	defer func() {
		for code := range created {
			_ = DeleteUrlAndAccessLogs(code, admin)
		}
	}()

	// 用户 A 的列表只包含自己名下两条
	urlsA, err := GetPagesShortUrls("", 1, 100, userA)
	if err != nil {
		t.Fatalf("user A list failed: %v", err)
	}
	codesA := map[string]bool{}
	for _, u := range urlsA {
		if u.CreatedBy != userA.ID {
			t.Errorf("user A list leaked link %s created_by = %d, want %d", u.ShortUrl, u.CreatedBy, userA.ID)
		}
		codesA[u.ShortUrl] = true
	}
	if !codesA[codeA1] || !codesA[codeA2] {
		t.Error("user A list missing own links")
	}
	if codesA[codeB1] {
		t.Error("user A list leaked user B link")
	}

	// admin 列表包含全部
	urlsAdmin, err := GetPagesShortUrls("", 1, 100, admin)
	if err != nil {
		t.Fatalf("admin list failed: %v", err)
	}
	codesAdmin := map[string]bool{}
	for _, u := range urlsAdmin {
		codesAdmin[u.ShortUrl] = true
	}
	for _, code := range []string{codeA1, codeA2, codeB1} {
		if !codesAdmin[code] {
			t.Errorf("admin list missing link %s", code)
		}
	}
}

// newTestUser 创建普通用户并返回其完整信息
func newTestUser(t *testing.T, prefix string) core.User {
	t.Helper()
	account := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	if err := storage.NewUser(account, "test-password-123", false); err != nil {
		t.Fatalf("create test user failed: %v", err)
	}
	user, err := storage.FindUserByAccount(account)
	if err != nil || user.IsEmpty() {
		t.Fatalf("find test user failed: %v", err)
	}
	if user.IsAdmin {
		t.Fatalf("new user %s should not be admin", account)
	}
	return user
}
