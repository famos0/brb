# brb
Bruteforce extensions on URLs (bak, bak1, 1, ~, etc.)

## Install

```bash
go get -u github.com/famos0/brb
```


## Use

```bash
brb by famos0

Bruteforcing URLs with specific extensions (like bak, bak1, 1, etc.)

 -i, --input-file <path>        Specify filepath
 -x, --extensions-list <path>   Specify extension list like bak,bak1,etc.
 -H, --header <header>          Specify header. Can be used multiple times
 -c, --cookies <cookies>        Specify cookies
 -x, --proxy <proxy>            Specify proxy
 -k, --insecure                 Allow insecure server connections when using SSL
 -t, --threads <int>            Number of thread. Default 10
```