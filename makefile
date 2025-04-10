BUILD_TAGS?=babble

# vendor uses Glide to install all the Go dependencies in vendor/
vendor:
	(rm glide.lock || rm -rf vendor ) && glide install

# install compiles and places the binary in GOPATH/bin
install:
	go install --ldflags '-extldflags "-static"' \
		--ldflags "-X github.com/BOTCoinNetwork/babble/src/version.GitCommit=`git rev-parse HEAD`" \
		./cmd/babble

# build compiles and places the binary in /build
build:
	CGO_ENABLED=0 go build \
		--ldflags "-X github.com/BOTCoinNetwork/babble/src/version.GitCommit=`git rev-parse HEAD`" \
		-o build/babble ./cmd/babble/

# dist builds binaries for all platforms and packages them for distribution
dist:
	@BUILD_TAGS='$(BUILD_TAGS)' sh -c "'$(CURDIR)/scripts/dist.sh'"

# dist builds aar for mobile android
mobile-dist:
	@BUILD_TAGS='$(BUILD_TAGS)' sh -c "'$(CURDIR)/scripts/dist_mobile.sh'"

tests:  test

test:
	glide novendor | xargs go test -count=1 -tags=unit

flagtest:
	glide novendor | xargs go test -count=1 -run TestFlagEmpty

extratests:
	glide novendor | xargs go test -count=1 -run Extra

alltests:
	glide novendor | xargs go test -count=1 
 
lint:
	glide novendor | xargs golint

.PHONY: vendor install build dist test flagtest extratests alltests tests mobile-dist
