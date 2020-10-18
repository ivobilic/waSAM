
USER_GH=eyedeekay
VERSION=0.32.31
packagename=sam3

echo:
	@echo "type make version to do release $(VERSION)"

version:
	gothub release -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(packagename) -t v$(VERSION) -d "version $(VERSION)"

del:
	gothub delete -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(packagename) -t v$(VERSION)

tar:
	tar --exclude .git \
		--exclude .go \
		--exclude bin \
		-cJvf ../$(packagename)_$(VERSION).orig.tar.xz .

copier:
	echo '#! /usr/bin/env sh' > deb/copy.sh
	echo 'for f in $$(ls); do scp $$f/*.deb user@192.168.99.106:~/DEBIAN_PKGS/$$f/main/; done' >> deb/copy.sh

fmt:
	find . -name '*.go' -exec gofmt -w -s {} \;
