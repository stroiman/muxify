package main_test

import (
	"math/rand"
	"os"
	"path/filepath"

	. "github.com/stroiman/muxify"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateRandomName() string {
	return randStringRunes(10)
}

func CreateRandomProjectName() string {
	return "project-" + CreateRandomName()
}

func MustCreateTestServer() TmuxServer {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configFile := filepath.Join(wd, "tmux.conf")
	return TmuxServer{
		SocketName: "test-socket",
		ConfigFile: configFile,
	}
}
