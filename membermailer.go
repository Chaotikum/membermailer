package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"bytes"
	"os/exec"
	"strings"
	"io/ioutil"
	"encoding/json"
	"text/template"
)

const (
	// RFC3339 style
	DATE_FORMAT string = "2006-01-02"

	TPL_START string = `
To: {{.Email}}
Content-Type: text/plain; charset=utf-8
`
	TPL_END string = `
.`
)

type Address struct {
	Zip		string
	City		string
	Street		string
	Number		string
}

type Subscription struct {
	Reduced		bool
	Fee		int
	Interval	string
}

type Member struct {
	Id		int
	Salutation	string
	Forename	string
	Surname		string
	Email		string
	Joined		string
	Address		Address
	Subscription	Subscription
}

var (
	date	time.Time
	dateStr	string
	from	string
	tplPath	string
)

func init() {
	flag.StringVar(&dateStr, "date", "now", "Date for datef()")
	flag.StringVar(&from, "from", "vorstand@chaotikum.org", "From-Header")
	flag.StringVar(&tplPath, "tpl", "", "Path to the template file, required")
}

func usage() {
	fmt.Printf("usage: %s -tpl file.tpl member.json [...]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func tpl_eq(args ...interface{}) bool {
	if len(args) == 0 {
		return false
	}
	x := args[0]
	switch x := x.(type) {
	case string, int, int64, byte, float32, float64:
		for _, y := range args[1:] {
			if x == y {
				return true
			}
		}
		return false
	}
	// XXX: verlgeicht keine struct etc
	return false
}

func tpl_datef(format string) string {
	return date.Format(format)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var err error
	if dateStr != "" && dateStr != "now" {
		date, err = time.Parse(DATE_FORMAT, dateStr)
		if err != nil {
			date = time.Now()
		}
	} else {
		date = time.Now()
	}

	// member files
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	tplData, err := ioutil.ReadFile(tplPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(10);
	}

	tplStr := strings.Join([]string{"From: ", from, TPL_START, string(tplData), TPL_END}, "")

	funcs := template.FuncMap{
		"eq":		tpl_eq,
		"datef":	tpl_datef,
	}
	tpl := template.Must(template.New("mail").Funcs(funcs).Parse(tplStr))

	for _, path := range args {
		jsonData, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(20);
		}

		var member Member
		err = json.Unmarshal(jsonData, &member)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(30);
		}

		var b bytes.Buffer
		err = tpl.Execute(&b, member)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tpl exec for member %d returned error: %s\n", member.Id, err)
			continue
		}

		cmd := exec.Command("/usr/sbin/sendmail", "-t", member.Email)
		stdin, _ := cmd.StdinPipe()
		err = cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(40)
		}
		b.WriteTo(stdin)
		stdin.Close()
		err = cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(50)
		}

		fmt.Printf("queued: %d (%s %s)\n", member.Id, member.Forename, member.Surname)
	}
}
