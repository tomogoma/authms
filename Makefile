.PHONY: clean build install uninstall

clean:
	go version
	rm -f bin/*
	rm -f install/vars.sh
	rm -rf install/docs
	rm -rf cmd/gcloud/conf/docs

build: clean
	go run build.go --goos "$(goos)" --goarch "$(goarch)" --goarm "$(goarm)"

install:
	$(MAKE) -C "install" install

uninstall:
	$(MAKE) -C "install" uninstall

