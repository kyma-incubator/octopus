# Integration with Kubectl

Octopus provides integration with `kubectl` to simplify working with it's resources.

## Cluster Test Suite Template

Status of a `ClusterTestSuite` contains many information.
Output from following command:
```
kubectl get cts -oyaml {suite mame}
```
can be overwhelming. To get more concise but still informative output, you can use template stored in `kubectl/suite.template`.
```
 kubectl get cts {suite name} -ogo-template-file --template={path to template file}
```

Output:
```
Name:           {suite name}
Concurrency:    1.0
MaxRetries:     0.0
Count:          1.0
StartTime:      2019-04-06T12:56:11Z
CompletionTime: 2019-04-06T13:22:38Z
Condition:      Failed
Tests:
    test-monitoring -
    test-events +
    test-ui ?
```

Next to the test names, status of the test executions is displayed
- `+` - test passed
- `-` - test failed
- `?` - test is in progress or not yet scheduled