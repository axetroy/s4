[![Build Status](https://travis-ci.com/axetroy/s4.svg?branch=master)](https://travis-ci.com/axetroy/s4)
[![Coverage Status](https://coveralls.io/repos/github/axetroy/s4/badge.svg?branch=master)](https://coveralls.io/github/axetroy/s4?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/axetroy/s4)](https://goreportcard.com/report/github.com/axetroy/s4)
![License](https://img.shields.io/github/license/axetroy/s4.svg)
![Repo Size](https://img.shields.io/github/repo-size/axetroy/s4.svg)

## s4

Perform remote server tasks on local computer

Features:

- [x] Cross platform support
- [x] Declarative workflow
- [x] Upload local files to remote
- [x] Download remote files to local
- [x] Execute commands on the remote server

### Usage

step 1: create a file name `.s4`

```s4
CONNECT root@192.168.0.1:22

RUN ls -lh
```

step 2: run the following command

```bash
> s4
[Step 1]: CONNECT root@192.168.0.1:22
? Please type remote server's password **********
[Step 2]: RUN ls -lh
total 20K
drwxr-xr-x  4 root root 4.0K Mar 15 10:10 test1
drwxr-xr-x  2 root root 4.0K Sep 23  2018 test2
drwxr-xr-x  6 root root 4.0K Sep 23  2018 test3
drwxr-xr-x  4 root root 4.0K Aug 27 16:25 test4
```

for more detail about command. print `s4 --help`

### Documentation

| Syntax   | Description                                              | Multiple | Example                                                |
| -------- | -------------------------------------------------------- | -------- | ------------------------------------------------------ |
| CONNECT  | connect to remote SSH server                             | ✖️       | `CONNECT root@192.168.0.1:22`                          |
| ENV      | set environmental variable for `RUN` command             | ☑️       | `ENV PRIVATE_KEY = 123`                                |
| VAR      | defining variables. it can use in anywhere               | ☑️       | `VAR PRIVATE_KEY = 123`<br/>`RUN echo {{PRIVATE_KEY}}` |
| CD       | change current working directory of remote server        | ☑️       | `CD /home/axetroy`                                     |
| UPLOAD   | upload local files to remote server                      | ☑️       | `UPLOAD start.py ./server`                             |
| DOWNLOAD | download remote files to local                           | ☑️       | `DOWNLOAD start.py ./server`                           |
| COPY     | copy file at remote server                               | ☑️       | `COPY data.db data.db.bak`                             |
| MOVE     | move file at remote server                               | ☑️       | `MOVE data.bak data.db`                                |
| DELETE   | delete files at remote server, directory will be ignored | ☑️       | `DELETE file1 file2`                                   |
| RUN      | run command at remote server                             | ☑️       | `RUN python ./remote/start.py`                         |
| BASH     | run bash script in local server                          | ☑️       | `BASH cat package.json \| grep version`                |
| CMD      | run command in local server                              | ☑️       | `RUN ["npm", "run", "build"]`                          |

<details><summary>CONNECT</summary>

connect to remote SSH server. Its format should be `username@address:port`

eg `CONNECT root@192.168.0.1:22`

</details>

<details><summary>ENV</summary>

set environmental variable for `RUN` command

eg `ENV PRIVATE_KEY = 123`

</details>

<details><summary>VAR</summary>

defining variables. It has 3 ways to define it.

### Set string literals

Its format is this `VAR {key} = {value}`

```s4
VAR PRIVATE_KEY = 123

RUN echo {{PRIVATE_KEY}}
```

### Set environmental variable

Its format is this `VAR {key} = ${envKey}:{tag}`

`tag` can be `local`/`remote`. Used to specify to get local/remote environment variables

```s4
VAR GOPATH_LOCAL = $GOPATH:local

VAR GOPATH_REMOTE = $GOPATH:remote

BASH echo "local GOPATH: {{GOPATH_LOCAL}}"
RUN echo "remote GOPATH: {{GOPATH_REMOTE}}"
```

### Set stdout from execute the command line

Its format is this `VAR {key} <= {bashCommand}`.

This will execute command at remote and set stdout to variable.

or use the format `VAR {key} <= ["{command}", "{argument1}", "{argument2}"]`. It will run in local

```s4
VAR GO_VERSION_LOCAL <= ["go", "version"]

VAR GO_VERSION_REMOTE <= go version

BASH echo "local version: {{GO_VERSION_LOCAL}}"
RUN echo "remote version: {{GO_VERSION_REMOTE}}"
```

```s4
VAR PRIVATE_KEY = 123
ENV PRIVATE_KEY = {{PRIVATE_KEY}}
RUN echo {{PRIVATE_KEY}}
```

</details>

<details><summary>CD</summary>

change current working directory of remote server

eg `CD /home/axetroy`

If the directory does not exist, an error will be thrown

</details>

<details><summary>UPLOAD</summary>

upload local files to remote server

eg `UPLOAD start.py ./server`

It required at least two parameters. The last parameter is the remote server's directory where should be uploaded.

The rest of the parameters are local files path.

</details>

<details><summary>DOWNLOAD</summary>

download remote files to local

eg `DOWNLOAD start.py ./server`

It required at least two parameters. The last parameter is the local directory where should be downloaded.

The rest of the parameters are remote files path.

</details>

<details><summary>COPY</summary>

copy file at remote server

eg `COPY data.db data.db.bak`

</details>

<details><summary>MOVE</summary>

move file at remote server

eg `MOVE data.db data.db.bak`

</details>

<details><summary>DELETE</summary>

delete files at remote server, for security, directory will be ignored

eg `DELETE file1 file2`

</details>

<details><summary>RUN</summary>

run command at remote server

eg `RUN python ./remote/start.py`

It supports multi-line wrap

```s4
RUN npm version \
    && npm run build \
    && npm run test \
    && npm run publish
```

</details>

<details><summary>BASH</summary>

run command in local

eg `RUN python ./local/start.py`

It supports multi-line wrap

```s4
RUN npm version \
    && npm run build \
    && npm run test \
    && npm run publish
```

</details>

<details><summary>CMD</summary>

run command in local

eg `RUN ["npm", "run", "build"]`

It supports multi-line wrap

</details>

### Installation

Download the executable file for your platform at [release page](https://github.com/axetroy/s4/releases)

Then set the environment variable.

eg, the executable file is in the `~/bin` directory.

```bash
# ~/.bash_profile
export PATH="$PATH:~/bin"
```

finally, try it out.

```bash
s4 --help
```

### Upgrade

You can re-download the executable and overwrite the original file.

or type the following command to upgrade to the latest version.

```bash
> s4 upgrade
```

### Build from source code

```bash
> go get -v -u github.com/axetroy/s4
> cd $GOPATH/src/github.com/axetroy/s4
> make build
> ls -lh ./bin

total 24624
-rw-r--r--  1 axetroy  staff   3.9M Aug 28 14:11 s4_linux_x64.tar.gz
-rw-r--r--@ 1 axetroy  staff   3.8M Aug 28 14:11 s4_osx_x64.tar.gz
-rw-r--r--  1 axetroy  staff   3.8M Aug 28 14:11 s4_win_x64.tar.gz
```

### Test

```bash
make test
```

### Why?

> Why do I need such a tool?
> What is its use?

In development, we need to operate remote servers locally, such as deploying services, restarting services, upload files, etc.

We can of course do this with a bash script.

But that is quite cumbersome.

So, I wrote this tool to release my hands.

I hope this helps you.

### License

The [MIT License](https://github.com/axetroy/s4/blob/master/LICENSE)
