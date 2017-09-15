.PHONY: install

build:
	go build -o bin/app

install:
	cd "install" && ./systemd-install.sh

uninstall:
	cd "install" && ./systemd-uninstall.sh

