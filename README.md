# Octopus

## Overview

Octopus is a testing framework that allows you to run tests defined as Docker images on a running cluster.
It was created to replace the `helm test` with providing additional features:
1. Selective testing.  
This can be especially useful for local development when a developer wants to execute only tests that 
his local changes can affect.  
1. Retries on failed tests. 
Flaky tests are an everyday reality. Definitely, developers should strive to improve tests and make them stable, but we should accept that they will exist.
1. Run tests multiple times.
This can be helpful when a developer adds a new test and he/she needs to validate if a test is stable or he wants to reproduce a problem that sometimes occurs on CI.
1. Full support for concurrent testing. 
A developer should be able to decide if a given test is prepared to be executed concurrently with other tests.
A person who executes tests should be able how many concurrent tests can be executed, depending on the cluster size. 

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