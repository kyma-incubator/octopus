# TestDefinition Custom Resource Definition

The `TestDefinition` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define acceptance tests executed as a Pod 
on a Kubernes cluster. 
To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd testdefinitions.testing.kyma-project.io -o yaml
```

## Sample custom resource
This is a sample resource that provides a simple test definition. 

```
apiVersion: testing.kyma-project.io/v1alpha1
kind: TestDefinition
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: test-dummy
spec:
  template:
    spec:
      containers:
        - name: test
          image: alpine:latest
          command:
            - "pwd"
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:-----------:|:-------------------:|:-------------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.template** |    **YES**   | Describes the Pod that will be created. This field is of `PodTemplateSpec` type from the Kubernetes API. Find its detailed description [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#podtemplatespec-v1-core)  |
| **spec.skip**     |    **NO**    | Indicates that a test should not be executed. The default value is `false`. This feature is not yet implemented. |
| **spec.disableConcurrency** | **NO** | Disallows running the given test concurrently. The default value is `false`. 
| **spec.timeout** | **NO** | Defines the maximal duration of a test, after which it is terminated and marked as **Failed**. This feature is not yet implemented.
| **spec.description** | **NO** | Describes the test case in detail (e.g. scope, test scenario, edge cases, known limitations etc.).


## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| [ClusterTestSuite](./crd-cluster-test-suite.md) |  Defines `TestDefinitions` to execute.  |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Octopus |  When a ClusterTestSuite is created, Octopus fetches TestDefinitions that match the ClusterTestSuite selectors.  |
