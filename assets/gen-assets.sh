#!/usr/bin/env zsh
# generate go code for binary assets like PNGs

source "$ZSH_FILES/functions.zsh"

echo target is $(this-script-dir)

ASSETS="$(this-script-dir)/_raw"
TARGET="$(this-script-dir)"

generate-go-code () {
    go-bindata -i "$1" -p "assets" 
}

for png in "$ASSETS/"*.png ; do
    generate-go-code "$png"
done

mv "$ASSETS"/*.png.go "$TARGET"
