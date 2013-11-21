GOPATH=/usr/local/lib/go:${PWD}
export GOPATH
GO=go

MARSHAL := message replicainfo

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

marshal: bin/bi
	$(foreach marshal, ${MARSHAL}, \
		bin/bi src/${marshal}/${marshal}.go > src/${marshal}/${marshal}_stub.go;)

fmt:
	cd src; ${GO} fmt *

clean:
	rm -rf bin/*
	rm `find . -iname "*_marshal.go"`
