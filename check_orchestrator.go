package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mackerelio/checkers"
)

var commands = map[string](func([]string) *checkers.Checker){
	// "replication":    checkReplication,
	"clusterhealth": checkClusterHealth,
	"clusterinfo":   checkClusterInfo,
	"status":        checkStatus,
}

type StatusResponse struct {
	Code    string
	Message string
	Details []string
}

type orchestratorOpts struct {
	Host   string `short:"H" long:"host" default:"localhost" description:"Hostname"`
	Port   string `short:"p" long:"port" default:"3000" description:"Port"`
	SSL    bool   `short:"S" long:"ssl" description:"Use SSL"`
	NoCert bool   `short:"I" long:"insecure" description:"Do not check SSL cert"`
}

func separateSub(argv []string) (string, []string) {
	if len(argv) == 0 || strings.HasPrefix(argv[0], "-") {
		return "", argv
	}
	return argv[0], argv[1:]
}

func sslPrefix(useSSL bool) string {
	if useSSL {
		return "https"
	}

	return "http"
}

func getHttpTransport(allowInsecure bool) *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: allowInsecure},
	}
}

func main() {
	subCmd, argv := separateSub(os.Args[1:])
	fn, ok := commands[subCmd]
	if !ok {
		fmt.Println(`Usage:
  check_orchestrator [subcommand] [OPTIONS]
SubCommands:`)
		for k := range commands {
			fmt.Printf("  %s\n", k)
		}
		os.Exit(1)
	}

	ckr := fn(argv)
	//fmt.Println(result)
	ckr.Name = fmt.Sprintf("ORCHESTRATOR_%s", strings.ToUpper(subCmd))
	ckr.Exit()

}
