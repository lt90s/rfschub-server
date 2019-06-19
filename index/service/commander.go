package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

type commander struct {
	c   *exec.Cmd
	in  io.WriteCloser
	out *bufio.Scanner
}

func newCommander(binPath string) *commander {
	cmd := exec.CommandContext(context.Background(), binPath, "--_interactive", "--fields=*")
	in, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err = cmd.Start(); err != nil {
		panic(err)
	}

	c := &commander{
		c:   cmd,
		in:  in,
		out: bufio.NewScanner(out),
	}

	c.selfTest()

	return c
}

func (c *commander) selfTest() {
	type helloMessage struct {
		Type    string `json:"_type"`
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	var h helloMessage
	line, err := c.readLine()
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(line, &h); err != nil {
		panic(err)
	}
	if h.Name != "Universal Ctags" {
		log.Panic("indexer must use Universal Ctags")
	}
	log.Debugf("ctags self test pass: ctags=%s version=%s", h.Name, h.Version)
}

func (c *commander) readLine() ([]byte, error) {
	if !c.out.Scan() {
		return nil, c.out.Err()
	}
	return c.out.Bytes(), nil
}

type request struct {
	Command  string `json:"command"`
	FileName string `json:"filename"`
	Size     int    `json:"size"`
}

type ResponseEntry struct {
	Type string `json:"_type"`

	Id        string `json:"_id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Pattern   string `json:"pattern"`
	Language  string `json:"language"`
	Line      int    `json:"line"`
	Kind      string `json:"kind"`
	Scope     string `json:"scope"`
	ScopeKind string `json:"scopeKind"`
}

func (c *commander) indexFile(fileName string, content []byte) (entries []ResponseEntry, err error) {
	r := request{
		Command:  "generate-tags",
		FileName: fileName,
		Size:     len(content),
	}

	rb, err := json.Marshal(r)
	if err != nil {
		log.Debug("IndexFile marshal request error", "request", r, "error", err)
		return
	}

	_, err = c.in.Write(append(rb, '\n'))
	if err != nil {
		log.Debug("IndexFile write request error", "request", r, "error", err)
		return
	}
	_, err = c.in.Write(content)
	if err != nil {
		log.Debug("IndexFile write content error", "content", string(content), "error", err)
		return
	}

	entries = make([]ResponseEntry, 0, 256)
	var entry ResponseEntry
	for {
		var line []byte

		line, err = c.readLine()
		if err != nil {
			log.Debug("IndexFile readLine error", "error", err)
			return
		}
		err = json.Unmarshal(line, &entry)
		if err != nil {
			return
		}
		if entry.Type == "" {
			err = errors.New("unexpected empty type")
		} else if entry.Type == "completed" {
			break
		}
		entries = append(entries, entry)
	}
	return
}

func (c *commander) stop() {
	c.c.Process.Kill()
}
