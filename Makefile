.PHONY:

pkg-list:
	go list -m all

pkg-clean:
	go clean
	go mod tidy

mod-clean:
	go clean -modcache