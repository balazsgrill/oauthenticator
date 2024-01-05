package file_test

import (
	"testing"

	"github.com/balazsgrill/oauthenticator/persistence/file"
)

func Test_loadjson(t *testing.T) {
	d := &file.Configdata{}
	err := d.Load("../../example.json")
	if err != nil {
		t.Fatal(err)
	}

	if d.Label_ != "example config" {
		t.Fail()
	}
	if d.Params["scope"] != "profile all" {
		t.Fail()
	}
}
