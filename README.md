# Indy Test Suites tools

Indy test suites are used as a cli program to simulate several test cases which are used commonly to test a runnable prod-like indy for its functions, including metadata retrieving with stress, build with stress and promote tests.

## How to build

* Make sure you have installed golang v1.11+
* Make sure you have correctly set your $GOPATH
* Make sure $GO111MODULE is set to "on" if your go version is under v1.13
* Run `make build`
* The `indy-test` binary executable file will be generated in bin directory.

## How to use

After build please run ${repo}/build/indy-test and see command help info for futher usage.


