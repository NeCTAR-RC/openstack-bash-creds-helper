OpenStack bash creds helper
===========================
This is a script to make managing OpenStack credentials easier, to be used in
combination with included bash functions below (and a bash completion file).

Update: now with colours!

Use
---
This script provides the following commands:

  * `chchreds` to change which credentials to load
  * `creds`    to load the existing credentials in a new environment
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
    export OS_AUTH_URL=http://keystone.domain.name:5000/v3/
    export OS_NO_CACHE=true
    export OS_PROJECT_NAME=tenant
    export OS_USERNAME=username
    export OS_PASSWORD=password
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
