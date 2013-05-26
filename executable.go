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

    "github.com/justjake/j3/assets"
    "github.com/justjake/j3/ui"
    "github.com/justjake/j3/wm"
)


// CONFIGURATION
// Opt-Shift-LeftMouseButton drags activate j3!
const KeyCombo = ui.KeyOption + "-Shift-1"

const (
    StripBackgroundColor = 0xcccccc
    IconMargin = 15 // space between icons and border
    IconPadding = 25 // space between two icons
)

var (
    // TODO: icons are not square. Make the geometry calculations right!
    // TODO: IconWidth and IconHeight replace IconSize
    IconSize = assets.Swap.Bounds().Dx() // assume square icons
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

func TranslateCoordinatesSync(X *xgbutil.XUtil, src, dest xproto.Window, x, y int) (dest_x, dest_y int, err error) {
    Xx, Xy := int16(x), int16(y)
    cookie := xproto.TranslateCoordinates(X.Conn(), src, dest, Xx, Xy)
    reply, err := cookie.Reply()
    if err != nil {
        return 0, 0, err
    }
    dest_x, dest_y = int(reply.DstX), int(reply.DstY)
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


// reposition all the icons named in `icon_names` in thier window
func configureHorizontalIcons(all_icons map[string]*ui.Icon, icon_names []string,  offsetY int){
    deltaX := IconPadding + IconSize
    offsetX := IconMargin

    for i, name := range icon_names {
        if icon, ok := all_icons[name]; ok {
            icon.Move(offsetX + deltaX * i, offsetY)
            icon.Window.Map()
        }
    }
}

// reposition all the icons named in `icon_names` in thier window
func configureVerticalIcons(all_icons map[string]*ui.Icon, icon_names []string,  offsetX int)  {
    deltaY := IconPadding + IconSize
    offsetY := IconMargin

    for i, name := range icon_names {
        if icon, ok := all_icons[name]; ok {
            icon.Move(offsetX, offsetY + deltaY * i)
            icon.Window.Map()
        }
    }
}


func main() {
    // establish X connection
    X, err := xgbutil.NewConn()
    fatal(err)

    // initiate extension tools
    shape.Init(X.Conn())
    mousebind.Initialize(X)

    // Detail our current window manager. Insures a minimum of EWMH compliance
    wm_name, err := ewmh.GetEwmhWM(X)
    fatal(err)
    log.Printf("Window manager: %s\n", wm_name)

    // Produce visual UI
    cross := makeCross(X)

    // map icons to thier actions
    // this is so we can choose the right action based on what icon the mouse is over
    // when our drag ends
    icons := make(map[string]*ui.Icon, len(assets.Named))
    win_to_action := make(map[xproto.Window]wm.WindowInteraction)
    for name, image := range assets.Named {
        // create icons for each asset image
        icon := ui.NewIcon(X, image, cross.Id)
        icons[name] = icon
        // if we have an action named the same thing as the icon, then link them!
        if action, ok := wm.Actions[name]; ok {
            win_to_action[icon.Window.Id] = action
        } else {
            // otherwise,
            // shade the icon because it has no action attatched to it
            icon.SetState(ui.StateDisabled)
        }
    }

    // correctly position and show the icons on the cross UI
    offset := IconMargin + 2 * (IconSize + IconPadding)

    vert_icons :=  []string{"ShoveTop", "SplitTop", "Swap", "SplitBottom", "ShoveBottom"}
    horiz_icons := []string{"ShoveLeft", "SplitLeft", "Swap", "SplitRight", "ShoveRight"}
    configureVerticalIcons(icons, vert_icons, offset)
    configureHorizontalIcons(icons, horiz_icons, offset)

    
    // show off the cross by making it mouse-draggable, and displaying it!
    // ui.MakeDraggable(X, cross.Id)
    // cross.Map()


    // define handlers for the three parts of any drag-drop operation
    dm := ui.DragManager{}
    handleDragStart := func(X *xgbutil.XUtil, rx, ry, ex, ey int) (cont bool, cursor xproto.Cursor) {
        // find the window we are trying to drag
        win, err := wm.FindManagedWindowUnderMouse(X)
        if err != nil {
            // don't continue the drag
            log.Printf("DragStart: could not get incoming window: %v\n", err)
            return false, 0
        }

        // cool awesome!
        log.Printf("DragStart: starting to drag %v\n", win)
        dm.StartDrag(win)
        // continue the drag
        return true, 0
    }

    handleDragStep := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
        log.Println("DragStep")
        // see if we have a window that ISN'T the incoming window
        win, err := wm.FindManagedWindowUnderMouse(X)
        if err != nil {
            // whatever
            log.Printf("DragStep: no window found or something: %v\n", err)
            return
        }

        // oh we have a window? and it isn't the start window!? And not the current target!?
        if win != dm.Incoming && win != dm.Target {
            // reposition the cross over it
            // TODO: actually do this, center operates on rects, and all I have is this xproto.Window
            dm.SetTarget(win)

            // get the target width/height
            target_geom, err := xwindow.New(X, win).Geometry()
            if err != nil {
                log.Printf("DragStep: issues getting target geometry: %v\n", err)
                return
            }

            // set the target goemetry X, Y to the actual x, y relative to the root window
            tx, ty, err := TranslateCoordinatesSync(X, win, X.RootWin(), 0, 0)
            if err != nil {
                log.Printf("DragStep: issue translating target coordinates to root coordinates: %v\n", err)
                return
            }
            target_geom.XSet(tx)
            target_geom.YSet(ty)
            x, y := centerOver(cross.Geom, target_geom)
            cross.Move(x, y)
            cross.Map()
        }
    }

    handleDragEnd := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
        log.Println("DragEnd")
        exit_early := false
        // get icon we are dropping over
        icon_win, err := wm.FindNextUnderMouse(X, cross.Id)
        if err != nil {
            log.Printf("DragEnd: icon not found: %v\n", err)
            exit_early = true
        }

        incoming, target, err := dm.EndDrag()
        // drag manager produces errors if we don't have both an Incoming and a Target yet
        if err != nil {
            log.Printf("DragEnd: drag manager state error: %v\n", err)
            exit_early = true
        }


        // we tried: hide UI
        cross.Unmap()

        // we had some sort of error, escape!
        if exit_early { return }

        // retrieve the action that this icon indicates
        if action, ok := win_to_action[icon_win]; ok {
            // create util-window objects from our window IDs
            if incoming_id, inc_ok := incoming.(xproto.Window); inc_ok {
                inc_win := xwindow.New(X, incoming_id)
                if target_id, t_ok := target.(xproto.Window); t_ok {
                    t_win := xwindow.New(X, target_id)


                    // perform the action!
                    action(t_win, inc_win)


                } else {
                    log.Println("DragEnd: target type error (was %v)\n", target)
                }
            } else {
                log.Println("DragEnd: incoming type error (was %v)\n", incoming)
            }
        } else {
            log.Printf("DragEnd: couldn't map window %v to an action", icon_win)
        }
    }

    mousebind.Drag(X, X.RootWin(), X.RootWin(), KeyCombo, true, 
        handleDragStart, 
        handleDragStep, 
        handleDragEnd)

    // start event loop, even though we have no events
    // to keep app from just closing
    xevent.Main(X)
}
        

