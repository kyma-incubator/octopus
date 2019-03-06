# Octopus

## Overview

Octopus is a testing framework that allows you to run tests defined as Docker images on a running cluster.


## Purpose
Octopus was created in response to `helm test` limitations:
1. There is no possibility to run only a subset of tests. 
This can be especially useful for local development, when developer wants to execute only tests that 
his local changes can affect.  
1. There is no possibility to retry only failed tests. Flaky tests are an everyday reality. 
Definitely, developers should strive to improve tests and make them stable, but we should accept that they will exist.
1. There is no possibility to run given test many times. 
This can be helpful when a developer adds a new test and he/she needs to validate if a test is stable or he wants to reproduce a problem that sometimes occurs on CI.
1. There is only limited support for concurrent testing. 
Developer should be able to decide if given test is prepared to be executed concurrently with other tests.
Person who executes tests should be able how many concurent tests can be executed, depending on the cluster size. 

## Prerequisites

Use the following tools to set up the project:

* Version 1.11 or higher of [Go](https://golang.org/dl/)
* Version 0.5.1 or higher of [Dep](https://github.com/golang/dep)
* Version 2.0.0 of [Kustomize](https://github.com/kubernetes-sigs/kustomize)
* Version 1.0.7 of [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Mockery](https://github.com/vektra/mockery) 

## Development

### Install dependencies

This project uses `dep` as the dependency manager. To install all required dependencies, use the following command:
```bash
make resolve
```

### Run tests

To test your changes before each commit, use the following command:

```bash
make validate
```