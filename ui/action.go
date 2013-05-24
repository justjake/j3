/*
Basic UI tools over Xwindows via xgb and xgbutil
*/
package ui

import (
    "github.com/BurntSushi/xgb/xproto"

    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil"

    "log"
)

const (
    KeyOption = "Mod1"
    KeySuper  = "Mod4"
    // The idea is to only activate on Mod masked events
    Mod = KeySuper + "-Shift"
)



// drag around windows with the mouse.
func MakeDraggable(X *xgbutil.XUtil, win xproto.Window) {
    // utility window for movement
    xwin := xwindow.New(X, win)

    // state
    var offsetX, offsetY int
    var lastX, lastY int

    // saves initial click location
    startDrag := func(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) (bool, xproto.Cursor) {
        offsetX = eventX
        offsetY = eventY
        lastX = rootX
        lastY = rootY

        // apparently the cursor is just ID 0
        return true, 0
    }
    // moves the window
    stepDrag := func(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) {
        // maintain mouse position within window
        toX := rootX - offsetX
        toY := rootY - offsetY

        // move window
        xwin.Move(toX, toY)
    }
    stopDrag := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {}

    // actually bind handler to window
    mousebind.Drag(X, win, win, "1", true, startDrag, stepDrag, stopDrag)
    log.Printf("MakeDraggable: activated window %v\n", xwin)
}
