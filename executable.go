package main
// j3
// Mousing assistance for i3 window manage3r

import (
    "github.com/BurntSushi/xgb/xproto"

    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xrect"
    "github.com/BurntSushi/xgbutil/xevent"  // we'll need it eventially
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/icccm"
    "github.com/BurntSushi/xgbutil/xwindow"
    _ "github.com/BurntSushi/xgbutil/xgraphics"

    "log"

    "image"

    "github.com/justjake/j3/assets"
    "github.com/justjake/j3/ui"
)

const (
    WmName = "i3"  // only runs if the window manager has this name
    StripBackgroundColor = 0xcccccc
    IconMargin = 20 // space around each iconyy in pixels
)

var (
    IconSize = assets.SwapCenter.Bounds().Dx() // assume square icons
    ActionStripWidth = IconMargin * 2 + IconSize   // action strips in vertical orientation
    ActionStripHeight = IconSize * 5 + IconMargin * 6 // 5 icons with margin between and at top and bottom 
    StripGeometryHorizontal = xrect.New(0, 0, ActionStripHeight, ActionStripWidth)
    StripGeometryVertical = xrect.New(0, 0, ActionStripWidth, ActionStripHeight)
    // see http://standards.freedesktop.org/wm-spec/wm-spec-latest.html#idp6304176
    StripWindowType = []string{ "_NET_WM_WINDOW_TYPE_SPLASH" }
)
    

func fatal(err error) {
    if err != nil {
        log.Panic(err)
    }
}

// return the correct X, Y to center rect A over rect B
func centerOver(a, b xrect.Rect) (x, y int) {
    b_center_x := b.X() + b.Width() / 2
    b_center_y := b.Y() + b.Height() / 2

    x = b_center_x - a.Width() / 2
    y = b_center_y - a.Height() / 2
    return
}

func centerChild(child, parent xrect.Rect) (x, y int) {
    a := child
    b := parent
    b_center_x := b.Width() / 2
    b_center_y := b.Height() / 2

    x = b_center_x - a.Width() / 2
    y = b_center_y - a.Height() / 2
    return
}



// set up a floating menu strip window with the right properties
// float, no-rezise window
func configure (X *xgbutil.XUtil, win *xwindow.Window) {
    ewmh.WmWindowTypeSet(X, win.Id, StripWindowType)

    geom, err := win.Geometry()
    fatal(err)

    log.Printf("configuring window %v - dimensions: %v", win, geom)

    // fix size using WM_NORMAL_HINTS
    // see http://tronche.com/gui/x/icccm/sec-4.html#s-4.1.2.3
    // note that this is not necissary now that we use a OverrideRedirect
    // window

    normal_hints := icccm.NormalHints{
        Flags: 16 | 32 | 512, // PMinSize | PMaxSize | PWinGravity
        MinWidth:  uint(geom.Width() )  ,
        MaxWidth:  uint(geom.Width() )  ,
        MinHeight: uint(geom.Height()),
        MaxHeight: uint(geom.Height()),
        WinGravity: 5, // center gravity
    }
    icccm.WmNormalHintsSet(X, win.Id, &normal_hints)

    // also be an override redirect popup????
    // see http://tronche.com/gui/x/icccm/sec-4.html#s-4.1.10

}

func createWindow(X *xgbutil.XUtil, geom xrect.Rect, parent *xwindow.Window) *xwindow.Window {
    // find x, y to center over parent
    parent_geo, err := parent.Geometry()
    fatal(err)
    x, y := centerChild(geom, parent_geo)

    // create the window
    win, err := xwindow.Generate(X)
    fatal(err)
    // create the window as an OverrideRedirect, which is UNMANAGED
    // by any window manager. 
    win.Create(parent.Id, x, y, geom.Width(), geom.Height(), 
        xproto.CwBackPixel | xproto.CwOverrideRedirect, 
        StripBackgroundColor, 1)
    return win
}

func createVertical(X *xgbutil.XUtil, parent *xwindow.Window) *xwindow.Window {
    // first create the window
    win := createWindow(X, StripGeometryVertical, parent)

    deltaY := IconMargin + IconSize
    offsetX := IconMargin
    offsetY := IconMargin

    imgs := []image.Image{assets.ShoveTop, assets.SplitTop, assets.SwapCenter, assets.SplitBottom, assets.ShoveBottom}
    for i, img := range imgs {
        icon := ui.NewIcon(X, img, win.Id)
        icon.Move(offsetX, offsetY + deltaY * i)
        icon.Window.Map()
        log.Printf("Created icon %v for vertical window\n", icon)
    }

    return win
}


func main() {
    // establish X connection
    X, err := xgbutil.NewConn()
    fatal(err)

    // make sure i3 is running or something
    wm_name, err := ewmh.GetEwmhWM(X)
    fatal(err)
    log.Printf("Window manager: %s\n", wm_name)
    //if wm_name != WmName {
    //    log.Panicf("Expected window manager to be '%s' but detected '%s' instead", WmName, wm_name)
    //}

    root := xwindow.New(X, X.RootWin())

    // create vertical options window
    vert := createVertical(X, root)

    // TODO - make windows floating
    //configure(X, vert)
    //configure(X, horiz)
    
    // TODO - show icons (!)

    // TODO - bind listeners on window events
    ui.MakeDraggable(X, vert.Id)

    // map windows -- this displays em!
    vert.Map()

    //ximg := xgraphics.NewConvert(X, assets.SwapCenter)
    //ximg.XShow()


    // start event loop, even though we have no events
    // to keep app from just closing
    xevent.Main(X)
}
        

