# oomd

A `kubectl` plugin to display the pods and containers which have recently been `OOMKilled`.

## Installation 

Via [`krew`](https://krew.sigs.k8s.io/)
```
kubectl krew install oomd
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

You can specify another namespace, as you would with other `kubectl` commands.

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
