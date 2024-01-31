package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
)

type clusterInfoOpts struct {
	orchestratorOpts
	URI string `short:"U" long:"uri" default:"api/clusters-info" description:"URI"`
}

type ClusterInfoResponse struct {
	ClusterName                            string
	ClusterAlias                           string
	ClusterDomain                          string
	CountInstances                         int
	HeuristicLag                           int
	HasAutomatedMasterRecovery             bool
	HasAutomatedIntermediateMasterRecovery bool
}

func checkClusterInfo(args []string) *checkers.Checker {

	opts := clusterInfoOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	psr.Usage = "clusterinfo [OPTIONS]"
	_, err := psr.ParseArgs(args)
	if err != nil {
		os.Exit(1)
	}
	uri := fmt.Sprintf("%s://%s:%s/%s", sslPrefix(opts.SSL), opts.Host, opts.Port, opts.URI)
	client := &http.Client{Transport: getHttpTransport(opts.NoCert)}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprintf("Could not connect to Orchestrator API on %s", uri))
	}
	if opts.HttpAuthName != "" && opts.HttpAuthPass != ""  {
		req.SetBasicAuth(opts.HttpAuthName, opts.HttpAuthPass)
	}
	resp, err := client.Do(req)
	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprintf("Could not connect to Orchestrator API on %s", uri))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// check first if we might have gotten a StatusResponse instead of a ClusterInfoResponse
	var status StatusResponse
	err = json.Unmarshal(body, &status)

	if err == nil {
		msg := status.Message
		checkSt := checkers.OK

		if status.Code != "OK" {
			checkSt = checkers.CRITICAL
		}

		return checkers.NewChecker(checkSt, msg)
	}

	// The response was not a StatusResponse, so try to process is as a ClusterInfoResponse
	var r []ClusterInfoResponse
	err = json.Unmarshal(body, &r)

	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprintf("Could read content for the Orchestrator API on %s\n%s", uri, err))
	}

	var aliases []string
	var aliasdetails []string
	for _, s := range r {
		alias := fmt.Sprintf("%s (HasAutomatedMasterRecovery = %t) (HasAutomtedIntermediateMasterRecovery = %t)",
			s.ClusterAlias, s.HasAutomatedMasterRecovery, s.HasAutomatedIntermediateMasterRecovery)

		aliases = append(aliases, s.ClusterAlias)
		aliasdetails = append(aliasdetails, alias)
	}

	if len(aliases) > 0 {
		return checkers.NewChecker(checkers.OK, fmt.Sprintf("This instance manages following clusters: %s\n%s", strings.Join(aliases, ", "), strings.Join(aliasdetails, "\n")))
	}

	return checkers.NewChecker(checkers.WARNING, "This Orchestrator is responding correctly but is not managing any clusters.")
}
