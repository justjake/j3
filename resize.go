package main

/*
Drag handler and support structure to manage resizing a window or
windows(s) in conjunction
*/

import (
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xrect"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/ewmh"

    "github.com/justjake/j3/wm"

    "container/list"
    "fmt"
    "log"
)

func rect_equals(a, b xrect.Rect) bool {
    return a.X() == b.X() && a.Y() == b.Y() && a.Width() == b.Width() && a.Height() == b.Height()
}


// TODO support multi-direction resize with bit masks
func ResizeDirection(X *xgbutil.XUtil, win *xwindow.Window, dir wm.Direction, px int) error {
    // we resize around the decor_geometry of the window

    geom, err := win.Geometry()
    if err != nil {
        return fmt.Errorf("Resize: coudn't get normal geometry: %v", err)
    }
    w, h := geom.Width(), geom.Height()


    if dir == wm.Left || dir == wm.Right {
        // horizontal resize
        w += px
    } else {
        h += px
    }

    // two-step resize -> move process, to compensate for WM peculiarities and window sizing hints
    // first save the initial position info
    pre_decor, err := win.DecorGeometry()
    if err != nil {
        return fmt.Errorf("Resize: coudn't get decorated geometry: %v", err)
    }


    // resize the window
    err = win.WMResize(w, h)
    if err != nil { return err }

    // wait for the geometry to change
    var post_decor xrect.Rect
    post_decor, err = win.DecorGeometry()
    if err != nil {
        return fmt.Errorf("Resize: coudn't get decorated geometry: %v", err)
    }

    // the opposite edge should stay in the same place
    op := dir.Opposite()
    pre_edge := EdgePos(pre_decor, op)
    post_edge := EdgePos(post_decor, op)
    delta := post_edge - pre_edge

    x, y := post_decor.X(), post_decor.Y()

    // move the window upwards by our height resize so that the bottom edge stays in the same place
    if dir == wm.Top || dir == wm.Bottom {
        y -= delta
    }

    // move the window right  by our resize so that the right stays in the same place
    if dir == wm.Left || dir == wm.Right {
        x -= delta
    }

    // move to lock opposite edge
    err = win.WMMove(x, y)
    if err != nil { return err }

    return nil
}

func SideOfRectangle(geom xrect.Rect, x, y int) wm.Direction {
    // construct algebraic functions to delinate the rectangle into sections
    // around the center point like this: [X]
    // these are a little confusing because x11 addresses coordinates from the top-left,
    // where traditional euclidean graphs address from the bottom-left
    w, h := geom.Width(), geom.Height()
    slope := float64(h) / float64(w)
    bl_to_tr_y := int(-1.0 * slope * float64(x)) + h
    tl_to_br_y := int(slope * float64(x))

    var dir wm.Direction

    if x < w/2 {
        // left half of the rectangle
        switch {
        // we must be above both lines and on the left side of the midpoint
        case y <= tl_to_br_y: dir = wm.Top

        // we must be below both lines and on the left side of the midpoint
        case y >= bl_to_tr_y: dir = wm.Bottom

        // we are between the two lines and on the left side
        default: dir = wm.Left
        }
    } else {
        // right half of the rectangle
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

    return dir
}

// return the coordinate part for an edge of a rectangle
// for the top edge, this is just rect.Y(), but for the right edge, it's
// rect.X() + rect.Width() to get the x-offset of the right edge
func EdgePos(rect xrect.Rect, dir wm.Direction) int {
    switch dir {
        case wm.Top:    return rect.Y()
        case wm.Right:  return rect.X() + rect.Width()
        case wm.Bottom: return rect.Y() + rect.Height()
        case wm.Left:   return rect.X()
    }
    log.Panic("Bad direction in EdgePosition")
    return 0
}

// move the incoming window so that it is directly adjacent to the target's edge
func AdjoinEdge(target, incoming *xwindow.Window, dir wm.Direction) error {
    t, err := target.DecorGeometry()
    if err != nil { return err }
    i, err := incoming.DecorGeometry()
    if err != nil { return err }

    delta := EdgePos(t, dir) - EdgePos(i, dir.Opposite())

    if dir == wm.Left || dir == wm.Right {
        return incoming.WMMove(delta + i.X(), i.Y())
    } else {
        return incoming.WMMove(i.X(), delta + i.Y())
    }
    return nil
}


func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

type ResizeDrag struct {
    Window      *xwindow.Window
    Direction   wm.Direction
    Adjacent    *list.List //[]*xwindow.Window  // the windows to resize in the opposite direction
    LastX       int            // original mouse down position
    LastY       int
}


// resize windows with 
func ManageResizingWindows(X *xgbutil.XUtil) {
    // var dragged *Edge

    var DRAG_DATA *ResizeDrag

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

        // create an xwindow.Window so we can get a rectangle to find our bearings from
        xwin := xwindow.New(X, win)
        geom, err := xwin.DecorGeometry()
        if err != nil {
            log.Printf("ResizeStart: geometry error: %v\n", err)
            return false, 0
        }

        // get what side of the rect our mouseclick was on
        x, y := int(reply.WinX), int(reply.WinY)
        dir := SideOfRectangle(geom, x, y)

        // get coordinate part for the edge. this is either X or Y.
        target_edge := EdgePos(geom, dir)

        log.Printf("ResizeStart: on window %v - %v. Direction/edge: %v/%v\n", win, geom, dir, target_edge)

        // find adjacent windows
        adjacent := list.New()

        // note that this is an intellegent request: the WM only gives us a list of visible, normal windows
        // we don't have to worry about moving hidden windows or something
        managed_windows, err := ewmh.ClientListGet(X)
        if err != nil {
            // we can safley ignore this error, because then we just fall back to resizing only this window
            log.Printf("ResizeStart: error getting EWMH client list: %v\n", err)
        } else {
            // select managed windows
            // always enough space
            // TODO: don't grossly overallocate
            for _, candidate_id := range managed_windows {
                // no need to run calculations for ourself!
                if candidate_id == win { continue }

                cand_window := xwindow.New(X, candidate_id)
                cand_geom, err := cand_window.DecorGeometry()
                if err != nil {
                    log.Printf("ResizeStart: couldn't get geometry for ajacency candidate %v: %v\n", candidate_id, err)
                    continue
                }

                cand_edge := EdgePos(cand_geom, dir.Opposite())
                if abs(cand_edge - target_edge) <= AdjacencyEpsilon {
                    // cool, edges are touching.
                    // make sure this window isn't totally above or below the candidate
                    // we do so by constructing a rect using the clicked window's edge
                    // and the candidate's orthagonal dimension
                    // if the rect overlaps, then this window is truly adjacent
                    //
                    // TODO: consider adding a mimumum overlap
                    if dir == wm.Top || dir == wm.Bottom {
                        // measuring X coords
                        if EdgePos(cand_geom, wm.Right) < EdgePos(geom, wm.Left) { continue }
                        if EdgePos(cand_geom, wm.Left) > EdgePos(geom, wm.Right) { continue }
                    } else {
                        if EdgePos(cand_geom, wm.Bottom) < EdgePos(geom, wm.Top) { continue }
                        if EdgePos(cand_geom, wm.Top) > EdgePos(geom, wm.Bottom) { continue }
                    }
                    // if a window has made it to here, it is adgacent!
                    // add it to the list
                    log.Printf("ResizeStart: will resize adjacent window: %v - %v\n", candidate_id, cand_geom)
                    adjacent.PushBack(cand_window)
                }
            }
        }

        // construct the drag data
        data := ResizeDrag{xwin, dir, adjacent, rx, ry}

        DRAG_DATA = &data

        // TODO: finish this
        // create an edge
        // find the adjacent windows
        // start the drag
        return true, 0
    }

    handleResize := func(rx, ry int) {
        delta := rx - DRAG_DATA.LastX
        if DRAG_DATA.Direction == wm.Top || DRAG_DATA.Direction == wm.Bottom {
            delta = ry - DRAG_DATA.LastY
        }

        if DRAG_DATA.Direction == wm.Left || DRAG_DATA.Direction == wm.Top {
            delta = delta * -1
        }

        target_geom, err := DRAG_DATA.Window.DecorGeometry()
        if err != nil {
            log.Printf("Geom retrieve err: %v\n", err)
            return
        }
        target_edge := EdgePos(target_geom, DRAG_DATA.Direction)

        // resize the target by the delta
        err = ResizeDirection(X, DRAG_DATA.Window, DRAG_DATA.Direction, delta)
        if err != nil {
            log.Printf("ResizeStep: can't resize target: %v\n", err)
            return
        }
        // calculate actual delta that occured
        // handles issues with terminal hints
        target_geom_a, err := DRAG_DATA.Window.DecorGeometry()
        if err != nil {
            log.Printf("ResizeStep: Geom retrieve err: %v\n", err)
            return
        }
        target_edge_a := EdgePos(target_geom_a, DRAG_DATA.Direction)
        delta = target_edge_a - target_edge
        if DRAG_DATA.Direction == wm.Left || DRAG_DATA.Direction == wm.Top {
            delta = delta * -1
        }

        
        // resize each adjacent window by the opposite
        for e := DRAG_DATA.Adjacent.Front(); e != nil; e = e.Next() {
            // extract window from the linked list
            adj_win := e.Value.(*xwindow.Window)

            // resize in the opposite direction, with the opposite delta
            // except the delta should be some actual delta calculated from our source window,
            // because issues with terminal windows happen


            err = ResizeDirection(X, adj_win, DRAG_DATA.Direction.Opposite(), -delta)
            if err != nil {
                log.Printf("ResizeStep: can't resize adjacent window %v: %v\n", adj_win, err)
                continue
            }
        }

        // save new coordinates
        DRAG_DATA.LastX = rx
        DRAG_DATA.LastY = ry
    }

    handleDragStep := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
        if DynamicDragResize {
            handleResize(rx, ry)
        }
    }

    handleDragEnd := func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
        // only run on high enough deltas.
        // use the adjacency epsilon here too
        delta := abs(rx - DRAG_DATA.LastX)
        if DRAG_DATA.Direction == wm.Top || DRAG_DATA.Direction == wm.Bottom {
            delta = abs(ry - DRAG_DATA.LastY)
        }

        if delta > AdjacencyEpsilon {
            handleResize(rx, ry)
        }
        DRAG_DATA = nil
    }

    // bind handler
    mousebind.Drag(X, X.RootWin(), X.RootWin(), KeyComboResize, true, 
        handleDragStart, 
        handleDragStep, 
        handleDragEnd)

}
