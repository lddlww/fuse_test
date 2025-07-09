% podman-generate-kube(1)
## NAME
podman-generate-kube - Generate Kubernetes YAML based on a pod or container

## SYNOPSIS
**podman generate kube** [*options*] *container* | *pod*

## DESCRIPTION
**podman generate kube** will generate Kubernetes Pod YAML (v1 specification) from a Podman container or pod. Whether
the input is for a container or pod, Podman will always generate the specification as a Pod. The input may be in the form
of a pod or container name or ID.

Note that the generated Kubernetes YAML file can be used to re-run the deployment via podman-play-kube(1).

## OPTIONS:

**--filename**, **-f**=**filename**

Output to the given file, instead of STDOUT. If the file already exists, `generate kube` will refuse to replace it and return an error.

**--service**, **-s**

Generate a Kubernetes service object in addition to the Pods. Used to generate a Service specification for the corresponding Pod output. In particular, if the object has portmap bindings, the service specification will include a NodePort declaration to expose the service. A
random port is assigned by Podman in the specification.

## Examples

Create Kubernetes Pod YAML for a container called `some-mariadb` .
```
$ sudo podman generate kube some-mariadb
# Generation of Kubenetes YAML is still under development!
#
# Save the output of this file and use kubectl create -f to import
# it into Kubernetes.
#
# Created with podman-0.11.2-dev
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: 2018-12-03T19:07:59Z
  labels:
    app: some-mariadb
  name: some-mariadb-libpod
spec:
  containers:
  - command:
    - docker-entrypoint.sh
    - mysqld
    env:
    - name: PATH
      value: /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    - name: TERM
      value: xterm
    - name: HOSTNAME
    - name: container
      value: podman
    - name: GOSU_VERSION
      value: "1.10"
    - name: GPG_KEYS
      value: "199369E5404BD5FC7D2FE43BCBCB082A1BB943DB \t177F4010FE56CA3336300305F1656F24C74CD1D8
        \t430BDF5C56E7C94E848EE60C1C4CBDCDCD2EFD2A \t4D1BB29D63D98E422B2113B19334A25F8507EFA5"
    - name: MARIADB_MAJOR
      value: "10.3"
    - name: MARIADB_VERSION
      value: 1:10.3.10+maria~bionic
    - name: MYSQL_ROOT_PASSWORD
      value: x
    image: quay.io/baude/demodb:latest
    name: some-mariadb
    ports:
    - containerPort: 3306
      hostPort: 36533
      protocol: TCP
    resources: {}
    securityContext:
      allowPrivilegeEscalation: true
      privileged: false
      readOnlyRootFilesystem: false
    tty: true
    workingDir: /
status: {}
```

Create Kubernetes Pod YAML for a pod called `demoweb` and include a service.
```
$ sudo podman generate kube -s demoweb
# Generation of Kubernetes YAML is still under development!
#
# Save the output of this file and use kubectl create -f to import
# it into Kubernetes.
#
# Created with podman-0.12.2-dev
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: 2018-12-18T15:16:06Z
  labels:
    app: demoweb
  name: demoweb-libpod
spec:
  containers:
  - command:
    - python3
    - /root/code/graph.py
    env:
    - name: PATH
      value: /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    - name: TERM
      value: xterm
    - name: HOSTNAME
    - name: container
      value: podman
    image: quay.io/baude/demoweb:latest
    name: practicalarchimedes
    resources: {}
    securityContext:
      allowPrivilegeEscalation: true
      capabilities: {}
      privileged: false
      readOnlyRootFilesystem: false
    tty: true
    workingDir: /root/code
status: {}
---
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: 2018-12-18T15:16:06Z
  labels:
    app: demoweb
  name: demoweb-libpod
spec:
  ports:
  - name: "8050"
    nodePort: 31269
    port: 8050
    protocol: TCP
    targetPort: 0
  selector:
    app: demoweb
  type: NodePort
status:
  loadBalancer: {}
```

## SEE ALSO
podman(1), podman-container(1), podman-pod(1), podman-play-kube(1)

## HISTORY
December 2018, Originally compiled by Brent Baude (bbaude at redhat dot com)
