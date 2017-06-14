package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	//"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
)

type statusOpts struct {
	orchestratorOpts
	URI string `short:"U" long:"uri" default:"api/health" description:"URI"`
}

func checkStatus(args []string) *checkers.Checker {
	opts := statusOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	psr.Usage = "status [OPTIONS]"
	_, err := psr.ParseArgs(args)
	if err != nil {
		os.Exit(1)
	}

	uri := fmt.Sprintf("%s://%s:%s/%s", sslPrefix(opts.SSL), opts.Host, opts.Port, opts.URI)
	client := &http.Client{Transport: getHttpTransport(opts.NoCert)}
	resp, err := client.Get(uri)
	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprintf("Could not connect to Orchestrator API on %s", uri))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var r StatusResponse
	json.Unmarshal(body, &r)

	msg := r.Message
	checkSt := checkers.OK

	if r.Code != "OK" {
		checkSt = checkers.CRITICAL
	}

	return checkers.NewChecker(checkSt, msg)
}
