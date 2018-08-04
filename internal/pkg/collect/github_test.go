package collect

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHub(t *testing.T) {
	localServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, err := ioutil.ReadFile("github_response.json")
		if err != nil {
			t.Error(err)
		}
		fmt.Fprint(w, string(content))
	}))

	result, err := GitHub(localServer.URL, "charlieegan3")

	if err != nil {
		t.Error(err)
	}
	if result.Repo.Name != "charlieegan3/dotfiles" {
		t.Error(result)
	}
	if result.Commit.Message != "Install rmagick and docker compose" {
		t.Error(result)
	}
}
