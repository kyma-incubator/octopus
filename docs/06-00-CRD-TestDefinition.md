---
title: TestDefinition
type: Custom Resource
---

The `TestDefinition` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define acceptance tests executed as a Pod 
on a Kubernes cluster. 
To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd TestDefinition -o yaml
```

## Sample custom resource
This is a sample resource that defines the simplest test definition. 

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
| **spec.template** |    **YES**   | An object that describes the pod that will be created. This field is of PodTemplateSpec type from Kubernetes API and its detailed description can be found [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#podtemplatespec-v1-core)  |
| **spec.skip**     |    **NO**    | Mark that test should not be executed. The default vaule is `false`. This feature is not yet implemented. |
| **spec.disableConcurrency** | **NO** | Disallow running the test concurrently with other tests. The default vaule is `false`. This feature is not yet implemented. 
| **spec.timeout** | **NO** | Define maximal duration of the test, after which it should be terminated and it's execution marked as a **Failed**. No The default vaule. This feature is not yet implemented. 

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| ClusterTestSuite |  ClusterTestSuite defines which `TestDefinitions` to execute.  |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Octopus |  When a `ClusterTestSuite` is created, `Octopus` fetches `TestDefinitions` that matches the  `ClusterTestSuite` selectors.  |