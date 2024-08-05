package muxify_test

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"

	. "github.com/stroiman/muxify"
)

var removeControlCharRegexp *regexp.Regexp = regexp.MustCompile(`\\\d{3}`)

func removeControlCharacters(s string) string {
	return removeControlCharRegexp.ReplaceAllString(s, "")
}

func GetLines(r io.ReadCloser) <-chan string {
	c := make(chan string)
	s := bufio.NewScanner(r)
	go func() {
		for s.Scan() {
			c <- s.Text()
		}
		close(c)
	}()
	return c
}

type TmuxControl struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func StartControlMode(server TmuxServer, session TmuxSession) (result TmuxControl, err error) {
	server.ControlMode = true
	result.cmd = server.Command("attach", "-t", session.Id).Cmd
	result.stdout, err = result.cmd.StdoutPipe()
	if err != nil {
		return
	}
	result.stdin, err = result.cmd.StdinPipe()
	if err != nil {
		return
	}
	err = result.cmd.Start()
	return
}

func MustStartControlMode(server TmuxServer, session TmuxSession) TmuxControl {
	result, err := StartControlMode(server, session)
	if err != nil {
		panic(err)
	}
	return result
}

func (c TmuxControl) Close() error {
	c.stdin.Close()
	return c.cmd.Wait()
}

func (c TmuxControl) MustClose() {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}

type TmuxOutputEvent struct {
	PaneId string
	Data   string
}
