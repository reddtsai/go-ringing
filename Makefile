.PHONY:

pkg-list:
	go list -m all

pkg-clean:
	go mod tidy