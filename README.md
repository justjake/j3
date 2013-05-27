# j3 - window management hotfix

copyright 2013, Jake Teton-Landis. All rights reserverd.

The goal of j3 is to suplement an existing window manager that you
already use and enjoy, by adding a few small window managemnt behaviors
for the mouse.

j3 should work with any EWMH-compliant, floating-style window manager.
The primary host window manager target is Fluxbox, but I intend to add
support for tiling window managers like i3 at some point.

## Features

j3 currently has 3 window management actions. These actions are
preformed by dragging a window (the *incoming* window) over another
window (the *target* window), and releasing the mouse over one of
several j3 drop targets.

 1. ### Swap

    ![Swap][swap]

    Drag a window onto the center swap of another window to **swap** 
    the positions and sizes of two windows

 2. ### Split

    ![Split - Top][st] ![Split - Right][sr] ![Split - Bottom][sb] 
    ![Split - Left][sl] 

    Drag a window onto any of the split icons to split the target
    window in half. The incoming window will take the highlighted
    half, as indicated on the icon, leaving the target to fill the
    other half.

 3. ### Shove

    ![Shove - Top][vt] ![Shove - Right][vr] ![Shove - Bottom][vb] 
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

To build j3, you must have a Go runtime, and development headers for
your X11 window server. For more information on X11 development with Go,
please see [The X Go Binding][1].

j3 is a go-installable go program. If you already have Go, then you can
just run 

    $ go get -u github.com/justjake/j3

If you don't have Go installed already, the [Go downloads page][2] has
pretty good instructions to get you started.

[1]: http://godoc.burntsushi.net/pkg/github.com/BurntSushi/xgb/
[2]: http://golang.org/doc/install#download

After installation, just run `j3 & disown` to start the manager.

## Configuration

j3 only responds to special key combinations. By default, the key
combination is Option-Shift in combination with a left-mouse-button
drag. This key combination can be changed by modifying the `KeyCombo`
constant at the top of `executable.go` and re-installing j3.

## Plans

We can seperate the issues into j3 into two categories: additional
features, and improvements to the present behavior. Here are the
features that I am interested in adding:

### Seam resizing

Seam-based resizing. The user should be able to resize adjacent
windows at the same time by dragging on the shared edge between the
windows. Right now the "split" action is still rather cumbersome if
you are using a floating window manager: if you want a different ratio,
you can use the split to start your layout, but you must then resize and
move both windows to achieve layout zen.

My plan for seam resizing is to implement the following algorithm, and
bind it to a different key combination than the current window movement
action.

    on-mouse-down:
        incoming = window-under-mouse()
        coords = get-mouse-coordinates()
        edge = incoming.edge-closest-to(coords)

        target = ->
            for win in get-managed-windows():
                if win.has-edge(edge):
                    return win

        if target isnt nil:
            seam-manager.start_drag(incoming, edge, target)
        else:
            // perform-normal-rescale
            resize-manager.start_drag(incoming, edge)

    on-mouse-move:
        move-edge-to-coords(
            correct-manager().edge(),
            get-mouse-coordinates()
        )

    on-mouse-up:
        correct-manager().end()
            
I don't really want to be testing if the mouse is over an edge, so we'll
do this Fluxbox style with resize-from-anywhere by detecting what edge
the mouse is closest to, by drawing an [X] across the window to divide
it into directional quadrants

### Usability improvements

- Display an icon while dragging the window
- hide the move window UI when the mouse leaves a window to the desktop
- handle overlapping windows.
  
  This is a difficult problem. If we have two indows like this:

        ______________________________
        |                            |   
        | OUTER                      |     
        |   _______________          |     
        |   |               |   
        |   | INNER         |   
        |   |               |   
        |   |               |   
        |   |               |   
        |   |               |   
        |   | _____________ |   
        |                            |     
        |                            |     
        |                            |     
        | ___________________________|

  How can we properly target OUTER with the UI, when it will be placed over INNER?

  This is a question that seems difficult to answer without introducing more 
  state into the program
