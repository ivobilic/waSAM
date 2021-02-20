#! /usr/bin/env sh
for f in $(ls); do scp $f/*.deb user@192.168.99.106:~/DEBIAN_PKGS/$f/main/; done
