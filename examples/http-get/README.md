# http-get

This is an example of how you can use `viassh` to create an HTTP client that jumps via one or more hosts before creating a connection from the last host in the chain.

## Build

```shell
go build
```

## Run

```shell
./http-get -via user@firsthost.com:22 -via user@secondhost.com:22 -target https://news.ycombinator.com/
```
