GOPATH=${PWD}
export GOPATH
GO=go

MARSHAL := message replicainfo

.PHONY: marshal replica client fmt

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
		bin/bi src/${marshal}/${marshal}.go > \
			src/${marshal}/${marshal}_marshal.go; )
	patch src/message/message_marshal.go src/message/marshal.patch

fmt:
	cd src; ${GO} fmt *

clean:
	rm -rf bin/* pkg/*
	$(foreach marshal, ${MARSHAL}, \
		rm -f src/${marshal}/${marshal}_marshal.go; )
	rm -f `find -regex ".+~\|.+/#.+#"`
