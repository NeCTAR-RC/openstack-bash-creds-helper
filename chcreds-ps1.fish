# chcreds-ps1 - show the currently loaded OpenStack credentials in your fish prompt.
#
# Source this file, then call chcreds_ps1 from your fish_prompt function:
#
#   function fish_prompt
#       printf '%s ' (chcreds_ps1)
#       # ... the rest of your prompt ...
#   end
#
# Behaviour is driven entirely by CHCREDS_PS1_* variables (see README for the full
# list). fish's set_color measures prompt width correctly, so no manual
# non-printing markers are needed. Colours may be named (red, blue, ...) or hex
# (e.g. ff8800).

# --- Configuration defaults ---------------------------------------------------
# `set -q` is true when the variable is set (even to an empty string), so an
# explicitly empty value set by the user is preserved.

set -q CHCREDS_PS1_PREFIX; or set -g CHCREDS_PS1_PREFIX '('
set -q CHCREDS_PS1_SUFFIX; or set -g CHCREDS_PS1_SUFFIX ')'
set -q CHCREDS_PS1_SEPARATOR; or set -g CHCREDS_PS1_SEPARATOR ' '

set -q CHCREDS_PS1_SYMBOL_ENABLE; or set -g CHCREDS_PS1_SYMBOL_ENABLE true
# CHCREDS_PS1_SYMBOL: "default" (or empty) for the cloud glyph, "openstack" for the
# nerd-font OpenStack glyph, or any other value to use it literally as the symbol.
set -q CHCREDS_PS1_SYMBOL; or set -g CHCREDS_PS1_SYMBOL default
set -q CHCREDS_PS1_SYMBOL_PADDING; or set -g CHCREDS_PS1_SYMBOL_PADDING false

set -q CHCREDS_PS1_HIDE_IF_NOCREDS; or set -g CHCREDS_PS1_HIDE_IF_NOCREDS true

set -q CHCREDS_PS1_PREFIX_COLOR; or set -g CHCREDS_PS1_PREFIX_COLOR ''
set -q CHCREDS_PS1_SYMBOL_COLOR; or set -g CHCREDS_PS1_SYMBOL_COLOR red
set -q CHCREDS_PS1_CRED_COLOR; or set -g CHCREDS_PS1_CRED_COLOR cyan
set -q CHCREDS_PS1_SUFFIX_COLOR; or set -g CHCREDS_PS1_SUFFIX_COLOR ''
set -q CHCREDS_PS1_BG_COLOR; or set -g CHCREDS_PS1_BG_COLOR ''

# Optional user hook functions (names of functions, not the code itself). Set
# CHCREDS_PS1_CRED_COLOR_FUNCTION to a function that takes the credential name and
# prints a colour to make the credential colour dynamic (e.g. by environment).
set -q CHCREDS_PS1_CRED_COLOR_FUNCTION; or set -g CHCREDS_PS1_CRED_COLOR_FUNCTION ''
set -q CHCREDS_PS1_CRED_FUNCTION; or set -g CHCREDS_PS1_CRED_FUNCTION ''

# --- Colour helper ------------------------------------------------------------

# _chcreds_ps1_colorize <colour> <text>
function _chcreds_ps1_colorize
    set -l color $argv[1]
    set -l text $argv[2]

    if test -z "$color"; and test -z "$CHCREDS_PS1_BG_COLOR"
        printf '%s' $text
        return
    end

    set -l args
    test -n "$color"; and set args $args $color
    test -n "$CHCREDS_PS1_BG_COLOR"; and set args $args --background $CHCREDS_PS1_BG_COLOR

    set_color $args
    printf '%s' $text
    set_color normal
end

# --- Main function ------------------------------------------------------------

function chcreds_ps1
    set -l cred ''
    set -q OS_CRED; and set cred $OS_CRED
    if test -z "$cred"
        test "$CHCREDS_PS1_HIDE_IF_NOCREDS" = true; and return 0
    end

    # Optional display-text transform.
    set -l cred_disp $cred
    if test -n "$CHCREDS_PS1_CRED_FUNCTION"; and functions -q $CHCREDS_PS1_CRED_FUNCTION
        set cred_disp ($CHCREDS_PS1_CRED_FUNCTION $cred)
    end

    # Cred colour: user hook wins, otherwise the static CHCREDS_PS1_CRED_COLOR.
    set -l cred_color $CHCREDS_PS1_CRED_COLOR
    if test -n "$CHCREDS_PS1_CRED_COLOR_FUNCTION"; and functions -q $CHCREDS_PS1_CRED_COLOR_FUNCTION
        set cred_color ($CHCREDS_PS1_CRED_COLOR_FUNCTION $cred)
    end

    # Symbol.
    set -l symbol ''
    if test "$CHCREDS_PS1_SYMBOL_ENABLE" = true
        switch "$CHCREDS_PS1_SYMBOL"
            case default ''
                set symbol '☁'
            case openstack
                set symbol \ue856
            case '*'
                set symbol $CHCREDS_PS1_SYMBOL
        end
        test "$CHCREDS_PS1_SYMBOL_PADDING" = true; and set symbol "$symbol "
    end

    _chcreds_ps1_colorize "$CHCREDS_PS1_PREFIX_COLOR" "$CHCREDS_PS1_PREFIX"
    if test -n "$symbol"
        _chcreds_ps1_colorize "$CHCREDS_PS1_SYMBOL_COLOR" "$symbol"
        printf '%s' "$CHCREDS_PS1_SEPARATOR"
    end
    _chcreds_ps1_colorize "$cred_color" "$cred_disp"
    _chcreds_ps1_colorize "$CHCREDS_PS1_SUFFIX_COLOR" "$CHCREDS_PS1_SUFFIX"
end
