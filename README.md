# Werify: PoC with a cheesy name #

Werify is a Go stdlib-only PoC attempt at distributed host checks.

## Setup ##

### Installation ###

With a [correctly configured](https://golang.org/doc/code.html#GOPATH) Go installation:

    go get -u github.com/disq/werify/cmd/werifyctl
    go get -u github.com/disq/werify/cmd/werifyd

This will install `werifyctl` and `werifyd` in your `$GOPATH/bin` directory.

### Build (Developers) ###

    git clone https://github.com/disq/werify
    cd werify
    make

This will build `werifyctl` and `werifyd`.

## Concepts ##

- `werifyd` is the server, listening on `TCP/30035`.
- Each `werifyd` stores a list of servers (like itself) in memory. 
- `werifyctl` is used to talk to one of the servers, usually the local one.
- In the course of the `werifyctl` run, the destination server becomes a coordinator node. Server list is edited using `werifyctl`.
- One or more operations can be ran on the server list of the coordinator node. `werifyctl` is a dummy RPC client.
- Separate `werifyd` instances can have separate lists of servers -- but none other than the coordinator node needs to have the list.
- The coordinator node will hand out work (to run and/or forward to servers) in the same env tag, collect results, and pass them to the calling `werifyctl`.

## Caveats ##

- If the coordinator node is one of the destination hosts, it should also be added to the server list (`./werifyctl add 127.0.0.1`)
- Each host is automatically identified by its first-referred `ip[:port]` pair. Adding a single host to multiple coordinators using different referral schemes (internal vs. external ip, forwarded port, etc.) is not supported.

### Persistent Server List ###

Persistent server list is not implemented directly. But after launching `werifyd`, `werifyctl` can be used to populate the list using a commands-file:

    cat examples/init.werifyd | ./werifyctl -

## Invocation ##

werifyd is the server, and werifyctl is the client. Work is always done on the server.

### Server ###

```
Usage of ./werifyd:
  -env string
        Env tag (default "dev")
  -port int
        Listen on port (default 30035)
  -w int
        Number of workers per operation (default runtime.NumCPU)
```

- `env` is the environment tag. It should match exactly on all `werifyd`/`werifyctl` instances and it is enforced on every RPC call.
- Number of workers (`-w`) applies to every worker-pool related event. The daemon utilizes multiple worker pools.

### Client ###

```
Usage: ./werifyctl [OPTION]... [COMMAND [PARAMS...]]

Available options:
  -connect string
        Connect to werifyd (default "localhost:30035")
  -env string
        Env tag (default "dev")
  -timeout duration
        Connect timeout (default 10s)

Available commands:
              add  Adds a host to werifyd
              del  Removes a host from werifyd
             list  Lists hosts in werifyd
       listactive  Lists active hosts in werifyd
     listinactive  Lists inactive hosts in werifyd
        operation  Runs operations from file on werifyd
              get  Get status of operation with handle
          refresh  Start health check on all hosts

Commands can also be specified from stdin using "-".
```

- `connect` parameter can be set with the environment variable `WERIFY_CONNECT`.
- `env` parameter can be set with the environment variable `WERIFY_ENV`.

## Operations File Format ##

Host checks/operation file is a JSON file. The first-level object keys are user specified. There isn't any imposed limit on the number of checks.

```
{
    "check_name": {
        "type": "type of check",
        "path": "value for the path parameter for check_name",
        "check": "value for the check parameter for check_name"
    },
    "name_for_another_check": {
        "type": "type of check",
        "path": "value for the path parameter for name_for_another_check",
        "check": "value for the check parameter for name_for_another_check"
    },
    ...
}
```

In the response, each check will be referred to by its key name and the Host's first-referred identifier. (See [Caveats](https://github.com/disq/werify#caveats))

A sample `ops.json` file is provided in the [examples](https://github.com/disq/werify/tree/master/examples) directory.

## Types of Host Checks ##

### File exists ###

Checks if the file or directory exists on the host system.

Parameters:
- `type`: Should be set to `file_exists`
- `path`: Full path to the file to check

### File contains ###

A type of line-by-line grep.

Parameters:
- `type`: Should be set to `file_contains`
- `path`: Full path to the file to check
- `check`: Text/contents to check

### Running Process ###

Checks if the given process is running on the host system. Linux (`/proc` filesystem) only.

Parameters:
- `type`: Should be set to `process_running`
- `path`: Full path to the process to check, ie. `/bin/bash`
- `check`: Basename of the process to check, ie. `bash`

At least one of `path` or `check` should be supplied.


## Example Run ##

Launch your daemon: (or multiple daemons on multiple hosts)
```
./werifyd
```

Choose one of the hosts (well...) as the coordinator node. This node will have a list of all the other nodes. For the sake of brevity, let's choose the local one.

Add each ip address, like so:
```
./werifyctl -connect 127.0.0.1 add 10.42.0.3
./werifyctl -connect 127.0.0.1 add 10.42.0.4
./werifyctl -connect 127.0.0.1 add 127.0.0.1 # Let the coordinator node run the checks as the others
```


You can also pipe the commands through stdin:

```
cat examples/init.werifyd | ./werifyctl -
```


You can now check the list, using the `werifyctl list` command:

```
Active hosts (3)
10.42.0.3:30035
10.42.0.4:30035
127.0.0.1:30035
Inactive hosts (0)
End of list
```

To run a host check operation, prepare an operations file. Then use the `werifyctl operation` command:
```
./werifyctl operation examples/ops.json
```

The output will be:
```
Operation submitted. To check progress, run: ./werifyctl get asv1
```
(`asv1` is a unique handle for the submitted operation)

You can then check the progress using `werifyctl get`:
```
./werifyctl get asv1
Host:10.42.0.3:30035 Operation:check_virus_file_exists Success:false
Host:10.42.0.3:30035 Operation:check_etc_hosts_has_4488 Success:false
Host:10.42.0.4:30035 Operation:check_virus_file_exists Success:false
Host:10.42.0.4:30035 Operation:check_etc_hosts_has_4488 Success:false
Host:127.0.0.1:30035 Operation:check_virus_file_exists Success:false
Host:127.0.0.1:30035 Operation:check_etc_hosts_has_4488 Success:false
Operation ended, took 3.445244ms
```
