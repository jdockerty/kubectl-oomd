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

POD                         CONTAINER        TERMINATION TIME
my-app-5bcbcdf97-722jp      infoapp          2022-11-04 10:51:48 +0000 GMT
my-app-5bcbcdf97-k52lg      infoapp          2022-11-04 10:51:48 +0000 GMT
my-app-5bcbcdf97-v9ff6      infoapp          2022-11-04 10:51:48 +0000 GMT
```

You can specify another namespace, as you would with other `kubectl` commands.

```
kubectl oomd -n oomkilled

POD                         CONTAINER        TERMINATION TIME
my-app-5bcbcdf97-722jp      infoapp          2022-11-04 10:51:48 +0000 GMT
my-app-5bcbcdf97-k52lg      infoapp          2022-11-04 10:51:48 +0000 GMT
my-app-5bcbcdf97-v9ff6      infoapp          2022-11-04 10:51:48 +0000 GMT
```

```
kubectl oomd --no-headers

my-app-5bcbcdf97-722jp      infoapp          2022-11-04 10:51:48 +0000 GMT
my-app-5bcbcdf97-k52lg      infoapp          2022-11-04 10:51:48 +0000 GMT
my-app-5bcbcdf97-v9ff6      infoapp          2022-11-04 10:51:48 +0000 GMT

```
