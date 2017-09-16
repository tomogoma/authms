.PHONY: clean build install uninstall

clean:
	go version
	rm -f bin/*
	rm -f install/vars.sh

build: clean
	go run build.go --goos "$(goos)" --goarch "$(goarch)" --goarm "$(goarm)"

install:
	$(MAKE) -C "install" install

uninstall:
	$(MAKE) -C "install" uninstall

