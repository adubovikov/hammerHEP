NAME?=hammerhep

all:
	CGO_ENABLED=1 GOOS=linux go build -ldflags "-s -w" -o $(NAME)
    #go build -a -ldflags '-extldflags "-static"' -o $(NAME)

static:
	CGO_ENABLED=1 GOOS=linux CGO_LDFLAGS="-lm -ldl" go build -a -ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo -o $(NAME)
    #CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o $(NAME)
    #go build -a -ldflags '-extldflags "-static"' -o $(NAME)

debug:
	go build -o $(NAME)

modules:
	go get ./...

docker:
	./build_docker.sh

package:
	./build_package.sh

.PHONY: clean
clean:
	rm -fr $(NAME)
