package storage

import (
	"testing"

	"ohurlshortener/utils"
)

func TestNewUser(t *testing.T) {
	init4Test(t)
	// NewUser("ohUrlShortener", "-2aDzm=0(ln_9^1")
	NewUser("ohUrlShortener1", "-2aDzm=0(ln_9^1", false)
	NewUser("ohUrlShortener2", "-2aDzm=0(ln_9^1", true)
}

func TestFindUserByPassword(t *testing.T) {
	init4Test(t)
	// 种子账号：token 即其存储的密码哈希
	user, err := FindUserByPassword("EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t")
	if err != nil {
		t.Fatal(err)
	}
	if user.IsEmpty() || user.Account != "ohUrlShortener" {
		t.Errorf("FindUserByPassword() = %+v, want ohUrlShortener", user)
	}
	if !user.IsAdmin {
		t.Error("seed account ohUrlShortener should be admin")
	}

	// 不存在的 token
	user, err = FindUserByPassword("nonexistent-token")
	if err != nil {
		t.Fatal(err)
	}
	if !user.IsEmpty() {
		t.Errorf("FindUserByPassword(nonexistent) = %+v, want empty", user)
	}
}

func init4Test(t *testing.T) {
	_, err := utils.InitConfig("../config.ini")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = InitDatabaseService()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestPasswordBase58Hash(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "TestPasswordBase58Hash", want: "EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t", wantErr: false, args: args{password: "-2aDzm=0(ln_9^1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PasswordBase58Hash(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("PasswordBase58Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PasswordBase58Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}
