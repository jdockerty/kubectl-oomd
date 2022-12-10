# oomd

A `kubectl` plugin to display the pods and containers which have recently been `OOMKilled`.

## Installation 

Via [`krew`](https://krew.sigs.k8s.io/)
```
kubectl krew install oomd
kubectl oomd
```

Or manually from the [releases](https://github.com/jdockerty/kubectl-oomd/releases) page, for example with `linux_amd64`

```shell
# Download the release for your OS
wget https://github.com/jdockerty/kubectl-oomd/releases/download/v0.0.7/oomd_linux_amd64.tar.gz

# Extract files from the archive
tar -xvf oomd_linux_amd64.tar.gz

# Move the binary into your PATH, renaming the application to 'kubectl-oomd' to
# satisfy the plugin convention
mv oomd "${HOME}"/.local/bin/kubectl-oomd

# Run the plugin
kubectl oomd
```

## Usage

Running the command will display the pods that have recently been `OOMKilled` in your current namespace.
This also shows the specific container which was killed too, helpful in the case of multi-container pods.


```
kubectl oomd

POD                        CONTAINER        REQUEST     LIMIT     TERMINATION TIME
my-app-5bcbcdf97-722jp     infoapp          1G          8G        2022-11-07 13:03:49 +0000 GMT
my-app-5bcbcdf97-7j5rd     infoapp          1G          8G        2022-11-07 14:35:34 +0000 GMT
my-app-5bcbcdf97-k8g8g     infoapp          1G          8G        2022-11-07 14:35:02 +0000 GMT
my-app-5bcbcdf97-mf65j     infoapp          1G          8G        2022-11-07 14:34:57 +0000 GMT
```

You can specify another namespace, as you would with other `kubectl` commands or use `--all-namespaces`/`-A` to check against them all.


```
kubectl oomd -n oomkilled

POD                        CONTAINER        REQUEST     LIMIT     TERMINATION TIME
my-app-5bcbcdf97-722jp     infoapp          1G          8G        2022-11-07 13:03:49 +0000 GMT
my-app-5bcbcdf97-7j5rd     infoapp          1G          8G        2022-11-07 14:35:34 +0000 GMT
my-app-5bcbcdf97-k8g8g     infoapp          1G          8G        2022-11-07 14:35:02 +0000 GMT
my-app-5bcbcdf97-mf65j     infoapp          1G          8G        2022-11-07 14:34:57 +0000 GMT
```

```
kubectl oomd --no-headers

my-app-5bcbcdf97-722jp     infoapp          1G          8G        2022-11-07 13:03:49 +0000 GMT
my-app-5bcbcdf97-7j5rd     infoapp          1G          8G        2022-11-07 14:35:34 +0000 GMT
my-app-5bcbcdf97-k8g8g     infoapp          1G          8G        2022-11-07 14:35:02 +0000 GMT
my-app-5bcbcdf97-mf65j     infoapp          1G          8G        2022-11-07 14:34:57 +0000 GMT
```

Experimental sorting is enabled through the `--sort-field` flag. By default, this is `none`.
At the moment, only `time` is supported which sorts by termination time of containers, this is mainly
useful in larger outputs across all namespaces (`-A`), used in conjunction with a pipe to `tail`.
A simple example of this in action is shown below.

```
# The default with no sorting.
kubectl oomd -n tracing
POD                    CONTAINER        REQUEST     LIMIT     TERMINATION TIME
jaeger-agent-4k845     jaeger-agent     100Mi       100Mi     2022-11-11 21:06:31 +0000 GMT
jaeger-agent-j5vb8     jaeger-agent     100Mi       100Mi     2022-11-09 23:20:38 +0000 GMT

# Most recently OOMKilled pods are shown first
kubectl oomd -n tracing --sort-field time
POD                    CONTAINER        REQUEST     LIMIT     TERMINATION TIME
jaeger-agent-j5vb8     jaeger-agent     100Mi       100Mi     2022-11-09 23:20:38 +0000 GMT
jaeger-agent-4k845     jaeger-agent     100Mi       100Mi     2022-11-11 21:06:31 +0000 GMT
```

### Development

If you wish to force some `OOMKilled` pods for testing purposes, you can use [`oomer`](https://github.com/jdockerty/oomer)
or simply run

    kubectl apply -f https://raw.githubusercontent.com/jdockerty/oomer/main/oomer.yaml

This will create the `oomkilled` namespace and a `Deployment` with pods that continually exit with code `137`,
in order to be picked up by `oomd`.
