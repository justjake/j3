/*
Basic UI tools over Xwindows via xgb and xgbutil
*/
package ui

import (
    "github.com/BurntSushi/xgb/xproto"

    _ "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/xgraphics"
    "github.com/BurntSushi/xgbutil"

    "image"
    "image/color"
    "log"
)

const (
    KeyOption = "Mod1"
    KeySuper  = "Mod4"
    // The idea is to only activate on Mod masked events
    Mod = KeySuper + "-Shift"
)

// create a correctly-sized window from an image
// almost not worth a function
// TODO: transparency

type Icon struct {
    Image   image.Image         // original icon, from PNG usually
    Parent  xproto.Window       // parent X window of the icon
    Window  *xwindow.Window     // window that the icon is drawn to
    ximage  *xgraphics.Image    // X11 image swap buffer, used to paint Image to Window
}

// get the pixels directly behind the icon
// right now just returns an image.Uniform hard-coded to the right shade of grey
// TODO actually do this
func (icn *Icon) getBackground() image.Image {
    log.Println("TODO: Icon.Blend()")
    bg_color := color.RGBA{0xcc, 0xcc, 0xcc, 0xff}
    bg := image.NewUniform(bg_color)
    return bg
}

// blend the RGBA original icon against the background of Icon.Parent
func (icn *Icon) Blend() {
    // get pixels "parent-pixels" of Parent that are behind Window
    bg := icn.getBackground()

    // copy "parent-pixels" into buffer "ximage", overwriting existing completely
    xgraphics.Blend(icn.ximage, bg, image.Point{0,0})

    // alpha-blend Image into buffer "ximage"
    xgraphics.Blend(icn.ximage, icn.Image, image.Point{0,0})

    // swap ximage into Window as background
    icn.ximage.XSurfaceSet(icn.Window.Id)
    icn.ximage.XDraw()
    icn.ximage.XPaint(icn.Window.Id)

    // free the pixbuff memory!
    icn.ximage.Destroy()
}

func (icn *Icon) Move(x, y int) {
    icn.Window.Move(x, y)
    icn.Blend()
}

func NewIcon(X *xgbutil.XUtil, img image.Image, parent xproto.Window) *Icon {
    ximg := xgraphics.NewConvert(X, img)
    win := ximg.Window(parent)
    icn := Icon{img, parent, win, ximg}
    return &icn
}

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

        log.Println("started drag")
        // apparently the cursor is just ID 0
        return true, 0
    }
    // moves the window
    stepDrag := func(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) {
        // ignore duplicate events
        log.Printf("Dragging, root{%v %v}, event{%v %v}", rootX, rootY, eventX, eventY)
        if rootX == lastX && rootY == lastY {
            log.Println("discarded")
            return
        }

        // maintain mouse position within window
        toX := rootX - offsetX
        toY := rootY - offsetY

        // move window
        xwin.Move(toX, toY)
    }
    stopDrag := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {}

    // actually bind handler to window
    mousebind.Drag(X, win, win, "1", true, startDrag, stepDrag, stopDrag)
}



