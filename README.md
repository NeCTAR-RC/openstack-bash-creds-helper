OpenStack bash creds helper
===========================
This is a tool to make managing OpenStack credentials easier, to be used in
combination with included bash functions below (and a bash completion file).

It is written in Go and supports the following Keystone authentication types:
* Password (scoped and unscoped)
* Password + TOTP
* ApplicationCredential

This scan your password store directory for any passwords ending in .openrc
and will display them in a list for you to choose.

The list is powered by the `fzf` tool, which is natively included in the
binary. This allows powerful auto-complete functionality and should make it
super quick to get the credentials you need loaded fast.

It supports openrc files that don't specify a project, and in those cases will
request a list of projects you're a member of from Keystone and allow you to
choose, saving you from duplicating credentials if you're a member of lots of
projects.

This tool also has preliminary support for TOTP, so for accounts that have a
registered TOTP secret, it can prompt for your 6-digit TOTP code (e.g.
Google Authenticator) before requesting a token from Keystone.

After loading your credentials and making a request to Keystone, the tool will
then set some environment variables for you to make subsequent OpenStack API
calls, with the token auth method.


Demo
----
<p align="center"><img width="800" src="chcreds.svg"></p>


Use
---
This bash functions script provides the following commands:

  * `chcreds` to select and load credentials as username/password in the current environment
  * `rmcreds` to clear the current credentials from your current environment
  * `prcreds` to print the current credentials

The `chcreds` function will call out to the oscreds binary to actually load the
credentials, and then provide the environment variables for token auth.


Using token auth
----------------

Using a Keystone token auth directly seems to works well with:
* OpenStack client
* OpenStack APIs

Some known exceptions are:

### Swiftclient

The swiftclient doesn't work directly, but can work with a token by specifying
`--os-auth-token` and `--os-storage-url` directly, where the storage URL is
found from the OpenStack catalog.

```
OS_STORAGE_URL=$(openstack catalog show object-store -f json | jq -r '.endpoints[] | select(.interface=="public" and .region=="Melbourne") | .url')
swift --os-auth-token $OS_TOKEN --os-storage-url $OS_STORAGE_URL
```

Installation
---------------
First, put the `oscreds` script to somewhere in your path (e.g. ~/.local/bin)

``` sh
    mkdir -p ~/.local/bin
    cp oscreds ~/.local/bin/
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
    export OS_AUTH_URL=https://keystone.domain.name:5000/
    export OS_PROJECT_NAME=myproject
    export OS_USERNAME=username
    export OS_PASSWORD=password

```

You can also omit any `OS_PROJECT_NAME` or `OS_PROJECT_ID` to optionally
request a list of projects that you have roles assigned to choose from.

To enable TOTP functionality (if TOTP is enabled for keystone/user and only in token mode) add:
``` sh
    export OS_TOTP_REQUIRED=true
```

And optionally, you could add `$(os_creds)` to your PS1 var for printing your
currently loaded credentials in your prompt. For example (coloured):

```
    PS1='\[\033[01;32m\]\u@\h\[\033[01;34m\] \w\[\033[01;33m\]$(os_creds)\[\033[00m\] \$ '
```
