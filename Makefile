.PHONY: install

clean:
	go version;\
	rm bin/*;\
	rm install/vars.sh

build: clean
	go run build.go --goos "$(goos)" --goarch "$(goarch)" --goarm "$(goarm)"

install:
	cd "install" && ./systemd-install.sh

uninstall:
	cd "install" && ./systemd-uninstall.sh

