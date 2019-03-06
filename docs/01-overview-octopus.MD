
# Overview

Octopus is a testing framework that allows you to run tests defined as Docker images on a running cluster.

## API

Octopus introduces two Custom Resources:

- `TestDefinition` defines your test as a Pod specification.
- `ClusterTestSuite` defines how to execute tests and what to execute. Status of the `ClusterTestSuite` contains 
information about executions of the tests. 

Detailed description of CRDs can be found [here](./file-does-not-exist.md)

