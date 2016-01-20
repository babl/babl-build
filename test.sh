#!/bin/sh
tar xzf test-modules.tgz
go test
rm -rf test-modules
