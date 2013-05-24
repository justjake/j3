package main
// j3
// Mousing assistance for i3 window manage3r

import (
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgb/shape"

    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xrect"
    "github.com/BurntSushi/xgbutil/xevent"  // we'll need it eventially
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/mousebind"
    _ "github.com/BurntSushi/xgbutil/xgraphics"

    "log"

    "image"

    "github.com/justjake/j3/assets"
    "github.com/justjake/j3/ui"
    "github.com/justjake/j3/wm"
)

const (
    WmName = "i3"  // only runs if the window manager has this name
    StripBackgroundColor = 0xcccccc
    IconMargin = 15 // space between icons and border
    IconPadding = 25 // space between two icons
)

var (
    // TODO: icons are not square. Make the geometry calculations right!
    // TODO: IconWidth and IconHeight replace IconSize
    IconSize = assets.SwapCenter.Bounds().Dx() // assume square icons
    ActionStripWidth = IconMargin * 2 + IconSize   // action strips in vertical orientation
    ActionStripHeight = IconSize * 5 + IconPadding * 4 + IconMargin * 2 // 5 icons with margin between and at top and bottom 

    StripGeometryHorizontal = xrect.New(0, 0, ActionStripHeight, ActionStripWidth)
    StripGeometryVertical = xrect.New(0, 0, ActionStripWidth, ActionStripHeight)
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

func makeCross(X *xgbutil.XUtil) *xwindow.Window {
    // create destination window
    // we will replace its geometry with one created using xgb.Shape via ui.ComposeShape
    cross, err := xwindow.Generate(X)
    fatal(err)
    cross.Create(X.RootWin(), 0, 0, ActionStripHeight, ActionStripHeight, 
        xproto.CwBackPixel | xproto.CwOverrideRedirect, 
        StripBackgroundColor, 1)

    // create geometries
    geo, err := cross.Geometry()
    fatal(err)

    var x, y int
    rects := make([]xrect.Rect, 2)
    // clone the strip goemetries because we're gonna change thier X, Y offsets
    rects[0] = xrect.New(StripGeometryVertical.Pieces())
    rects[1] = xrect.New(StripGeometryHorizontal.Pieces())

    // center the geometry legs over the target window, and thus each other
    x, y = centerOver(rects[0], geo)
    rects[0].XSet(x)
    rects[0].YSet(y)
    x, y = centerOver(rects[1], geo)
    rects[1].XSet(x)
    rects[1].YSet(y)

    // compose the two rects into a + (!)
    err = ui.ComposeShape(X, cross.Id, rects)
    fatal(err)

    return cross
}


func addHorizontalIcons(X *xgbutil.XUtil, win *xwindow.Window, offsetY int) []*ui.Icon {
    deltaX := IconPadding + IconSize
    offsetX := IconMargin

    imgs := []image.Image{assets.ShoveLeft, assets.SplitLeft, nil, assets.SplitRight, assets.ShoveRight}
    icons := make([]*ui.Icon, len(imgs))
    for i, img := range imgs {

        // drop center (nil) icon
        if img == nil {
            continue
        }

        icon := ui.NewIcon(X, img, win.Id)
        icon.Move(offsetX + deltaX * i, offsetY)
        icon.Window.Map()

        icons[i] = icon

        // disable Shove actions for now -- they make no sense for general win managers
        if i == 0 || i == 4 {
            icon.SetState(ui.StateDisabled)
        }
        log.Printf("Created icon %v for horizontal window\n", icon)
    }

    return icons

}

// add the vertical icons row to a window
func addVerticalIcons(X *xgbutil.XUtil, win *xwindow.Window, offsetX int) []*ui.Icon {

    deltaY := IconPadding + IconSize
    offsetY := IconMargin

    imgs := []image.Image{assets.ShoveTop, assets.SplitTop, assets.SwapCenter, assets.SplitBottom, assets.ShoveBottom}
    icons := make([]*ui.Icon, len(imgs))
    for i, img := range imgs {
        icon := ui.NewIcon(X, img, win.Id)
        icon.Move(offsetX, offsetY + deltaY * i)
        icon.Window.Map()

        icons[i] = icon

        // disable Shove actions for now -- they make no sense for general win managers
        if i == 0 || i == 4 {
            icon.SetState(ui.StateDisabled)
        }
        log.Printf("Created icon %v for vertical window\n", icon)
    }
    return icons
}


func main() {
    // establish X connection
    X, err := xgbutil.NewConn()
    fatal(err)

    // initiate extension tools
    shape.Init(X.Conn())
    mousebind.Initialize(X)

    // make sure i3 is running or something
    wm_name, err := ewmh.GetEwmhWM(X)
    fatal(err)
    log.Printf("Window manager: %s\n", wm_name)
    //if wm_name != WmName {
    //    log.Panicf("Expected window manager to be '%s' but detected '%s' instead", WmName, wm_name)
    //}

    // Produce visual UI
    cross := makeCross(X)

    // show icons
    offset := IconMargin + 2 * (IconSize + IconPadding)
    vert := addVerticalIcons(X, cross, offset)
    _ = addHorizontalIcons(X, cross, offset)

    // TODO - bind listeners on window events
    ui.MakeDraggable(X, cross.Id)

    // map windows -- this displays em!
    cross.Map()



    // Bind events

    dm := wm.NewDragManager(X, cross)
    // center action
    dz := wm.NewDropZone(vert[2], wm.Swap)
    handleRootClick := func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
        // retrieve window
        log.Println("Starting to retrieve window for click")
        win, err := wm.FindUnderMouse(X)
        if err != nil {
            log.Printf("Issues handling root click: %v\n", err)
            return
        }

        // construct utility window for improved logging (?)
        xwin := xwindow.New(X, win)
        _, err = xwin.Geometry()
        if err != nil {
            log.Printf("Cannot get window [%v] geometry: %v\n", xwin, err)
            return
        }

        // WOO WE HAVE A WINDOW HUSTON
        log.Printf("RootWindowClick: touched window %v\n", xwin)

        // the three drag steps, emulated with mousepresses - easier
        if !dm.Dragging {
            dm.StartDrag(xwin)
            return
        }

        if dm.Target == nil {
            dm.SetTarget(xwin)
            return
        }

        // end the drag if we've gotten to here
        err = dm.EndDrag(dz)
        if err != nil {
            log.Printf("RootWindowClick end drag error: %v\n", err)
        } else {
            log.Printf("RootWindowClick: success in drag-ending!\n")
        }

    }
    mousebind.ButtonPressFun(handleRootClick).Connect(X, X.RootWin(), ui.Mod+"-1", false, true)

    // start event loop, even though we have no events
    // to keep app from just closing
    xevent.Main(X)
}
        

