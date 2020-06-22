# self-update

A program that performs self-update by executing supplied upgrade binary.

## Build

Run `make build`. It will produce `self-update` binary in `dist/` directory.

To run the app, you may find `make run` handy; it compiles and runs the application with `-dev` flag.

### Windows

To cross-compile for Windows, run `make build-windows`.

### Linter

Run `make lint`.

### Test

Run `make test`.


## Usage

For the simplest setup and most verbose output, run `./self-update -dev`, then navigate to [localhost:8080](http://localhost:8080).

### Flags

- `-bind` specifies hostname and port on which the service will bind itself
- `-dev` formats logs in human-readable form and shows debug logs
- `-upgrade` is used solely by the upgrade mechanism and should not be used by end-users
- `-upgrade-bind` specifies hostname and port on which the service will temporarily bind itself during upgrade process
- `-upgrade-dir` specifies the directory where the service will look for binaries which will be used in the upgrade process
- `-version` prints version

## Architecture

The service versioning is based on [Semantic Versioning](http://semver.org).

### Upgrade candidates

To make a list of upgrade candidates, the application:

1. Scans `upgrade-dir` for executables
2. Calls `<executable> -version` on each
3. The latest version is chosen from the collection of executable-version pairs

### Upgrade

From the old service perspective:

1. Get the latest upgrade candidate
2. Execute `<upgrade binary> -upgrade true -upgrade-bind <value passed> -bind <value passed>`
3. Shutdown HTTP server
4. Call `GET /replace` on upgrade binary's temporary server.
5. Exit

From the new service perspective:

1. Start temporary server with `/replace` endpoint
2. Wait for `GET /replace`
3. Start the proper HTTP server

Keep in mind that this upgrade process is far from perfection (see [Known issues](#known-issues)).

### Security

The upgrade mechanism bases on local storage which is assumed to be safe.

It is the operator's duty to supply the service with trusted binaries.

This would be unacceptable on a production environment. See [Known issues](#known-issues).

## Known issues

### The state transition isn't well-defined and has no rollbacks

The mechanism responsible for the upgrade should be a nondeterministic finite automaton.

```
          Failure
    +------------------+
    v                  |
+---+---+        +-----+-----+           +-------+
|       |        |           |  Success  |       |
|  Run  |        |  Upgrade  +---------->+  Die  |
|       |        |           |           |       |
+---+---+        +-----+-----+           +-------+
    |                  ^
    +------------------+
        GET /replace
```

Although, given the limited time for this project, it couldn't be done.

### The service may fail to upgrade

The service that gets upgraded calls `GET /replace` and upon HTTP 200, dies.

Obviously, status 200 does not guarantee the success of upgrade procedure.

This can be solved in a couple of ways:

1. Third HTTP server

The service with the newer version would use it to announce success on binding to the target port, allowing the former to roll back.

2. Signals

Similarly to the first solution, one service would signal its state to the other. However, it would be compatible with only UNIX systems.

3. Auto repair

The service with the older version would start the upgraded service and then test its endpoints.

On failure, the service would be killed and the proper HTTP server restored.

### Sleep-based synchronisation

The Javascript on the client's side waits N seconds before it redirects the browser to `/`.

This should be done with a periodical HTTP call or a websocket to ensure that the upgrade is done.

Also, the upgrade process itself waits for N seconds when it starts the new instance to ensure it had enough time to prepare for `/replace` call.

This would mitigated with any of [those solutions](#the-service-may-fail-to-upgrade).

### Security

Mechanism based on a trust in human operators cannot be considered safe.

A proper way to secure this application would involve signing the binaries.

Windows and macOS have their own built-in signature mechanisms.

For Linux, this is way more complex. There is no one solution.

Usually, applications are installed by a package manager which ensures that the repository delivers safe binaries.

A quick solution for this particular case would be using GPG signatures, but any public key cryptography solution would be applicable.

### Git history

This repository's history would not allow usage of `git bisect` which would be unacceptable in a real life project.

### Tests

There are no real tests in this repository that would check upgrade's behaviour.

Those tests can't be unit tests, but acceptance tests.

I would use PyTest for that.


### Process supervision

There is no obvious way to monitor and collect logs from the upgraded server.

One of the correct solutions on UNIX systems would be to use system init instead of just spawning a new process.