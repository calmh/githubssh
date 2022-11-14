package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
)

var cli struct {
	GithubUser string `arg:"" desc:"Github user" default:"calmh"`
	KeysPath   string `arg:"" type:"path" desc:"Path to authorized_keys" default:"~/.ssh/authorized_keys"`
}

func main() {
	kong.Parse(&cli)

	localKeys, err := localKeys(cli.KeysPath)
	if err != nil {
		log.Fatalln("Reading existing keys:", err)
	}

	remoteKeys, err := githubKeys(cli.GithubUser)
	if err != nil {
		log.Fatalln("Reading remote keys:", err)
	}

	var ks keyset
	ks.Add(localKeys...)
	ks.Add(remoteKeys...)

	if err := saveKeys(cli.KeysPath, ks.Keys()); err != nil {
		log.Fatalln("Saving keys:", err)
	}
}

func localKeys(path string) ([]string, error) {
	bs, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bs), "\n"), nil
}

func githubKeys(user string) ([]string, error) {
	res, err := http.Get(fmt.Sprintf("https://github.com/%s.keys", user))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d", res.StatusCode)
	}
	defer res.Body.Close()
	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bs), "\n"), nil
}

func saveKeys(path string, keys []string) error {
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	bs := []byte(strings.Join(keys, "\n") + "\n")
	if err := os.WriteFile(path+".tmp", bs, 0600); err != nil {
		return err
	}
	return os.Rename(path+".tmp", path)
}

type keyset struct {
	keysWithComment map[string]struct{}
	keysNoComment   map[string]bool
}

func (ks *keyset) Add(keys ...string) {
	if ks.keysWithComment == nil {
		ks.keysWithComment = make(map[string]struct{})
	}
	if ks.keysNoComment == nil {
		ks.keysNoComment = make(map[string]bool)
	}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || strings.HasPrefix(key, "#") {
			continue
		}
		fields := strings.Fields(key)
		switch len(fields) {
		case 2:
			// Key has no comment, likely from GitHub. Add it to the
			// no-comment set only if we didn't already see the same key
			// with a comment.
			if _, ok := ks.keysNoComment[key]; !ok {
				ks.keysNoComment[key] = true
			}
		case 3:
			// Key has a comment, add it to the with-comment set and taint
			// the no-comment set.
			ks.keysWithComment[key] = struct{}{}
			ks.keysNoComment[strings.Join(fields[:2], " ")] = false
		}
	}
}

func (ks *keyset) Keys() []string {
	var keys []string
	for key := range ks.keysWithComment {
		keys = append(keys, key)
	}
	for key, use := range ks.keysNoComment {
		if !use {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
