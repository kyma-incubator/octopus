# Kubectl Extensions

Octopus provides integration with kubectl to simplify work on its resources.


## Concise template for ClusterTestSuite

You can get the full status of ClusterTestSuite by running:
```
kubectl get cts -oyaml {suite mame}
```
However, the output of this status contains a lot of information, which can be overwhelming. 
To get more concise but still informative output, use the template stored in `kubectl/suite.template`. 
To download the recent version of the template, run:
```
curl -LO https://raw.githubusercontent.com/kyma-incubator/octopus/master/kubectl/suite.template
```
Then, set the `--template` argument to point to the downloaded file:
```
 kubectl get cts {suite name} -ogo-template-file --template={path to template file}
```

The simplified output looks as follows:
```
Name:           {suite name}
Concurrency:    1.0
MaxRetries:     1.0
StartTime:      2019-04-06T12:56:11Z
CompletionTime: 2019-04-06T13:22:38Z
Condition:      Succeeded
Tests:
    test-monitoring - + 
    test-events +
    test-ui ?
```

Next to the test name, there is a symbol which specifies the status of a given test execution:
- `+` - test execution passed
- `-` - test execution failed
- `?` - test execution is in progress or not yet scheduled

>**NOTE:** In the given example, `test-monitoring` failed at first, then it was retried****, and finally it succeeded.  