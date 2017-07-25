# check_orchestrator
[![Build Status](https://travis-ci.org/mcrauwel/go-check-orchestrator.svg?branch=master)](https://travis-ci.org/mcrauwel/go-check-orchestrator)

This repository contains a Nagios / Icinga check to monitor [Orchestrator](https://github.com/github/orchestrator).

This check was written by Matthias Crauwels <matthias.crauwels@UGent.be> at Ghent University. It is published with an [MIT license](LICENSE)

## Usage
```
$ bin/check_orchestrator
Usage:
check_orchestrator [subcommand] [OPTIONS]
SubCommands:
clusterhealth
clusterinfo
status
```

## Commands
### status
#### Usage
```
$ bin/check_orchestrator status -h
Usage:
check_orchestrator status [OPTIONS]
Application Options:
-H, --host= Hostname (default: localhost)
-p, --port= Port (default: 3000)
-S, --ssl Use SSL
-I, --insecure Do not check SSL cert
-U, --uri= URI (default: api/health)
Help Options:
-h, --help Show this help message
```

#### Success
```
$ bin/check_orchestrator status
ORCHESTRATOR_STATUS OK: Application node is healthy
```

#### Errors
```
$ bin/check_orchestrator status
ORCHESTRATOR_STATUS CRITICAL: Application node is unhealthy dial tcp 127.0.0.1:20192: getsockopt: connection refused
```

### clusterinfo
#### Usage
```
$ bin/check_orchestrator clusterinfo -h
Usage:
check_orchestrator clusterinfo [OPTIONS]
Application Options:
-H, --host= Hostname (default: localhost)
-p, --port= Port (default: 3000)
-S, --ssl Use SSL
-I, --insecure Do not check SSL cert
-U, --uri= URI (default: api/clusters-info)
Help Options:
-h, --help Show this help message
```

#### Success
```
$ bin/check_orchestrator clusterinfo
ORCHESTRATOR_CLUSTERINFO OK: This instance manages following clusters: 127.0.0.1:20192 (HasAutomatedMasterRecovery = false) (HasAutomtedIntermediateMasterRecovery = false), localhost:20192 (HasAutomatedMasterRecovery = false) (HasAutomtedIntermediateMasterRecovery = false)
```

### clusterhealth
#### Usage
```
$ bin/check_orchestrator clusterhealth --help
Usage:
check_orchestrator clusterhealth --alias=<clusteralias> [OPTIONS]
Application Options:
-a, --alias= ClusterAlias
-H, --host= Hostname (default: localhost)
-p, --port= Port (default: 3000)
-S, --ssl Use SSL
-I, --insecure Do not check SSL cert
-t, --timeout= Timeout for SecondsSinceLastSeen (default: 300)
-w, --lag-warning= Slave lag warning threshold (default: 300)
-c, --lag-critical= Slave lag critical threshold (default: 600)
Help Options:
-h, --help Show this help message
```

#### Success
```
$ bin/check_orchestrator clusterhealth --alias=127.0.0.1:20192
ORCHESTRATOR_CLUSTERHEALTH OK: Cluster 127.0.0.1:20192 is doing OK
```

#### Errors
- no alias
```
$ bin/check_orchestrator clusterhealth
the required flag `-a, --alias' was not specified
```

- multiple writers
```
$ bin/check_orchestrator clusterhealth --alias=127.0.0.1:20192
ORCHESTRATOR_CLUSTERHEALTH CRITICAL: [SPLIT BRAIN] There are 2 writable servers in cluster 127.0.0.1:20192
```

- Slave thread(s) not running
```
$ bin/check_orchestrator clusterhealth --alias=127.0.0.1:20192
ORCHESTRATOR_CLUSTERHEALTH CRITICAL: In cluster 127.0.0.1:20192 the Slave_IO-thread is not running on host 127.0.0.1:20195
```
```
$ bin/check_orchestrator clusterhealth --alias=127.0.0.1:20192
ORCHESTRATOR_CLUSTERHEALTH CRITICAL: In cluster 127.0.0.1:20192 the Slave_SQL-thread is not running on host 127.0.0.1:20195
```

- Slave lag
```
$ bin/check_orchestrator clusterhealth --alias 127.0.0.1:20192 -w 30 -c 60
ORCHESTRATOR_CLUSTERHEALTH WARNING: In cluster 127.0.0.1:20192 host 127.0.0.1:20195 is 53 seconds lagging (warning threshold 30)

 $ bin/check_orchestrator clusterhealth --alias 127.0.0.1:20192 -w 30 -c 60
ORCHESTRATOR_CLUSTERHEALTH CRITICAL: In cluster 127.0.0.1:20192 host 127.0.0.1:20195 is 65 seconds lagging (critical threshold 60)
```

*note* the clusterhealth-command take the `downtime` setting in Orchestrator into account...
