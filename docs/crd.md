#CRD

Octopus introduces two Custom Resources:

- `TestDefinition` defines your test as a Pod specification.
- `ClusterTestSuite` defines how to execute tests and what to execute. Status of the `ClusterTestSuite` contains 
information about executions of the tests. 
