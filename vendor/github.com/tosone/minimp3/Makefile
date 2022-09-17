test: clean
	go test

benchmark:
	go test -bench=.

coverage: clean
	go test -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean:
	$(RM) *.out *.html *.pcm
