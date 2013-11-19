GOPATH=/usr/local/lib/go:${PWD}
export GOPATH
GO=go

STUBS := message replicainfo

.PHONY: stubs replica client fmt

all: replica client

replica:
	${GO} install replica

client:
	${GO} install client

gobin-codegen/src/bi/bi.go:
	hg clone https://code.google.com/p/gobin-codegen/

bin/bi: gobin-codegen/src/bi/bi.go
	cd gobin-codegen; ${MAKE}
	mkdir -p bin
	cp gobin-codegen/src/bi/bi bin/bi

stubs: bin/bi
	$(foreach stub, ${STUBS}, \
		bin/bi src/${stub}/${stub}.go > src/${stub}/${stub}_stub.go;)

fmt:
	cd src; ${GO} fmt *

clean:
	rm -rf bin/*
	rm `find . -iname "*_stub.go"`
