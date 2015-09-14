penStack bash creds helper
===========================
This is a script to make managing OpenStack credentials easier, to be used in combination with the .bashrc functions below.

Use
---
This script provides the following commands:

  * `chchreds` to change which credentials to load
  * `creds`    to load the existing credentials in a new environment
  * `rmcreds`  to clear the current credentials from your current environment

Installation
---------------
First, put the `openstack_creds` script to somewhere in your path.

Then add the following functions to your .bashrc:

``` sh
    function creds() {
        if [ -f $HOME/.openstack-creds/.creds ]; then
            source $HOME/.openstack-creds/.creds
            echo "$OS_CRED_FILE"
        else
            chcreds
        fi
    }
    function chcreds() {
        nectar_creds && creds
    }
    function rmcreds() {
        source $HOME/.openstack-creds/.rmcreds
        rm -f $HOME/.openstack-creds/.creds
    }
```

Create the directory `~/.openstack-creds`

``` sh
    mkdir  ~/.openstack-creds
    chmod 700 ~/.openstack-creds
```
    
Then add any OpenStack credentials files to `~/.openstack-creds`

The format of the credential files should look something like this:

``` sh
    export OS_AUTH_URL=http://keystone.domain.name:5000/v2.0/
    export OS_NO_CACHE=true
    export OS_TENANT_NAME=tenant
    export OS_USERNAME=username
    export OS_PASSWORD=password
```

And optionally, you could add this for adding the currently loaded credentials to your bash prompt:

``` sh
    function os_creds {
        [ -z $OS_CRED_FILE ] || echo " ${OS_CRED_FILE}"
    }
```
Then add $(os_creds) to your PS1 var. For example (coloured):

```
    PS1='\[\033[01;32m\]\u@\h\[\033[01;34m\] \w\[\033[01;33m\]$(os_creds)\[\033[00m\] \$ '
```
