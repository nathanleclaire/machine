.PHONY: test validate-dco validate-gofmt

default: build

test:
	script/test

validate-dco:
	script/validate-dco

validate-gofmt:
	script/validate-gofmt

validate: validate-dco validate-gofmt test

build: clean
	script/build

remote: clean
	script/build-remote

rmi:
	docker rmi docker-machine-build

clean:
	rm -f docker-machine_*
	rm -rf Godeps/_workspace/pkg
