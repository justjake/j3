package main
// j3
// Mousing assistance for i3 window manage3r

import (
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgb/shape"

    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/mousebind"

    "log"

    "github.com/justjake/j3/assets"
    "github.com/justjake/j3/ui"
    "github.com/justjake/j3/wm"
    "github.com/justjake/j3/util"
)



// CONFIGURATION //////////////////////////////////////////////////////////////
// feel free to change the values here to customize j3

const (
    // Opt-Shift-LeftMouseButton drags activate j3!
    // MoveKeyCombo is the binding used for j3's window movement functions
    // it is in the format (MOD_NAME-)+[1-5]
    // where MOD_NAME is any X11 keyname, and 1-5 is the mouse button number
    MoveKeyCombo = ui.KeyOption + "-Shift-1"

    // Look-and-feel options
    BackgroundColor = 0x262626  // in hexadecimal #ff00ff style
    IconMargin = 15 // space between icons and border
    IconPadding = 25 // space between two icons
)

///////////////////////////////////////////////////////////////////////////////


var (
    // TODO: icons are not square. Make the geometry calculations right!
    // TODO: IconWidth and IconHeight replace IconSize
    IconSize = assets.Swap.Bounds().Dx() // assume square icons
)

func makeCross(X *xgbutil.XUtil) *ui.Cross {
    // create a basic cross. We will have to initalize the window later.
    cross_ui := ui.NewCross(assets.Named, IconSize, IconMargin, IconPadding)
    vert_icons :=  []string{"ShoveTop", "SplitTop", "Swap", "SplitBottom", "ShoveBottom"}
    horiz_icons := []string{"ShoveLeft", "SplitLeft", "Swap", "SplitRight", "ShoveRight"}

    _, err := cross_ui.CreateWindow(X, len(vert_icons), BackgroundColor)
    util.Fatal(err)

    // position icons on the cross
    offset := IconMargin + 2 * (IconSize + IconPadding)
    cross_ui.LayoutHorizontalIcons(horiz_icons, offset)
    cross_ui.LayoutVerticalIcons(vert_icons, offset)

    return cross_ui
}



func main() {

    // I don't want to retype all of these things
    // TODO: find/replace fatal with util.Fatal
    fatal := util.Fatal

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

    // create the cross UI
    cross_ui := makeCross(X)
    cross := cross_ui.Window

    // map the icons on the cross the the actions they should perform 
    // when objects are dropped over them
    win_to_action := make(map[xproto.Window]wm.WindowInteraction)
    for name, icon := range cross_ui.Icons {
        if action, ok := wm.Actions[name]; ok {
            win_to_action[icon.Window.Id] = action
        } else {
            // otherwise,
            // shade the icon because it has no action attatched to it
            icon.SetState(ui.StateDisabled)
        }
    }

    // define handlers for the three parts of any drag-drop operation
    dm := util.DragManager{}
    handleDragStart := func(X *xgbutil.XUtil, rx, ry, ex, ey int) (cont bool, cursor xproto.Cursor) {
        // find the window we are trying to drag
        win, err := wm.FindManagedWindowUnderMouse(X)
        if err != nil {
            // don't continue the drag
            log.Printf("DragStart: could not get incoming window: %v\n", err)
            return false, 0
        }

        // cool awesome!
        dm.StartDrag(win)
        // continue the drag
        return true, 0
    }

    handleDragStep := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
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
            tx, ty, err := wm.TranslateCoordinatesSync(X, win, X.RootWin(), 0, 0)
            if err != nil {
                log.Printf("DragStep: issue translating target coordinates to root coordinates: %v\n", err)
                return
            }
            target_geom.XSet(tx)
            target_geom.YSet(ty)
            x, y := util.CenterOver(cross.Geom, target_geom)
            cross.Move(x, y)
            cross.Map()
        }
    }

    handleDragEnd := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
        exit_early := false
        // get icon we are dropping over
        icon_win, _, err := wm.FindNextUnderMouse(X, cross.Id)
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

    mousebind.Drag(X, X.RootWin(), X.RootWin(), MoveKeyCombo, true, 
        handleDragStart, 
        handleDragStep, 
        handleDragEnd)

    ///////////////////////////////////////////////////////////////////////////
    // Window resizing behavior spike
    ManageResizingWindows(X)

    // start event loop, even though we have no events
    // to keep app from just closing
    xevent.Main(X)
}
        

