# Kubectl-suite

Kubectl-suite is a command line tool to manage Octopus resources. Currently it supports following Kubernetes clusters:
- minikube
- GCP

## Overview

## Prerequisites

Use the following tools to set up the project:

* Version 1.11 or higher of [Go](https://golang.org/dl/)
* Version 0.5.1 or higher of [Dep](https://github.com/golang/dep)

## Installation
Kubectl-suite can be configured as a `kubectl` [plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).
```bash
make resolve
cd ./cmd/kubectl-suite
go install .
```

After installation, plugin is used by specifying:
```
kubectl suite ...
```

## Usage
1. Get description of ClusterTestSuite:
```
kubectl suite {suite name}
```
The sample output looks as follows:
```
+--------------+--------------------------------+
|     NAME     |             VALUE              |
+--------------+--------------------------------+
| Name         | testsuite-all-2019-04-05-07-36 |
| Concurrency  |                              1 |
| Count        |                              1 |
| Duration     | 25m4s                          |
| Condition    | Succeeded                      |
| Tests        |                             14 |
| In Progress  |                              0 |
| Success      |                             14 |
| Failures     |                              0 |
| Executions.  |                             14 |
| Failed tests | -                              |
+--------------+--------------------------------+
```