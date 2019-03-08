# Tutorial
Follow this tutorial to create `TestDefinition` and `ClusterTestSuite` CRDs and to see Octopus in action.

1. Prepare Kubernetes cluster and ensure that `kubectl` is properly configured. One option is to use `minikube`.  
1. Clone Octopus directory to your `$GOPATH`
```
cd $GOPATH/src
mkdir -p github.com/kyma-incubator/
cd github.com/kyma-incubator/
git clone https://github.com/kyma-incubator/octopus.git
```
1. Create `TestDefinition` and `ClusterTestSuite` CRDs:

```bash
cd octopus
kubectl apply -f config/samples/testdefinition.yaml
kubectl apply -f config/samples/testsuite.yaml

```
2. Run Octopus:
```
cd $GOPATH/src/github.com/kyma-incubator/octopus
make run
```
The sample output looks as follows:
```
{"level":"info","ts":...,"logger":"cts_controller","msg":"Initialize suite","suite":"testsuite-all"}
{"level":"info","ts":...,"logger":"cts_controller","msg":"Testing pod created","suite":"testsuite-all","podName":"octopus-testing-pod-jz9qq","podNs":"default"}
{"level":"info","ts":...,"logger":"cts_controller","msg":"Do nothing, suite is finished","suite":"testsuite-all"}
```

As you can see from the output, Octopus detects ClusterTestSuite and initialize it with the matching TestDefinition. 
Then, a testing Pod is created. When a testing Pod finishes, processing of the ClusterTestSuite is completed.
You can see that the suite is marked as a **Succeeded**:

```
kubectl get ClusterTestSuite testsuite-all -oyaml
```
The sample output looks as follows:
```
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  name: testsuite-all
status:
  completionTime: 2019-03-06T12:37:20Z
  conditions:
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