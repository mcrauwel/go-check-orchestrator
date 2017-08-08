package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
)

type ClusterDetailResponse struct {
	Key                  ServerKey
	InstanceAlias        string
	Uptime               int
	ServerID             int
	ServerUUID           string
	Version              string
	VersonComment        string
	FlavorName           string
	ReadOnly             bool
	MasterKey            ServerKey
	IsDetachedMaster     bool
	Slave_SQL_Running    bool
	Slave_IO_Running     bool
	IsDetached           bool
	SecondsBehindMaster  JsonInt64
	SlaveLagSeconds      JsonInt64
	SecondsSinceLastSeen JsonInt64
	IsDowntimed          bool
	SQLDelay             int
}

type ServerKey struct {
	Hostname string
	Port     int
}

type JsonInt64 struct {
	Int64 int64
	Valid bool
}

type clusterHealtOpts struct {
	ClusterAlias string `required:"true" short:"a" long:"alias" description:"ClusterAlias"`
	orchestratorOpts
	SecondsSinceLastSeenThreshold int64 `short:"t" long:"timeout" default:"300" description:"Timeout for SecondsSinceLastSeen"`
	SlaveLagWarningThreshold      int64 `short:"w" long:"lag-warning" default:"300" description:"Slave lag warning threshold"`
	SlaveLagCriticalThreshold     int64 `short:"c" long:"lag-critical" default:"600" description:"Slave lag critical threshold"`
}

func checkClusterHealth(args []string) *checkers.Checker {

	opts := clusterHealtOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	psr.Usage = "clusterhealth --alias=<clusteralias> [OPTIONS]"
	_, err := psr.ParseArgs(args)
	if err != nil {
		os.Exit(1)
	}

	clusterAlias := opts.ClusterAlias
	uri := fmt.Sprintf("%s://%s:%s/api/cluster/alias/%s", sslPrefix(opts.SSL), opts.Host, opts.Port, clusterAlias)
	client := &http.Client{Transport: getHttpTransport(opts.NoCert)}
	resp, err := client.Get(uri)
	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprintf("Could not connect to Orchestrator API on %s", uri))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// check first if we might have gotten a StatusResponse instead of a ClusterDetailResponse (example: "No cluster found for alias")
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

	// The response was not a StatusResponse, so try to process is as a ClusterDetailResponse
	var r []ClusterDetailResponse
	err = json.Unmarshal(body, &r)

	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprintf("Could read content for the Orchestrator API on %s\n%s", uri, err))
	}

	nrOfWriters := 0
	for _, s := range r {
		if !s.ReadOnly {
			nrOfWriters++
		}
	}

	if nrOfWriters > 1 {
		return checkers.NewChecker(checkers.CRITICAL,
			fmt.Sprintf("[SPLIT BRAIN] There are %d writable servers in cluster %s", nrOfWriters, clusterAlias))
	}

	if nrOfWriters < 1 {
		return checkers.NewChecker(checkers.CRITICAL,
			fmt.Sprintf("[READ ONLY CLUSTER] There are %d writable servers in cluster %s", nrOfWriters, clusterAlias))
	}

	for _, s := range r {
		if s.IsDowntimed {
			continue
		}

		if s.MasterKey.Hostname != "" {
			if !s.Slave_IO_Running {
				return checkers.NewChecker(checkers.CRITICAL,
					fmt.Sprintf("In cluster %s the Slave_IO-thread is not running on host %s:%d",
						clusterAlias, s.Key.Hostname, s.Key.Port))
			}

			if !s.Slave_SQL_Running {
				return checkers.NewChecker(checkers.CRITICAL,
					fmt.Sprintf("In cluster %s the Slave_SQL-thread is not running on host %s:%d",
						clusterAlias, s.Key.Hostname, s.Key.Port))
			}
		}

		slaveLagSeconds := s.SlaveLagSeconds.Int64
		if s.SQLDelay > 0 {
			slaveLagSeconds = s.SlaveLagSeconds.Int64 - int64(s.SQLDelay)
		}
		if slaveLagSeconds > opts.SlaveLagCriticalThreshold {
			return checkers.NewChecker(checkers.CRITICAL,
				fmt.Sprintf("In cluster %s host %s:%d is %d seconds lagging (critical threshold %d)",
					clusterAlias, s.Key.Hostname, s.Key.Port, s.SlaveLagSeconds.Int64, opts.SlaveLagCriticalThreshold))
		}

		if slaveLagSeconds > opts.SlaveLagWarningThreshold {
			return checkers.NewChecker(checkers.WARNING,
				fmt.Sprintf("In cluster %s host %s:%d is %d seconds lagging (warning threshold %d)",
					clusterAlias, s.Key.Hostname, s.Key.Port, s.SlaveLagSeconds.Int64, opts.SlaveLagWarningThreshold))
		}

		if s.SecondsSinceLastSeen.Int64 > opts.SecondsSinceLastSeenThreshold {
			return checkers.NewChecker(checkers.WARNING,
				fmt.Sprintf("In cluster %s the host %s:%d was not seen for %d seconds (warning limit %d)",
					clusterAlias, s.Key.Hostname, s.Key.Port, s.SecondsSinceLastSeen.Int64, opts.SecondsSinceLastSeenThreshold))
		}
	}

	return checkers.NewChecker(checkers.OK, fmt.Sprintf("Cluster %s is doing OK", clusterAlias))
}
