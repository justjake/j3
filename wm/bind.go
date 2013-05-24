package wm

/*
bind.go handles action binding and state management for j3
The job of this file is to
    * note if a drag is occuring
    * track the current incoming (drag-start) window
    * deduce the current target (moused-over) window
    * move the cross over the target
    * handle drop-actions on the cross's window interaction icons
*/

import (
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/ewmh"

    "github.com/justjake/j3/ui"
    "errors"
    "fmt"
)
    

// we will eventually build the cross out of DropZones instead of bare icons
// DropZones link the presentaion element (the icon) with a window action
type DropZone struct {
    Icon      *ui.Icon 
    Action    WindowInteraction
}
// we can get the window under the mouse from X11
// we use that information to find the DropZone, and thus the action,
// from the icon's window
var windowToDropZone map[xproto.Window]*DropZone = make(map[xproto.Window]*DropZone, 10)

func NewDropZone(icn *ui.Icon, action WindowInteraction) *DropZone {
    dz := DropZone{icn, action}
    id := icn.Window.Id
    windowToDropZone[id] = &dz
    return &dz
}

// find the window directly under the mosue cursor at this time
// synchronous
// TODO: reconsider using QueryPointer directly for dat async performance
func FindUnderMouse(X *xgbutil.XUtil) (xproto.Window, error) {
    // start query pointer request
    cookie := xproto.QueryPointer(X.Conn(), X.RootWin())

    // construct a hashset of the managed windows
    clients, err := ewmh.ClientListGet(X)
    if err != nil {
        return 0, fmt.Errorf("FindUnderMouse: could not retrieve EWHM client list: %v", err)
    }
    managed := make(map[xproto.Window]bool, len(clients))
    for _, win := range clients {
        managed[win] = true
    }

    // block and get reply for client
    reply, err := cookie.Reply()
    if err != nil {
        return 0, err
    }
    child := reply.Child

    // ingore the rest for now
    return child, nil

    root  := X.RootWin()

    for child != root {
        // test to see if child is a managed window
        _, ok := managed[child]
        if ok {
            return child, nil
        }

        // traverse upwards
        tree, err := xproto.QueryTree(X.Conn(), child).Reply()
        if err != nil {
            return 0, fmt.Errorf("FindUnderMouse: tree traversal error: %v", err)
        }
        child = tree.Parent
    }
    // we didn't find the window :(
    return 0, fmt.Errorf("FindUnderMouse: window under mouse %v not found in EWMH clients %v", reply.Child, clients)
}

// meat and potatoes of our window manager
// tracks drags, invokes drops
type DragManager struct {
    Dragging bool
    Target   *xwindow.Window
    Incoming *xwindow.Window

    // actuals
    cross    *xwindow.Window
    x        *xgbutil.XUtil
}


func NewDragManager(X *xgbutil.XUtil, cross *xwindow.Window) *DragManager {
    dm := DragManager{
        Dragging: false,
        Target: nil,
        Incoming: nil,
        cross: cross,
        x: X,
    }
    return &dm
}
    

func (dm *DragManager) StartDrag(incoming *xwindow.Window) {
    dm.Dragging = true
    dm.Incoming = incoming

    // TODO: listen harder?
}

func (dm *DragManager) SetTarget(target *xwindow.Window) {
    dm.Target = target
}

// this needs significant reconsideration
func (dm *DragManager) EndDrag(dz *DropZone) error {
    // end the drag no matter what
    t, i := dm.Target, dm.Incoming
    dm.Target, dm.Incoming = nil, nil

    if !dm.Dragging {
        return errors.New("Cannot end drag: not currently dragging")
    }
    dm.Dragging = false

    if i == nil {
        return errors.New("Cannot end drag: no incoming window")
    }

    if t == nil {
        return errors.New("Cannot end drag: no target window")
    }

    // invoke the drag zone action
    return dz.Action(t, i)
}
