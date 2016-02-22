#!/bin/sh
tar xzf test-modules.tgz
go test -v
rm -rf test-modules
