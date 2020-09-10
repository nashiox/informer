package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/urfave/cli"
)

var Commands = []cli.Command{
	commandWatch,
	commandReview,
	commandList,
}

var commandWatch = cli.Command{
	Name:  "watch",
	Usage: "",
	Description: `
`,
	Action: doWatch,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "output,o",
			Usage: "",
		},
	},
}

var commandReview = cli.Command{
	Name:  "review",
	Usage: "",
	Description: `
`,
	Action: doReview,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "delay,d",
			Value: 100,
			Usage: "",
		},
	},
}

var commandList = cli.Command{
	Name:  "list",
	Usage: "",
	Description: `
`,
	Action: doList,
}

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func doWatch(c *cli.Context) {
	if len(c.Args()) < 1 {
		_ = cli.ShowAppHelp(c)
		os.Exit(1)
	}

	tty := c.Args()[0]
	output := c.String("output")

	if output == "" {
		fp, err := ioutil.TempFile("/tmp", "informer")
		assert(err)
		defer fp.Close()
		output = fp.Name()
	}

	if !strings.HasPrefix(tty, "pts/") {
		fmt.Fprintf(os.Stderr, "Unrecognized psuedo terminal [%s]\n", tty)
		os.Exit(2)
	}

	if _, err := os.Stat("/dev/" + tty); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Psuedo terminal [%s] currently does NOT exist.\n", tty)
		os.Exit(2)
	}

	debug("DEBUG: Scanning for psuedo terminal ", tty)

	out, err := exec.Command("ps", "fauwwx").Output()
	assert(err)
	psreg := regexp.MustCompile(
		`\n(\S+)\s+(\d+)\s+\S+\s+\S+\s+\S+\s+\S+\s+\?\s+\S+\s+\S+\s+\S+\s+\S+[\|\\_ ]+\S*\bsshd\b.*\n\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+` + tty + `\s`,
	)

	if !psreg.Match(out) {
		fmt.Fprintf(os.Stderr, "Unable to locate corresponding ssh session for [%s]", tty)
		os.Exit(2)
	}

	pid := string(psreg.FindSubmatch(out)[2])

	cmd := exec.Command("strace", "-e", "read", "-s16384", "-q", "-x", "-p", pid, "-o", output)
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill() }()

	tmp, err := tail.TailFile(output, tail.Config{Follow: true})
	assert(err)

	fds := make(map[int]string, 2)
	keys := make([]int, 2)

	tmpreg := regexp.MustCompile(`(read)\((\d+), "(.*)"`)
	for line := range tmp.Lines {
		if tmpreg.Match([]byte(line.Text)) {
			group := tmpreg.FindSubmatch([]byte(line.Text))

			key, err := strconv.Atoi(string(group[2]))
			assert(err)
			fds[key] = string(group[1])
			if len(fds) >= 2 {
				for i := range fds {
					keys = append(keys, i)
				}
				sort.Ints(keys)
				break
			}
		}
	}
	tmp.Kill(nil)

	out, err = exec.Command("clear").Output()
	assert(err)
	fmt.Print(string(out))

	t, err := tail.TailFile(output, tail.Config{Follow: true})
	assert(err)
	defer t.Kill(nil)

	outreg := regexp.MustCompile(
		fmt.Sprintf(`read\(%d, "(.*)"`, keys[len(keys)-1]),
	)
	hexreg := regexp.MustCompile(`\\x(..)`)

	for line := range t.Lines {
		if outreg.Match([]byte(line.Text)) {
			s := string(outreg.FindSubmatch([]byte(line.Text))[1])

			s = hexreg.ReplaceAllStringFunc(s, func(src string) string {
				tname := hexreg.FindStringSubmatch(src)
				h, err := hex.DecodeString(tname[1])
				assert(err)

				return string(h)
			})

			s = strings.Replace(s, `\t`, "\t", -1)
			s = strings.Replace(s, `\r`, "\r", -1)
			s = strings.Replace(s, `\n`, "\n", -1)
			s = strings.Replace(s, `\\`, "\\", -1)
			s = strings.Replace(s, `\"`, `"`, -1)

			fmt.Print(s)
		}
	}
}

func doReview(c *cli.Context) {
	if len(c.Args()) < 1 {
		_ = cli.ShowAppHelp(c)
		os.Exit(1)
	}

	delay := c.Int("delay")

	fp, err := os.Open(c.Args()[0])
	assert(err)

	fds := make(map[int]string, 2)
	keys := make([]int, 2)

	scanner := bufio.NewScanner(fp)

	tmpreg := regexp.MustCompile(`(read)\((\d+), "(.*)"`)
	for scanner.Scan() {
		text := []byte(scanner.Text())
		if tmpreg.Match(text) {
			group := tmpreg.FindSubmatch(text)

			key, err := strconv.Atoi(string(group[2]))
			assert(err)
			fds[key] = string(group[1])
			if len(fds) >= 2 {
				for i := range fds {
					keys = append(keys, i)
				}
				sort.Ints(keys)
				break
			}
		}
	}

	fp.Close()

	out, err := exec.Command("clear").Output()
	assert(err)
	fmt.Print(string(out))

	fp, err = os.Open(c.Args()[0])
	assert(err)
	defer fp.Close()

	outreg := regexp.MustCompile(
		fmt.Sprintf(`read\(%d, "(.*)"`, keys[len(keys)-1]),
	)
	hexreg := regexp.MustCompile(`\\x(..)`)

	scanner = bufio.NewScanner(fp)
	for scanner.Scan() {
		text := []byte(scanner.Text())
		if outreg.Match(text) {
			s := string(outreg.FindSubmatch(text)[1])

			s = hexreg.ReplaceAllStringFunc(s, func(src string) string {
				tname := hexreg.FindStringSubmatch(src)
				h, err := hex.DecodeString(tname[1])
				assert(err)

				return string(h)
			})

			s = strings.Replace(s, `\t`, "\t", -1)
			s = strings.Replace(s, `\r`, "\r", -1)
			s = strings.Replace(s, `\n`, "\n", -1)
			s = strings.Replace(s, `\\`, "\\", -1)
			s = strings.Replace(s, `\"`, `"`, -1)

			fmt.Print(s)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	fmt.Println()
	assert(scanner.Err())
}

func doList(c *cli.Context) {
	out, err := exec.Command("w", "-hs").Output()
	assert(err)

	fmt.Println(string(out))
}
