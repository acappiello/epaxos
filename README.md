ePaxos
======

An implementation of ePaxos in Go.
A senior thesis research project by Alex Cappiello.

Current State
-------------

Dependencies are added upon receiving a new request, but they are not
subsequently handled yet.

Not Yet Implemented
-------------------

A lot...

Compilation
-----------

Instead of committing the generated marshal/unmarahal code, it is generated,
cloning gobin-codegen if not present. These must be explicitly regenerated if
the underlying code is changed. There is currently a bug such that "commands"
needs to be added to the imports of message_marshal.go
```
make marshal
```
The rest is straightforward:
```
make
```

Running
-------

Currently, one replica briefly acts as a master until initial connections are
made (I'm considering just changing this to read a file of a list of hosts
though). All other nodes are told to connect to this node on startup via CLI.

For example:
```
cd bin
./replica -p 5000
./replica -p 5001 -h localhost -np 5000
./replcia -p 5002 -h localhost -np 5000
./client -p 5000
```

External Libraries
------------------

* Stub Generator: https://code.google.com/p/gobin-codegen/
* Set Type: https://github.com/deckarep/golang-set
