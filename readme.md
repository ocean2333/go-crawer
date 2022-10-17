# A crawer implemented in Golang

## Usage

### mac & linux
1. `git clone` 
2. `cd go-crawer/src`
3. `make all`
4. start etcd service on configured address
5. `cd bin && ./go-crawer`
6. use `cd bin && ./admin` to control the crawer now

### windows
1. `git clone` 
2. `cd go-crawer/src`
3. `.\build.ps1`
4. start etcd service on configured address
5. `cd bin`
6. use `./go-crawer` to start crawer service
7. use `cd bin && ./admin` to control the crawer now

## Features

### Use json to config the crawer

### Store album meatadata in etcd and store album images in local file system

### View albums in web page(TODO)

## code structure
```
//TODO
```

## TODO

### update readme.md
### develop a web service with react, show album/picture result
### add text parser and show on web