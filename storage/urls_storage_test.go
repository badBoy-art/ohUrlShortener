package storage

import (
	"database/sql"
	"testing"
	"time"

	"ohurlshortener/core"

	"github.com/bxcodec/faker/v3"
)

func TestInsertShortUrls(t *testing.T) {
	init4Test(t)

	admin, err := FindUserByAccount("ohUrlShortener")
	if err != nil || admin.IsEmpty() {
		t.Fatalf("find admin user failed: %v", err)
	}

	for i := 0; i < 10000; i++ {
		destUrl := faker.URL()
		shortUrl, _ := core.GenerateShortLink(destUrl)
		url := core.ShortUrl{DestUrl: destUrl, ShortUrl: shortUrl, CreatedAt: time.Now(), Valid: true, Memo: sql.NullString{String: destUrl, Valid: true}, CreatedBy: admin.ID}
		err := InsertShortUrl(url)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestFindAllShortUrlsAfterID(t *testing.T) {
	init4Test(t)

	const pageSize = 100

	var lastID int64
	var pages int
	for {
		urls, err := FindAllShortUrlsAfterID(lastID, pageSize)
		if err != nil {
			t.Fatal(err)
		}
		if len(urls) == 0 {
			break
		}
		pages++
		if len(urls) > pageSize {
			t.Fatalf("page %d returned %d rows, want <= %d", pages, len(urls), pageSize)
		}
		for i, u := range urls {
			if u.ID <= lastID {
				t.Fatalf("page %d row %d id %d not strictly greater than cursor %d", pages, i, u.ID, lastID)
			}
			if i > 0 && u.ID <= urls[i-1].ID {
				t.Fatalf("page %d ids not strictly ascending: %d after %d", pages, u.ID, urls[i-1].ID)
			}
		}
		lastID = urls[len(urls)-1].ID
	}
	if pages == 0 {
		t.Fatal("expected at least one page of short urls")
	}
	t.Logf("walked %d pages up to id %d", pages, lastID)
}

func TestDeleteShortUrlWithAccessLogs(t *testing.T) {

	init4Test(t)

	url1 := core.ShortUrl{ShortUrl: "hello"}
	url2 := core.ShortUrl{ShortUrl: "hello"}
	url3 := core.ShortUrl{ShortUrl: "hello"}

	type args struct {
		shortUrl core.ShortUrl
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "TestCase1", args: args{url1}, wantErr: false},
		{name: "TestCase1", args: args{url2}, wantErr: false},
		{name: "TestCase1", args: args{url3}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteShortUrlWithAccessLogs(tt.args.shortUrl); (err != nil) != tt.wantErr {
				t.Errorf("DeleteShortUrlWithAccessLogs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
