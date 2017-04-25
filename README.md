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

