OpenStack bash creds helper
===========================
This is a script to make managing OpenStack credentials easier, to be used in
combination with included bash functions below (and a bash completion file).

This can optionally be used in conjunction with `fzf` the command line fuzzy
finder (https://github.com/junegunn/fzf) for a faster, more interactive chooser
(recommended!)


Demo
----
<p align="center"><img width="800" src="chcreds.svg"></p>


Use
---
This script provides the following commands:

  * `chcreds` to select and load credentials as username/password in the current environment
  * `creds`    to load the existing selected credentials as username/password in a new environment
  * `chcreds_token` to select and load credentials as a token in the current environment
  * `creds_token`    to load the existing selected credentials as a token in a new environment
  * `rmcreds`  to clear the current credentials from your current environment
  * `prcreds`  to print the current credentials


Installation
---------------
First, put the `openstack_creds` script to somewhere in your path (e.g. ~/.local/bin)

``` sh
    mkdir -p ~/.local/bin
    cp openstack_creds ~/.local/bin/
```

Then source bash functions in your `.bashrc`:

``` sh
	# openstack credentials
    . ~/source/openstack-bash-creds-helper/bash-functions
```

Add your OpenStack openrc credentials files into pass, ensuring they have a
.openrc extension.

``` sh
    pass insert -m my-password.openrc
```

The format of the credential files should look something like this:

``` sh
    export OS_AUTH_URL=http://keystone.domain.name:5000/
    export OS_NO_CACHE=true
    export OS_PROJECT_NAME=tenant
    export OS_USERNAME=username
    export OS_PASSWORD=password
    export OS_IDENTITY_API_VERSION=3
```
To enable TOTP functionality (if TOTP is enabled for keystone/user and only in token mode) add:
``` sh
    export CHCREDS_MFA_TOTP_PASS=true
```


And optionally, you could add `$(os_creds)` to your PS1 var for printing your
currently loaded credentials in your prompt. For example (coloured):

```
    PS1='\[\033[01;32m\]\u@\h\[\033[01;34m\] \w\[\033[01;33m\]$(os_creds)\[\033[00m\] \$ '
```

Bash completion
---------------
A bash completion script is also included for your convenience.

To install it for your user, the following should work:

``` sh
    mkdir -p ~/.local/share/bash-completion/completions
    cp bash-completion ~/.local/share/bash-completion/completions/chcreds
```

Depending on your version of bash-completion, you may have to install this file into
`/etc/bash_completion.d` instead (e.g. Ubuntu Xenial).

You can then use tab completion to complete the filename of the credentials file.
