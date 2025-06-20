LINUX_PKGS := $(wildcard dist/*.deb) $(wildcard dist/*.rpm)
.PHONY: fury $(LINUX_PKGS)
fury: $(LINUX_PKGS)
$(LINUX_PKGS):
	fury push --as distribworks $@

PACKAGE_NAME          := github.com/distribworks/dkron
GOLANG_CROSS_VERSION  ?= v1.22

.PHONY: clean
clean:
	rm -f main
	rm -f *_SHA256SUMS
	rm -f dkron-*
	rm -rf build/*
	rm -rf builder/skel/*
	rm -f *.deb
	rm -f *.rpm
	rm -f *.tar.gz
	rm -rf tmp
	rm -rf ui-dist
	rm -rf ui/build
	rm -rf ui/node_modules
	GOBIN=`pwd` go clean -i ./builtin/...
	GOBIN=`pwd` go clean

.PHONY: web

web/node_modules: web/package.json
	cd web; bun install
	# touch the directory so Make understands it is up to date
	touch web/node_modules

internal/ui/ui-dist: web/node_modules web/public/* web/src/* web/src/*/*
	rm -rf internal/ui/ui-dist
	cd web; yarn build --out-dir ../internal/ui/ui-dist

proto: types/sinx.pb.go types/executor.pb.go types/pro.pb.go

types/%.pb.go: proto/%.proto
	protoc -I proto/ --go_out=types --go_opt=paths=source_relative --go-grpc_out=types --go-grpc_opt=paths=source_relative $<

web: internal/ui/ui-dist

main: internal/ui/ui-dist types/sinx.pb.go types/executor.pb.go *.go */*.go */*/*.go */*/*/*.go
	GOBIN=`pwd` go install ./builtin/...
	go mod tidy
	go build main.go
