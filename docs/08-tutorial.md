## Tutorial
In this tutorial, you create `TestDefinition` and `ClusterTestSuite` to show Octopus in action.

1. Prepare data
Create `TestDefinition` and `ClusterTestSuite`:

```bash
cd $GOPATH/src/github.com/kyma-incubator/octopus
kubectl apply -f config/samples/testdefinition.yaml
kubectl apply -f config/samples/testsuite.yaml

```
Created `TestDefinition` is very simple, it just executes `pwd` command in the `alpine:latest` 
image, so it should always succeed.  
 

2. Run the controller:
```
cd $GOPATH/src/github.com/kyma-incubator/octopus
make run
```
Then you should see in the logs similar output:
```
{"level":"info","ts":...,"logger":"cts_controller","msg":"Initialize suite","suite":"testsuite-all"}
{"level":"info","ts":...,"logger":"cts_controller","msg":"Testing pod created","suite":"testsuite-all","podName":"octopus-testing-pod-jz9qq","podNs":"default"}
{"level":"info","ts":...,"logger":"cts_controller","msg":"Do nothing, suite is finished","suite":"testsuite-all"}
```

After controller finished processing created test suite, it is be marked as a `Succeeded`:
`kubectl get ClusterTestSuite testsuite-all -oyaml`

```
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: testsuite-all
  uid: ba1a98e3-3ef5-11e9-b7e0-3af374a09d13
spec:
  maxRetries: 1
status:
  completionTime: 2019-03-06T12:37:20Z
  conditions:
  - status: "False"
    type: Running
  - status: "True"
    type: Succeeded
  results:
  - executions:
    - completionTime: 2019-03-06T12:37:20Z
      id: octopus-testing-pod-jz9qq
      podPhase: Succeeded
      startTime: 2019-03-06T12:37:15Z
    name: test-kubeless
    namespace: default
    status: Succeeded
  startTime: 2019-03-06T12:37:15Z

```

Under the hood, Octopus created the following testing Pod:
```
apiVersion: v1
kind: Pod
metadata:
  generateName: octopus-testing-pod-
  labels:
    testing.kyma-project.io/created-by-octopus: "true"
    testing.kyma-project.io/def-name: test-kubeless
    testing.kyma-project.io/suite-name: testsuite-all
  name: octopus-testing-pod-jz9qq
  namespace: default
  ownerReferences:
  - apiVersion: testing.kyma-project.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: ClusterTestSuite
    name: testsuite-all
    uid: ba1a98e3-3ef5-11e9-b7e0-3af374a09d13
spec:
  containers:
  - command:
    - pwd
    image: alpine:latest
    imagePullPolicy: Always
    name: test
    resources: {}
  restartPolicy: Never

```