## introduction

`rfschub-server` is the backend of web application `rfschub`, `rfsc` is the abbreviation for the famous slogan **Read the Fxxking Source Code**. You can create source code reading project or join other's reading project, make some comments while reading the code. Have fun reading code!

## Getting started

### prerequisite

1. `go` at least 1.12
2. mongodb
3. git default in /usr/local/bin
4. [universal-ctags](https://github.com/universal-ctags/ctags) default in /usr/local/bin
5. [syntect_server](https://github.com/sourcegraph/syntect_server) default in /usr/local/bin

### start dev server

```sh
go get github.com/lt90s/rfschub-server
cd rfschub-server
./dev/start.sh
```

After starting the server, head to [rfschub-web](https://github.com/lt90s/rfschub-web) to see how to serve the webUI.

