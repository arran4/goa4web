package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
	"github.com/chzyer/readline"
)

// replCmd provides an interactive shell for running goa4web commands.
type replCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

type replJob struct {
	id   int
	cmd  string
	done chan error
}

func parseReplCmd(parent *rootCmd, args []string) (*replCmd, error) {
	c := &replCmd{rootCmd: parent}
	c.fs = config.NewRuntimeFlagSet("repl")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *replCmd) Run() error {
	hist := filepath.Join(os.TempDir(), "goa4web_repl_history")
	rl, err := readline.NewEx(&readline.Config{Prompt: "> ", HistoryFile: hist})
	if err != nil {
		return err
	}
	defer rl.Close()

	var mu sync.Mutex
	jobs := map[int]*replJob{}
	jobID := 0

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		}
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		if strings.HasPrefix(line, "set ") {
			parts := strings.SplitN(strings.TrimSpace(line[4:]), "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			} else {
				fmt.Println("usage: set KEY=VALUE")
			}
			continue
		}
		if line == "jobs" {
			mu.Lock()
			for id, j := range jobs {
				select {
				case err, ok := <-j.done:
					if ok {
						close(j.done)
						if err != nil {
							fmt.Printf("[%d] %s: %v\n", id, j.cmd, err)
						} else {
							fmt.Printf("[%d] %s: done\n", id, j.cmd)
						}
						delete(jobs, id)
					} else {
						fmt.Printf("[%d] %s: done\n", id, j.cmd)
						delete(jobs, id)
					}
				default:
					fmt.Printf("[%d] %s: running\n", id, j.cmd)
				}
			}
			mu.Unlock()
			continue
		}
		if strings.HasPrefix(line, "wait ") {
			idStr := strings.TrimSpace(line[5:])
			mu.Lock()
			j, ok := jobs[atoi(idStr)]
			mu.Unlock()
			if ok {
				err := <-j.done
				if err != nil {
					fmt.Printf("[%d] %s: %v\n", j.id, j.cmd, err)
				} else {
					fmt.Printf("[%d] %s: done\n", j.id, j.cmd)
				}
				mu.Lock()
				delete(jobs, j.id)
				mu.Unlock()
			}
			continue
		}

		background := false
		if strings.HasSuffix(line, "&") {
			background = true
			line = strings.TrimSpace(strings.TrimSuffix(line, "&"))
		}

		if strings.HasPrefix(line, "!") {
			shcmd := strings.TrimSpace(line[1:])
			cmd := exec.Command("sh", "-c", shcmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if background {
				mu.Lock()
				jobID++
				j := &replJob{id: jobID, cmd: line, done: make(chan error, 1)}
				jobs[jobID] = j
				mu.Unlock()
				go func() { j.done <- cmd.Run() }()
				fmt.Printf("[%d] %s\n", j.id, j.cmd)
			} else {
				if err := cmd.Run(); err != nil {
					fmt.Println(err)
				}
			}
			continue
		}

		args := strings.Fields(line)
		args = expandEnv(args)
		r, err := parseRoot(append([]string{c.Prog()}, args...))
		if err != nil {
			if !errors.Is(err, flag.ErrHelp) {
				fmt.Println(err)
			}
			continue
		}
		if background {
			mu.Lock()
			jobID++
			j := &replJob{id: jobID, cmd: line, done: make(chan error, 1)}
			jobs[jobID] = j
			mu.Unlock()
			go func() { j.done <- r.Run() }()
			fmt.Printf("[%d] %s\n", j.id, j.cmd)
		} else {
			if err := r.Run(); err != nil && !errors.Is(err, flag.ErrHelp) {
				fmt.Println(err)
			}
			r.Close()
		}
	}

	return nil
}

func expandEnv(args []string) []string {
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = os.ExpandEnv(a)
	}
	return out
}

// Usage prints command usage information with examples.
func (c *replCmd) Usage() { executeUsage(c.fs.Output(), "repl_usage.txt", c) }

func (c *replCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*replCmd)(nil)

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
