
## Usage
The following assumes you have the plugin installed via

```shell
kubectl krew install oomlie
```

Once installed, simply run 

```shell
kubectl oomlie
```

This will display the pods which have been `OOMKilled`, if there are no pods which meet this requirement, then there will be no output.

## How it works

This simply checks the current or provided namespace, using your current context, for the pod/container statuses.
As the exit code of the `OOMKilled` workloads if `137`, we can perform a simple check for this and collect all of the pods which
meet this condition.
