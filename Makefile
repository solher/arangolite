
#%# TEST_ARGS         : Specify tests to run.
TEST_ARGS=''
#%#                   : Extra arguments to pass to 'find' when chosing go-files to run gofmt on
GOFMT_ARGS=

default: help

.PHONY: test
test         : ##Runs the unit tests. Run only specific tests by setting the TEST_ARGS variable
	go test -v -run $(TEST_ARGS)

.PHONY: check_go_fmt
check_go_fmt : ##Runs gofmt on go files and print the diff if any style iconsistencies are found
	$(eval export GOFILES_TO_CHECK=$(shell find . -type f -name '*.go' $(GOFMT_ARGS)))
	@gofmt -e -d ${GOFILES_TO_CHECK}
	@gofmt -l ${GOFILES_TO_CHECK}

.PHONY: help
help         : ##Show this help
	@echo "-----------------------------------------------------------------"
	@echo "  Available make targets"
	@echo "-----------------------------------------------------------------"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -rn "s/(\S+)\s+:+?\s##(.*)/\1 \"\2\"/p" | xargs printf " %-20s    : %5s\n"
	@echo
	@echo "-----------------------------------------------------------------"
	@echo "  Available environment configurations to set with 'make [VARIABLE=value] <target>'"
	@echo "-----------------------------------------------------------------"
	@fgrep -h "#%#" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -rn "s/#%#\s+(\S+)\s+:+?\s+(.*)/\1 \"\2\"/p" | xargs printf " %-20s    : %5s\n"
	@echo
	@echo " Example running a specific test:\n  make test TEST_ARGS=TestIsErrUnique"
	@echo
