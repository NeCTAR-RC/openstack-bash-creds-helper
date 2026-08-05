OpenStack shell creds helper
============================
This is a tool to make managing OpenStack credentials easier, to be used in
combination with included shell function scripts (and completion files) for
bash and fish.

It is written in Go and supports the following Keystone authentication types:
* Password (scoped and unscoped)
* Password + TOTP
* Application Credential

This scan your password store directory for any passwords ending in `.openrc`
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
Google Authenticator, Yubikey OATH) before requesting a token from Keystone.

After loading your credentials and making a request to Keystone, the tool will
then set some environment variables for you to make subsequent OpenStack API
calls, with the token auth method.


Demo
----
<p align="center"><img width="800" src="chcreds.svg"></p>


Use
---
The shell function scripts in this repository provide the following commands:

  * `chcreds` to select and load credentials as username/password in the current environment
  * `recreds` to reload the current credential, which is useful if your token expires
  * `rmcreds` to clear the current credentials from your current environment
  * `prcreds` to print the current credentials


How it works
-------------
The `chcreds` function will call out to the `oscreds` binary to present the list
of credentials available.

Once a credential is chosen, `oscreds` will call out to `pass` to actually
decrypt and return the contents to `oscreds`.

`oscreds` will then interpret the credentials and make subsequent API calls to
Keystone to eventually return an OpenStack token.

The shell function scripts will load the appropriate environment variable for
OpenStack CLI tools to work (almost) seamlessly.

Optionally, you can display your currently loaded credentials in your prompt.
See [Prompt customisation](#prompt-customisation) below for the recommended,
fully customisable `chcreds-ps1` segment.

For a quick, no-frills option you can reference `$OS_CRED` directly:

**Bash** — add `${OS_CRED:+ \[$OS_CRED\]}` to your `PS1` var. For example (coloured):

```
    PS1='\[\033[01;32m\]\u@\h\[\033[01;34m\] \w\[\033[01;33m\]${OS_CRED:+ \[$OS_CRED\]}\[\033[00m\] \$ '
```

**Fish** — add the following to your `fish_prompt` function:

```
    if set -q OS_CRED
        printf ' [%s]' $OS_CRED
    end
```


Prompt customisation
--------------------
For a richer, customisable prompt segment (inspired by
[kube-ps1](https://github.com/jonmosco/kube-ps1)), source the `chcreds-ps1.sh` file
(`chcreds-ps1.fish` for fish) and add `$(chcreds_ps1)` to your prompt. It renders a
segment like `(☁ team/prod)`.

**Bash** — source `chcreds-ps1.sh` from your `.bashrc` and reference `chcreds_ps1` in `PS1`:

``` sh
    source /path/to/chcreds-ps1.sh
    PS1='$(chcreds_ps1) \u@\h \w \$ '
```

**Zsh** — the same file works in zsh; enable `PROMPT_SUBST` so the segment is
evaluated each time:

``` sh
    source /path/to/chcreds-ps1.sh
    setopt PROMPT_SUBST
    PROMPT='$(chcreds_ps1) %n@%m %~ %# '
```

**Fish** — source `chcreds-ps1.fish` and call `chcreds_ps1` from your `fish_prompt`:

``` fish
    source /path/to/chcreds-ps1.fish

    function fish_prompt
        printf '%s ' (chcreds_ps1)
        # ... the rest of your prompt ...
    end
```

### Dynamic (environment-aware) colours

By default the credential name uses the single colour `CHCREDS_PS1_CRED_COLOR`. To
colour it dynamically (for example by environment), define your own function
that takes the credential name and prints a colour, then point
`CHCREDS_PS1_CRED_COLOR_FUNCTION` at it:

``` sh
    chcreds_cred_color() {
        case "$1" in
            *prod*) printf 'red' ;;
            *dev*)  printf 'magenta' ;;
            *test*) printf 'yellow' ;;
            *)      printf 'cyan' ;;
        esac
    }
    export CHCREDS_PS1_CRED_COLOR_FUNCTION=chcreds_cred_color
```

The equivalent in fish:

``` fish
    function chcreds_cred_color
        switch (string lower -- $argv[1])
            case '*prod*'; printf 'red'
            case '*dev*';  printf 'magenta'
            case '*test*'; printf 'yellow'
            case '*';      printf 'cyan'
        end
    end
    set -gx CHCREDS_PS1_CRED_COLOR_FUNCTION chcreds_cred_color
```

### Configuration

All options are `CHCREDS_PS1_*` environment variables. Set them before or after
sourcing the file. Colours accept a name (`black`, `red`, `green`, `yellow`,
`blue`, `magenta`, `cyan`, `white`), a 256-colour code (`0`-`255`; fish accepts
a hex value such as `ff8800`), or an empty string for no colour.

| Variable                    | Default | Purpose                                                        |
|-----------------------------|---------|----------------------------------------------------------------|
| `CHCREDS_PS1_PREFIX`             | `(`     | Opening text of the segment                                    |
| `CHCREDS_PS1_SUFFIX`             | `)`     | Closing text of the segment                                    |
| `CHCREDS_PS1_SEPARATOR`          | space   | Text between the symbol and the credential name                |
| `CHCREDS_PS1_SYMBOL_ENABLE`      | `true`  | Show the symbol                                                |
| `CHCREDS_PS1_SYMBOL`             | `default` | `default` for the cloud glyph `☁`, `openstack` for the nerd-font glyph, or any other value to use it literally |
| `CHCREDS_PS1_SYMBOL_PADDING`     | `false` | Add a space after the symbol (helps with some glyph fonts)     |
| `CHCREDS_PS1_HIDE_IF_NOCREDS`    | `true`  | Show nothing when no credentials are loaded                    |
| `CHCREDS_PS1_PREFIX_COLOR`       | (none)  | Colour of the prefix                                           |
| `CHCREDS_PS1_SYMBOL_COLOR`       | `red`   | Colour of the symbol                                           |
| `CHCREDS_PS1_CRED_COLOR`         | `cyan`  | Colour of the credential name (unless a colour function is set) |
| `CHCREDS_PS1_SUFFIX_COLOR`       | (none)  | Colour of the suffix                                           |
| `CHCREDS_PS1_BG_COLOR`           | (none)  | Background colour for the coloured parts                       |
| `CHCREDS_PS1_CRED_COLOR_FUNCTION`| (unset) | Function name that returns the credential colour              |
| `CHCREDS_PS1_CRED_FUNCTION`      | (unset) | Function name that transforms the displayed credential text    |

To use the nerd-font OpenStack glyph (requires a
[Nerd Font](https://www.nerdfonts.com/)) instead of the default cloud symbol:

``` sh
    export CHCREDS_PS1_SYMBOL=openstack
```

Or set any other value to use it literally as the symbol, for example
`export CHCREDS_PS1_SYMBOL='OS'`.


Choosing the exported auth type
-------------------------------
By default, `chcreds` exports a Keystone token (`OS_AUTH_TYPE=token`). Pass
`--password` to export password auth variables instead (`OS_AUTH_TYPE=password`
with `OS_USERNAME`, `OS_PASSWORD` and the user domain), so clients authenticate
themselves on each request. This is useful for long-running sessions where a
token would expire, or for tools that do not support token auth.

``` sh
    chcreds --password my-cloud
```

This works with the interactive project selector too: `chcreds` still
authenticates once to list your projects and verify access to your selection,
then exports the password variables scoped to the chosen project
(`OS_PROJECT_ID` and `OS_PROJECT_NAME`) instead of a token. It also combines
with `--project`, domain scope and system scope in the same way.

The `--token` flag selects the default token behaviour explicitly, and the two
flags are mutually exclusive. `--password` cannot be used with application
credentials, and for credentials with `OS_TOTP_REQUIRED=true` the exported
password alone will generally not satisfy Keystone, so a warning is printed.

Unlike `OS_CRED_PASSTHROUGH=true`, which skips Keystone entirely and dumps the
credential file as-is, `--password` still validates the credentials and
resolves the project scope for you.

Using token auth
----------------
Using a Keystone token auth directly seems to works well with:
* OpenStack client
* OpenStack APIs

Some known exceptions are documented below:

### Swiftclient

The swiftclient doesn't work directly, but can work with a token by specifying
`--os-auth-token` and `--os-storage-url` directly, where the storage URL is
found from the OpenStack catalog.

```
OS_STORAGE_URL=$(openstack catalog show object-store -f json | jq -r '.endpoints[] | select(.interface=="public" and .region=="Melbourne") | .url')
swift --os-auth-token $OS_TOKEN --os-storage-url $OS_STORAGE_URL
```

Installation
------------
You can grab the latest build from the GitHub project releases page, or see
below for instructions on building it yourself.

Once you have the `oscreds` binary, put it somewhere in your path
(e.g. `~/.local/bin`)

``` sh
    mkdir -p ~/.local/bin
    cp oscreds ~/.local/bin/
```

### Bash (and Zsh)

Grab a copy of the `bash-functions` file from this repo
and drop it into your `.bashrc.d` (or similar) or source it from your `.bashrc`
(or `.zshrc`) to load automatically in your shell.

Optionally source the `chcreds-ps1.sh` file too, to enable the customisable prompt
segment (see [Prompt customisation](#prompt-customisation)):

``` sh
    source /path/to/chcreds-ps1.sh
```

### Fish

Source the `fish-functions` file from your `~/.config/fish/config.fish`:

``` sh
    source /path/to/fish-functions
```

Or copy the individual functions into `~/.config/fish/functions/` as
autoloaded `.fish` files (e.g. `~/.config/fish/functions/chcreds.fish`).

Optionally source `chcreds-ps1.fish` for the customisable prompt segment
(see [Prompt customisation](#prompt-customisation)):

``` sh
    source /path/to/chcreds-ps1.fish
```

Adding credentials
------------------
Add your OpenStack openrc credentials files into pass, ensuring they have a
.openrc extension for oscreds to find them.

``` sh
    pass insert -m my-password.openrc
```

You can then arrange the files in your password store in a way that is
appropriate for your use.

Credential examples
-------------------

Standard password auth
``` sh
    export OS_AUTH_URL=https://keystone.domain.name/
    export OS_PROJECT_NAME=myproject
    export OS_USERNAME=username
    export OS_PASSWORD=password
```

Application credential
``` sh
    export OS_AUTH_URL=https://keystone.domain.name/
    export OS_AUTH_TYPE=v3applicationcredential
    export OS_APPLICATION_CREDENTIAL_ID=app_cred_id
    export OS_APPLICATION_CREDENTIAL_SECRET=app_cred_secret
```

Instead of defining a project, you can set `OS_CRED_PROJECT_DISCOVER=true`
to request a list of projects that you have roles assigned to choose from.
`OS_CRED_*` variables only control chcreds behaviour and are never
exported to your environment.

``` sh
    export OS_AUTH_URL=https://keystone.domain.name/
    export OS_USERNAME=username
    export OS_PASSWORD=password
    export OS_CRED_PROJECT_DISCOVER=true
```

For a domain-scoped account, set `OS_DOMAIN_NAME` or `OS_DOMAIN_ID` and omit
any project variables. The domain scope is passed through for the client to
use. If `OS_CRED_PROJECT_DISCOVER=true` is also set, the domain appears
as an extra choice in the project selection instead.

``` sh
    export OS_AUTH_URL=https://keystone.domain.name/
    export OS_USERNAME=username
    export OS_PASSWORD=password
    export OS_USER_DOMAIN_NAME=mydomain
    export OS_DOMAIN_NAME=mydomain
```

To skip token authentication entirely and load the variables from the
credential file as-is, set `OS_CRED_PASSTHROUGH=true`. This leaves
`OS_PASSWORD` set in the environment for clients to authenticate with
directly, which can be useful where token rescoping is not permitted or a
tool does not support token auth.

``` sh
    export OS_AUTH_URL=https://keystone.domain.name/
    export OS_USERNAME=username
    export OS_PASSWORD=password
    export OS_USER_DOMAIN_NAME=mydomain
    export OS_DOMAIN_NAME=mydomain
    export OS_AUTH_TYPE=password
    export OS_CRED_PASSTHROUGH=true
```

To enable TOTP functionality (if password + TOTP is enabled for identity)
then you need to append `OS_TOTP_REQUIRED=true` to your openrc to trigger
the TOTP prompt.

``` sh
    export OS_AUTH_URL=https://keystone.domain.name/
    export OS_PROJECT_NAME=myproject
    export OS_PROJECT_ID=1234567890abcdef
    export OS_USERNAME=username
    export OS_PASSWORD=password
    export OS_USER_DOMAIN_NAME=Default
    export OS_PROJECT_DOMAIN_NAME=Default
    export OS_TOTP_REQUIRED=true
```

Shell completion
----------------
Completion scripts for both bash and fish are included.

### Bash

To install it for your user, the following should work:

``` sh
    mkdir -p ~/.local/share/bash-completion/completions
    cp bash-completion ~/.local/share/bash-completion/completions/chcreds
```

### Fish

To install it for your user, copy the completion file to your fish completions directory:

``` sh
    mkdir -p ~/.config/fish/completions
    cp fish-completion ~/.config/fish/completions/chcreds.fish
```

You can then use tab completion to complete the filename of the credentials file.

Building
--------
A simple `go build` should suffice to compile the binary.
