#!/bin/bash
# shellcheck shell=bash
#
# chcreds-ps1 - show the currently loaded OpenStack credentials in your shell prompt.
#
# Works in both bash and zsh. Source this file, then add $(chcreds_ps1) to your
# prompt:
#
#   bash:  PS1='$(chcreds_ps1) \u@\h \w \$ '
#   zsh:   setopt PROMPT_SUBST; PROMPT='$(chcreds_ps1) %n@%m %~ %# '
#
# Behaviour is driven entirely by CHCREDS_PS1_* variables (see README for the full
# list). Set any of them before or after sourcing this file.

# --- Configuration defaults ---------------------------------------------------
# The `=` form (no colon) only assigns when the variable is *unset*, so an
# explicitly empty value set by the user is preserved.

# Prefix/suffix defaults contain parentheses, so assign them via an unset check
# rather than a `${VAR=...}` default (keeps the parser, and shellcheck, happy
# while still preserving an explicitly empty value set by the user).
[ -z "${CHCREDS_PS1_PREFIX+set}" ] && CHCREDS_PS1_PREFIX='('
[ -z "${CHCREDS_PS1_SUFFIX+set}" ] && CHCREDS_PS1_SUFFIX=')'
: "${CHCREDS_PS1_SEPARATOR= }"

: "${CHCREDS_PS1_SYMBOL_ENABLE=true}"
# CHCREDS_PS1_SYMBOL: "default" (or empty) for the cloud glyph, "openstack" for the
# nerd-font OpenStack glyph, or any other value to use it literally as the symbol.
: "${CHCREDS_PS1_SYMBOL=default}"
: "${CHCREDS_PS1_SYMBOL_PADDING=false}"

: "${CHCREDS_PS1_HIDE_IF_NOCREDS=true}"

: "${CHCREDS_PS1_PREFIX_COLOR=}"
: "${CHCREDS_PS1_SYMBOL_COLOR=red}"
: "${CHCREDS_PS1_CRED_COLOR=cyan}"
: "${CHCREDS_PS1_SUFFIX_COLOR=}"
: "${CHCREDS_PS1_BG_COLOR=}"

# Optional user hook functions (names of functions, not the code itself). Set
# CHCREDS_PS1_CRED_COLOR_FUNCTION to a function that takes the credential name and
# prints a colour to make the credential colour dynamic (e.g. by environment).
: "${CHCREDS_PS1_CRED_COLOR_FUNCTION=}"
: "${CHCREDS_PS1_CRED_FUNCTION=}"

# --- Shell detection ----------------------------------------------------------
# Prompts embedded via command substitution need non-printing markers so the
# shell measures the prompt width correctly. bash uses the raw readline markers
# \001 / \002; zsh uses %{ %}.
if [ -n "${ZSH_VERSION:-}" ]; then
    _chcreds_ps1_open='%{'
    _chcreds_ps1_close='%}'
else
    _chcreds_ps1_open=$'\001'
    _chcreds_ps1_close=$'\002'
fi

# --- Colour helpers -----------------------------------------------------------

# Map a colour (named or a 0-255 256-colour code) to the numeric SGR
# foreground parameters. Empty for an unknown/blank colour.
_chcreds_ps1_ansi_fg() {
    case "$1" in
        black)   printf '30' ;;
        red)     printf '31' ;;
        green)   printf '32' ;;
        yellow)  printf '33' ;;
        blue)    printf '34' ;;
        magenta) printf '35' ;;
        cyan)    printf '36' ;;
        white)   printf '37' ;;
        '')      : ;;
        *) [[ "$1" =~ ^[0-9]+$ ]] && printf '38;5;%s' "$1" ;;
    esac
}

# Same, for the background.
_chcreds_ps1_ansi_bg() {
    case "$1" in
        black)   printf '40' ;;
        red)     printf '41' ;;
        green)   printf '42' ;;
        yellow)  printf '43' ;;
        blue)    printf '44' ;;
        magenta) printf '45' ;;
        cyan)    printf '46' ;;
        white)   printf '47' ;;
        '')      : ;;
        *) [[ "$1" =~ ^[0-9]+$ ]] && printf '48;5;%s' "$1" ;;
    esac
}

# _chcreds_ps1_colorize <colour> <text>
# Print <text> wrapped in the given foreground colour (plus CHCREDS_PS1_BG_COLOR as
# background, if set), with shell-appropriate non-printing markers. Falls back
# to plain text when no colour applies.
_chcreds_ps1_colorize() {
    local color="$1" text="$2" fg bg params
    fg=$(_chcreds_ps1_ansi_fg "$color")
    bg=$(_chcreds_ps1_ansi_bg "$CHCREDS_PS1_BG_COLOR")
    params="$fg"
    [[ -n "$bg" ]] && params="${params:+$params;}$bg"

    if [[ -z "$params" ]]; then
        printf '%s' "$text"
        return
    fi

    printf '%s%s%s%s%s%s%s' \
        "$_chcreds_ps1_open" $'\033['"$params"'m' "$_chcreds_ps1_close" \
        "$text" \
        "$_chcreds_ps1_open" $'\033[0m' "$_chcreds_ps1_close"
}

# --- Main function ------------------------------------------------------------

chcreds_ps1() {
    local cred="${OS_CRED:-}"
    if [[ -z "$cred" ]]; then
        [[ "$CHCREDS_PS1_HIDE_IF_NOCREDS" == "true" ]] && return 0
    fi

    # Optional display-text transform.
    local cred_disp="$cred"
    if [[ -n "$CHCREDS_PS1_CRED_FUNCTION" ]] && type "$CHCREDS_PS1_CRED_FUNCTION" >/dev/null 2>&1; then
        cred_disp=$("$CHCREDS_PS1_CRED_FUNCTION" "$cred")
    fi

    # Cred colour: user hook wins, otherwise the static CHCREDS_PS1_CRED_COLOR.
    local cred_color="$CHCREDS_PS1_CRED_COLOR"
    if [[ -n "$CHCREDS_PS1_CRED_COLOR_FUNCTION" ]] && type "$CHCREDS_PS1_CRED_COLOR_FUNCTION" >/dev/null 2>&1; then
        cred_color=$("$CHCREDS_PS1_CRED_COLOR_FUNCTION" "$cred")
    fi

    # Symbol.
    local symbol=""
    if [[ "$CHCREDS_PS1_SYMBOL_ENABLE" == "true" ]]; then
        case "$CHCREDS_PS1_SYMBOL" in
            default|'') symbol="☁" ;;
            openstack)  symbol=$'\ue856' ;;
            *)          symbol="$CHCREDS_PS1_SYMBOL" ;;
        esac
        [[ "$CHCREDS_PS1_SYMBOL_PADDING" == "true" ]] && symbol="$symbol "
    fi

    local out=""
    out+=$(_chcreds_ps1_colorize "$CHCREDS_PS1_PREFIX_COLOR" "$CHCREDS_PS1_PREFIX")
    if [[ -n "$symbol" ]]; then
        out+=$(_chcreds_ps1_colorize "$CHCREDS_PS1_SYMBOL_COLOR" "$symbol")
        out+="$CHCREDS_PS1_SEPARATOR"
    fi
    out+=$(_chcreds_ps1_colorize "$cred_color" "$cred_disp")
    out+=$(_chcreds_ps1_colorize "$CHCREDS_PS1_SUFFIX_COLOR" "$CHCREDS_PS1_SUFFIX")

    printf '%s' "$out"
}

# vim:syntax=sh
