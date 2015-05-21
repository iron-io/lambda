# How to make a build for Windows

### 1\. Build host version of go

First step is to build host version of go:

```
$ cd $GOROOT/src
$ ./make.bash
```

Next you need to build the rest of go compilers and linkers. There is a small program to do that:

```
$ cat ~/bin/buildcmd
#!/bin/sh
set -e
for arch in 8 6; do
        for cmd in a c g l; do
                go tool dist install -v cmd/$arch$cmd
        done
done
exit 0
```

Last step is to build windows versions of standard commands and libraries. And small script for that too:

```
$ cat ~/bin/buildpkg
#!/bin/sh
if [ -z "$1" ]; then
        echo 'GOOS is not specified' 1>&2
        exit 2
else
        export GOOS=$1
        if [ "$GOOS" = "windows" ]; then
                export CGO_ENABLED=0
        fi
fi
shift
if [ -n "$1" ]; then
        export GOARCH=$1
fi
cd $GOROOT/src
go tool dist install -v pkg/runtime
go install -v -a std
```

You can run it like that:

```
$ ~/bin/buildpkg windows 386
```

to build windows/386 version of Go commands and packages. You can, probably, see it from script.

Now we're ready to build our windows executable:

```
$ GOOS=windows GOARCH=amd64 go build -o iron.exe ./*.go
```

We just need to find Windows computer to run our iron.exe.

[source](https://code.google.com/p/go-wiki/wiki/WindowsCrossCompiling)