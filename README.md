[![Build Status](https://travis-ci.com/axetroy/s4.svg?branch=master)](https://travis-ci.com/axetroy/s4)
[![Coverage Status](https://coveralls.io/repos/github/axetroy/s4/badge.svg?branch=master)](https://coveralls.io/github/axetroy/s4?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/axetroy/s4)](https://goreportcard.com/report/github.com/axetroy/s4)
![License](https://img.shields.io/github/license/axetroy/s4.svg)
![Repo Size](https://img.shields.io/github/repo-size/axetroy/s4.svg)

## s4

Perform remote server tasks on local computer

Features:

- [x] Declarative workflow
- [x] Upload local files to remote
- [x] Download remote files to local
- [x] Execute commands on the remote server

### Usage

step 1: create a file name `.s4`

```s4
HOST 192.168.0.1 # remote SSH server address

PORT 2222 # remote SSH server port

USERNAME axetroy # remote SSH server username

CWD /home/axetroy # set the current work dir to '/home/axetroy'

UPLOAD star.py README.md ./server # upload files to remote server `/home/axetroy/server`

RUN python ./server/start.py # execute command on remote server
```

step 2: run the following command

```bash
> s4
```

for more detail about command. print `s4 --help`

### Documentation

| Syntax   | Description                                                              | Multiple | Example                      |
| -------- | ------------------------------------------------------------------------ | -------- | ---------------------------- |
| HOST     | remote SSH server address                                                | ✖️       | HOST 192.168.0.1             |
| PORT     | remote SSH server port                                                   | ✖️       | PORT 2022                    |
| USERNAME | remote SSH server username                                               | ✖️       | USERNAME axetroy             |
| ENV      | set environmental variable for `RUN` command                             | ☑️       | ENV PRIVATE_KEY = 123        |
| CWD      | set current work dir for remote server                                   | ☑️       | CWD /home/axetroy            |
| UPLOAD   | upload local files to remote server                                      | ☑️       | UPLOAD start.py ./server     |
| DOWNLOAD | download remote files to local                                           | ☑️       | DOWNLOAD start.py ./server   |
| COPY     | copy file on remote server                                               | ☑️       | COPY data.db data.db.bak     |
| MOVE     | move file on remote server                                               | ☑️       | MOVE data.bak data.db        |
| DELETE   | delete files on remote server, for security, directory cannot be removed | ☑️       | DELETE file1 file2           |
| RUN      | run command in remote command                                            | ☑️       | RUN python ./server/start.py |

### Build from source code

```bash
> go get -v -u github.com/axetroy/s4
> cd $GOPATH/src/github.com/axetroy/s4
> make build
> ls -lh ./bin

total 85976
-rwxr-xr-x  1 axetroy  staff   7.1M Aug 26 20:19 linux_x64_s4
-rwxr-xr-x  1 axetroy  staff   6.4M Aug 26 20:19 linux_x86_s4
-rwxr-xr-x  1 axetroy  staff   7.0M Aug 26 20:19 osx_x64_s4
-rwxr-xr-x  1 axetroy  staff   6.3M Aug 26 20:19 osx_x86_s4
-rwxr-xr-x  1 axetroy  staff   6.9M Aug 26 20:19 windows_x64_s4.exe
-rwxr-xr-x  1 axetroy  staff   6.2M Aug 26 20:19 windows_x86_s4.exe
```

### Test

```bash
make test
```

### License

The [MIT License](https://github.com/axetroy/s4/blob/master/LICENSE)
