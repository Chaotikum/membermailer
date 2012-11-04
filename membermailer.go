package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"io/ioutil"
	"encoding/json"
	"text/template"
)

const (
	TPL_START string = `From: vorstand@chaotikum.org
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

type User struct {
	Id		int
	Salutation	string
	Forename	string
	Surname		string
	Email		string
	Joined		string
	Address		Address
	Subscription	Subscription
}

func usage() {
	fmt.Printf("usage: %s tpl data.json\n", os.Args[0]);
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

func main() {
	// TODO: Flag for From:, optional, default vorstand@
	// also tpl, all the other files are json data files
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		usage()
	}

	tplData, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(10);
	}

	jsonData, err := ioutil.ReadFile(args[1])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(20);
	}

	var user User
	err = json.Unmarshal(jsonData, &user)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(30);
	}

	tplStr := strings.Join([]string{TPL_START, string(tplData), TPL_END}, "")

	funcs := template.FuncMap{"eq": tpl_eq}
	tpl := template.Must(template.New("mail").Funcs(funcs).Parse(tplStr))

	cmd := exec.Command("/usr/sbin/sendmail", "-t", user.Email)
	stdin, _ := cmd.StdinPipe()
	err = cmd.Start()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(40)
	}
	// TODO: execute the template before sendmail, in case of errors
	tpl.Execute(stdin, user)
	stdin.Close()
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(40)
	}
}
