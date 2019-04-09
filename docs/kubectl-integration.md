# Integration with Kubectl

Octopus provides integration with `kubectl` to simplify the work on its resources.

## ClusterTestSuite template

Status of a `ClusterTestSuite` contains many information.
Output from following command:
```
kubectl get cts -oyaml {suite mame}
```
can be overwhelming. To get more concise but still informative output, you can use template stored in `kubectl/suite.template`.
To download recent version of the template, execute the following command:
```
curl -LO https://raw.githubusercontent.com/kyma-incubator/octopus/master/kubectl/suite.template
```
Then, set the `template` argument to path to the downloaded file:
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

In the above example, `test-monitoring` failed at the first time, then it was retried and finally it succeeded.  
