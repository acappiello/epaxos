GOPATH=/usr/local/lib/go:${PWD}
export GOPATH
GO=go

MARSHAL := message replicainfo

.PHONY: marshal replica client fmt

all: replica client

replica: src/mapset/set.go
	${GO} install replica

client:
	${GO} install client

gobin-codegen/src/bi/bi.go:
	hg clone https://code.google.com/p/gobin-codegen/

src/mapset/set.go:
	git clone https://github.com/deckarep/golang-set src/mapset

bin/bi: gobin-codegen/src/bi/bi.go
	cd gobin-codegen; ${MAKE}
	mkdir -p bin
	cp gobin-codegen/src/bi/bi bin/bi

marshal: bin/bi
	$(foreach marshal, ${MARSHAL}, \
		bin/bi src/${marshal}/${marshal}.go > \
			src/${marshal}/${marshal}_marshal.go; )

fmt:
	cd src; ${GO} fmt *

clean:
	rm -rf bin/*
	rm `find -regex ".+~\|.+/#.+#"`
	$(foreach marshal, ${MARSHAL}, \
		rm src/${marshal}/${marshal}_marshal.go; )
