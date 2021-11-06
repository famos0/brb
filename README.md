# brb
Bruteforce extensions on URLs (bak, bak1, 1, ~, etc.)

## Install

```bash
go get -u github.com/zblurx/brb
```


## Use

```bash
brb by zblurx

Bruteforcing URLs with specific extensions (like .bak, .bak1, ~, .old, etc.)

 -i, --input-file <path>                Specify filepath
 -x, --extensions-list <list>           Specify extensions in comma separated list. Default .bak,.bak1,~..old
 -H, --header <header>                  Specify header. Can be used multiple times
 -c, --cookies <cookies>                Specify cookies
 -x, --proxy <proxy>                    Specify proxy
 -k, --insecure                         Allow insecure server connections when using SSL
 -t, --threads <int>                    Number of thread. Default 10
 -b, --status-code-blacklist <list>     Comme separated list of status code not to output
```