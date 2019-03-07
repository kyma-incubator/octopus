---
title: ClusterTestSuite
type: Custom Resource
---

The `ClusterTestSuite` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define a suite of tests to execute and how to execute them. 
To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd ClusterTestSuite -o yaml
```

## Sample custom resource

This is a sample resource that requests for execution of all tests defined on a Kubernetes cluster.

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
| **spec.selectors** | **NO** | Defines which tests should be executed. You can define tests by specifying their names or labels. Selectors are additive. If not defined, all tests from all Namespaces are executed.
| **spec.selectors.matchNames** | **NO** | List of TestDefinitions to execute. For every element on the list, specify **name** and **namespace** that refers to a TestDefinition. This feature is not yet implemented. |
| **spec.selectors.matchLabels** | **NO** | Lists labels that match TestDefinitions labels. A TestDefinition is selected if at least one label matches. This feature is not yet implemented. | 
| **spec.concurrency** | **NO** | Defines how many tests can be executed at the same time, which depends on cluster size and its load. The default value is `1`. This feature is not yet implemented. 
| **spec.suiteTimeout** | **NO** | Defines the maximal suite duration after which test executions are interrupted and marked as **Failed**. The default value is one hour. This feature is not yet implemented. 
| **spec.count** | **NO** | Defines how many times every test should be executed. **Spec.Count** and **Spec.MaxRetries** are mutually exclusive. The default value is `1`. This feature is not yet implemented. 
| **spec.maxRetries** | **NO** | Defines how many times a given test is retried in case of its failure. A suite is marked as a **Succeeded** even if some test failed and then finally succeeded. The default value is `0`, which means that there are no retries of a given test. This feature is not yet implemented. 

## Custom resource status

This table lists all the possible status fields together with their descriptions:

| Field             |  Description |
|:-----------------:|:-------------:|
| **status.startTime** | Specifies time when the suite's test execution starts. |
| **status.completionTime** | Specifies time when the suite's test execution finishes. |
| **status.conditions** | Lists the suite conditions. |
| **status.conditions[].type** | Specifies the type of condition. These are the possible suite conditions: **Uninitialized**, **Running**, **Error**, **Failed**, and **Succeeded**. |
| **status.conditions[].status** | Determines if the suite is in a given state. The possible values are **True**, **False**, and **Unknown**. |
| **status.conditions[].reason** | Specifies one-word, CamelCase reason for the condition's last transition. This field may be empty. |
| **status.conditions[].message** | Provides a human-readable message with details about the last transition. This field may be empty. |
| **status.results[]** | Gathers all executions for a given TestDefinition. |
| **status.results[].name** | Specifies a name of a given TestDefinition. |
| **status.results[].namespace** | Specifies a Namespace where a TestDefinition is defined. |
| **status.results[].status** | Provides the status of a TestDefinition. The possible values are **NotYetScheduled**, **Scheduled**, **Running**, **Unknown**, **Failed**, **Succeeded**, and **Skipped**. |
| **status.results[].executions[]** | Lists executions for a given TestDefinition. |
| **status.results[].executions[].id** | Provides the ID of an execution that is the same as the testing Pod name. |
| **status.results[].executions[].podPhase** | Specifies the phase of the testing Pod. The possible values are **Pending**, **Running**, **Succeeded**, **Failed**, and **Unknown**. |
| **status.results[].executions[].startTime** | Time when testing Pod was observed in **Running** phase. |
| **status.results[].executions[].completionTime** | Time when testing pod was observed in **Succeeded** or **Failed** phase. |
| **status.results[].executions[].reason** | Provides one-word, CamelCase reason for the Pod's phase last transition. |
 | **status.results[].executions[].message** | Provides a human-readable message with details about last Pod's phase transition. |



## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| TestDefinition | Defines test  |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Octopus |  When a ClusterTestSuite is created, Octopus starts executing tests. |
