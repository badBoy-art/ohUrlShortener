// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package service

import (
	"fmt"
	"testing"
	"time"

	"ohurlshortener/core"
	"ohurlshortener/storage"

	"github.com/bxcodec/faker/v3"
)

// adminOperator 返回本地库中的 admin 用户（ohUrlShortener）
func adminOperator(t *testing.T) core.User {
	t.Helper()
	admin, err := storage.FindUserByAccount("ohUrlShortener")
	if err != nil {
		t.Fatalf("find admin user failed: %v", err)
	}
	return admin
}

func TestGenerateShortUrl(t *testing.T) {

	init4Test(t)

	admin := adminOperator(t)
	for i := 0; i < 100000; i++ {
		url := faker.URL()
		_, err := GenerateShortUrl(url, url+" | memo", 0, admin)
		if err != nil {
			t.Error(err)
			continue
		}
	}
}

func TestGenerateShortUrlIdempotentAndAppendDests(t *testing.T) {
	init4Test(t)

	admin := adminOperator(t)
	destUrl := fmt.Sprintf("https://idem.example.com/%d", time.Now().UnixNano())
	defer func() {
		code, _ := core.GenerateShortLink(destUrl)
		_ = DeleteUrlAndAccessLogs(code, admin)
	}()

	// 首次创建
	code1, err := GenerateShortUrlWithDests(destUrl, "", 0, nil, admin)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// 同一长链接重复创建：幂等返回同一短码，且不报错
	code2, err := GenerateShortUrlWithDests(destUrl, "", 0, nil, admin)
	if err != nil {
		t.Fatalf("idempotent create failed: %v", err)
	}
	if code1 != code2 {
		t.Errorf("idempotent create returned different codes: %s != %s", code1, code2)
	}

	// 追加多目标地址（携带已有短码）
	dests := []core.ShortUrlDest{
		{Label: "pc", DestUrl: "https://pc.example.com"},
		{Label: "mobile", DestUrl: "https://m.example.com"},
	}
	res, err := AppendShortUrlDests(code1, dests, admin)
	if err != nil {
		t.Fatalf("append dests failed: %v", err)
	}
	if res != code1 {
		t.Errorf("append dests returned %s, want %s", res, code1)
	}

	// label 冲突：报错且不覆盖原 dest_url（防篡改）
	conflict := []core.ShortUrlDest{
		{Label: "pc", DestUrl: "https://pc-v2.example.com"},
	}
	if _, err := AppendShortUrlDests(code1, conflict, admin); err == nil {
		t.Fatal("append dests with label conflict should fail")
	}

	// 新增一个不冲突的 label：追加成功
	more := []core.ShortUrlDest{
		{Label: "ios", DestUrl: "https://ios.example.com"},
	}
	if _, err := AppendShortUrlDests(code1, more, admin); err != nil {
		t.Fatalf("append new label failed: %v", err)
	}

	// 数据库落库校验
	saved, err := loadShortUrlDests(code1)
	if err != nil {
		t.Fatalf("load dests failed: %v", err)
	}
	if saved["pc"] != "https://pc.example.com" {
		t.Errorf("pc label = %q, want unchanged https://pc.example.com", saved["pc"])
	}
	if saved["mobile"] != "https://m.example.com" {
		t.Errorf("mobile label = %q, want kept https://m.example.com", saved["mobile"])
	}
	if saved["ios"] != "https://ios.example.com" {
		t.Errorf("ios label = %q, want appended https://ios.example.com", saved["ios"])
	}

	// Redis 缓存重写校验
	cached, err := Search4ShortUrl(code1)
	if err != nil {
		t.Fatalf("search cached url failed: %v", err)
	}
	if cached.Dests["pc"] != "https://pc.example.com" {
		t.Errorf("cached pc label = %q, want https://pc.example.com", cached.Dests["pc"])
	}
	if cached.Dests["ios"] != "https://ios.example.com" {
		t.Errorf("cached ios label = %q, want https://ios.example.com", cached.Dests["ios"])
	}

	// 追加到不存在的短码：报错
	if _, err := AppendShortUrlDests("NONEXIST", dests, admin); err == nil {
		t.Error("append dests to nonexistent short url should fail")
	}
}
