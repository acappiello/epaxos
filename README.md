ePaxos
======

An implementation of ePaxos in Go.
A research project by Alex Cappiello.

Current State
-------------

The core of the algorithm is mostly implemented, but it's a long way from being
useful. I'm not actively working on this and might not get back to it.

Not Yet Implemented
-------------------

Cleaning up old commands and replica recovery is not handled. Performance also
needs a lot of help.

Compilation
-----------

Instead of committing the generated marshal/unmarahal code, it is generated,
cloning gobin-codegen if not present. These must be explicitly regenerated if
the underlying code is changed. The stub generator doesn't import some things
that need to be, so there are patches for that.
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
