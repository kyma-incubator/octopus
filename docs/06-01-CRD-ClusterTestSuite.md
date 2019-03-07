---
title: ClusterTestSuite
type: Custom Resource
---

The `ClusterTest` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define which tests to execute and how execute them. 
To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd ClusterTestSuite -o yaml
```

## Sample custom resource

This is a sample resource that requests to execute all tests defined on a Kubernetes cluster.

```
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: testsuite-all
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.selectors** | **NO** | Defines which tests should be executed. You can define tests by specifying their names or labels. Selectors are additive.  If nothing is provided, all tests from all namespaces will be executed.
| **spec.selectors.matchNames** | **NO** | List of TestDefinitions to execute. For every element on the list, specify **name** and **namespace** that refers to a TestDefinition. This feature is not yet implemented. |
| **spec.selectors.matchLabels** | **NO** | List of labels that matches labels of TestDefinitions. A TestDefinition is selected if **at least** one label matches. This feature is not yet implemented. | 
| **spec.concurrency** | **NO** | Defines how many tests can be executed at the same time. Depends on a cluster size and it's load. Default value is 1. This feature is not yet implemented. 
| **spec.suiteTimeout** | **NO** | Defines maximal suite duration after which test executions are interrupted and the suite is marked as a Failed. Default values is one hour. This feature is not yet implemented. 
| **spec.count** | **NO** | Defines how many times every test should be executed. **Spec.Count** and **Spec.MaxRetries** are mutually exclusive. Default value is 1. This feature is not yet implemented. 
| **spec.maxRetries** | **NO** | Defines how many times, a given test will be retried in case of it's failure. A Suite is marked as a Succeeded even if some test failed and then finally succeed. Default value is 0 - no retries. This feature is not yet implemented. 

## Custom resource status

This table lists all the possible status fields together with their descriptions:

| Field             |  Description |
|:-----------------:|:-------------:|
| **status.startTime** | Time when execution of suite's test starts. |
| **status.completionTime** | Time when execution of suite's test finished. |
| **status.conditions** | List of the suite conditions. |
| **status.conditions[].type** | Type of the condition. Suite can be in the following conditions: **Uninitialized**, **Running**, **Error**, **Failed**, **Succeeded**. |
| **status.conditions[].status** | Determines if the suite is in given state. Possible values are: **True**, **False**, **Unknown**. |
| **status.conditions[].reason** | One-word CamelCase reason for the condition's last transition. This field may be empty. |
| **status.conditions[].message** | Human-readable message indicating details about last transition. This field may be empty. |
| **status.results[]** | Test Result gathers all execution for given `TestDefinition` |
| **status.results[].name** | A name of the given `TestDefinition` |
| **status.results[].namespace** | A namespace where `TestDefinition` was defined |
| **status.results[].status** | A status of the `TestDefinition`. Possible values are: **NotYetScheduled**, **Scheduled**, **Running**, **Unknown**, **Failed**, **Succeeded**, **Skipped**. |
| **status.results[].executions[]** | List of executions for given `TestDefinition` |
| **status.results[].executions[].id** | Id of the execution that equals to the testing Pod name. |
| **status.results[].executions[].podPhase** | Phase of the testing Pod. Possible values are: **Pending**, **Running**, **Succeeded**, **Failed**, **Unknown** |
| **status.results[].executions[].startTime** | Time when testing Pod was observed in **Running** phase. |
| **status.results[].executions[].completionTime** | Time when testing pod was observed in **Succeeded** or **Failed** phase. |
| **status.results[].executions[].reason** | One-word, CamelCase reason for the Pod's phase last transition. |
 | **status.results[].executions[].message** | Human-readable message indicating details about last transition of Pod's phase. |



## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| TestDefinition | Defines test  |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Octopus |  When a `ClusterTestSuite` is created, `Octopus` starts executing tests. |