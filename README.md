# self-update

A program that can perform self-update by starting a binary with a new version.

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
3. Shutdown http server
4. Call `GET /replace` on upgrade binary's temporary server.
5. Exit

From the new service perspective:

1. Start temporary server with `/replace` endpoint
2. Wait for `GET /replace`
3. Start the proper http server

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

1. Third httpd

The service with the newer version would use it to announce success on binding to the target port, allowing the former to roll back.

2. Signals

Similarly to the first solution, one service would signal its state to the other. However, it would be compatible with only UNIX systems.

3. Auto repair

The service with the older version would start the upgraded service and then test its endpoints.

On failure, the service would be killed and the proper httpd restored.

### Security

Mechanism based on a trust in human operators cannot be considered safe.

A proper way to secure this application would involve signing the binaries.

Windows and macOS have their own built-in signature mechanisms.

For Linux, this is way more complex. There is no one solution.

Usually, applications are installed by a package manager which ensures that the repository delivers safe binaries.

A quick solution for this particular case would be using GPG signatures, but any public key cryptography solution would be applicable.
