# j3 - window management hotfix

copyright 2013, Jake Teton-Landis. All rights reserverd.

The goal of j3 is to suplement an existing window manager that you
already use and enjoy, by adding a few small window managemnt behaviors
for the mouse.

## Features

j3 currently has 3 window management actions. These actions are
preformed by dragging a window (the *incoming* window) over another
window (the *target* window), and releasing the mouse over one of
several j3 drop targets.

    1. ![Swap][swap]
       Drag a window onto the center swap of another window to **swap** 
       the positions and sizes of two windows

    2. ![Split - Top][st] ![Split - Right][sr] ![Split - Bottom][sb] 
       ![Split - Left][sl] 
       Drag a window onto any of the split icons to split the target
       window in half. The incoming window will take the highlighted
       half, as indicated on the icon, leaving the target to fill the
       other half.

    3. ![Shove - Top][vt] ![Shove - Right][vr] ![Shove - Bottom][vb] 
       ![Shove - Left][vl] 
       Drag a window onto any of the shove icons to move the incoming
       window to that side of the target window. The incoming window
       will be resized so that its edges are flush with the target.

[swap]: https://raw.github.com/justjake/j3/master/assets/_raw/swap-center.png
[st]: https://raw.github.com/justjake/j3/master/assets/_raw/split-top.png
[sr]: https://raw.github.com/justjake/j3/master/assets/_raw/split-right.png
[sb]: https://raw.github.com/justjake/j3/master/assets/_raw/split-bottom.png
[sl]: https://raw.github.com/justjake/j3/master/assets/_raw/split-left.png
[vt]: https://raw.github.com/justjake/j3/master/assets/_raw/shove-top.png
[vr]: https://raw.github.com/justjake/j3/master/assets/_raw/shove-right.png
[vb]: https://raw.github.com/justjake/j3/master/assets/_raw/shove-bottom.png
[vl]: https://raw.github.com/justjake/j3/master/assets/_raw/shove-left.png
       

## Running

j3 should work with any EWMH-compliant, floating-style window manager.
The primary host target is Fluxbox.

To build j3, you must have a Go runtime, and development headers for
your X11 window server.

j3 is a go-installable go program. If you already have Go, then you can
just run 

    $ go get -u github.com/justjake/j3

If you don't have Go installed already, the [Go downloads page][1] has
pretty good instructions to get you started.

After installation, just run `j3 & disown` to start the manager.

## Configuration

j3 only responds to special key combinations. By default, the key
combination is Option-Shift in combination with a left-mouse-button
drag. This key combination can be changed by modifying the `KeyCombo`
constant at the top of `executable.go` and re-installing j3.
