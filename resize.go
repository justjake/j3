package main

/*
Drag handler and support structure to manage resizing a window or
windows(s) in conjunction
*/

import (
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/mousebind"

    "github.com/justjake/j3/wm"
    "github.com/justjake/j3/ui"

    "log"
)

const KeyComboResize = ui.KeyOption+"-Control-1"

// i am pretty bad at software engineering whist drunk
type Edge struct {
    Side      wm.Direction
    Window    *xwindow.Window
    Adjacent  []*xwindow.Window
}

type Point struct {
    X, Y int
}

// resize windows with 
func ManageResizingWindows(X *xgbutil.XUtil) {
    // var dragged *Edge

    handleDragStart := func(X *xgbutil.XUtil, rx, ry, ex, ey int) (cont bool, cursor xproto.Cursor) {
        // get the clicked window
        win, err := wm.FindManagedWindowUnderMouse(X)
        if err != nil {
            log.Printf("ResizeStart: couldn't find window under mouse: %v\n", err)
            return false, 0
        }
        // get coordinates inside the clicked window
        _, reply, err := wm.FindNextUnderMouse(X, win)
        if err != nil {
            log.Printf("ResizeStart: couldn't get coordinates of click inside win %v: %v\n", win, err)
            return false, 0
        }
        x, y := int(reply.WinX), int(reply.WinY)

        // create an xwindow.Window so we can get a rectangle to find our bearings from
        xwin := xwindow.New(X, win)
        geom, err := xwin.Geometry()
        if err != nil {
            log.Printf("ResizeStart: geometry error: %v\n", err)
            return false, 0
        }

        // construct algebraic functions to delinate the window into sections
        // around the center point
        // these are a little confusing because x11 addresses coordinates from the top-left,
        // where traditional euclidean graphs address from the bottom-left
        w, h := geom.Width(), geom.Height()
        slope := float64(h) / float64(w)

        bl_to_tr := func (x int) (y int) {
            return int(-1.0 * slope * float64(x)) + h
        }
        tl_to_br := func (x int) (y int) {
            return int(slope * float64(x))
        }

        bl_to_tr_y := bl_to_tr(x)
        tl_to_br_y := tl_to_br(x)
        var dir wm.Direction

        // decide what edge we are
        if x < w/2 {
            log.Println("X on left side")
            if y <= tl_to_br_y {
                // we must be above both lines and on the left side of the midpoint
                dir = wm.Top
            } else if y >= bl_to_tr_y {
                // we must be below both lines and on the left side of the midpoint
                dir = wm.Bottom
            } else {
                // we are between the two lines and on the left side
                dir = wm.Left
            }
        } else {
            if y <= bl_to_tr_y {
                // we must be above both lines and on the left side of the midpoint
                dir = wm.Top
            } else if y >= tl_to_br_y {
                // we must be below both lines and on the left side of the midpoint
                dir = wm.Bottom
            } else {
                // we are between the two lines and on the left side
                dir = wm.Right
            }
        }
        log.Printf("ResizeStart: click at (%v, %v). tl_to_br(x)=%v, bl_to_tr(x)=%v  Direction: %v\n", x, y, tl_to_br_y, bl_to_tr_y, xwin.Id, dir.String())

        // TODO: finish this
        // create an edge
        // find the adjacent windows
        // start the drag
        return true, 0
    }

    // bindings for testing
    mousebind.ButtonPressFun(
        func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
            _, _ = handleDragStart(X, 0, 0, 0, 0)
        }).Connect(X, X.RootWin(), KeyComboResize, false, true)

}
