package wm

import (
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/xrect"

    "github.com/justjake/j3/util"

    "time"
    "fmt"
)

var MoveResizeTimeout = time.Millisecond * 50

// return channel type for WiatForGeometryUpdate
type GeometryUpdate struct {
    Geometry    xrect.Rect
    Error       error
}
// Sometimes window managers are really slow about upating window geometry and position after 
// we issue a resize or move request. This function polls for a change in the window dimensions
//
// this function is kina a horrible hack, and kinda a cool thing about Go.
func PollForGeometryUpdate(win *xwindow.Window, old_geom xrect.Rect, timeout time.Duration,
     updates chan *GeometryUpdate) {
    update := GeometryUpdate{}
    timeout_channel := time.After(timeout)
    for {
        select {
        case <-timeout_channel:
            update.Error = fmt.Errorf("PollForGeometryUpdate: timeout; no geometry change after %v.", timeout)
            updates <-&update
            return
        default:
            //log.Println("default action in WaitForGeometryUpdate")
            new_geom, err := win.DecorGeometry()
            if err != nil {
                update.Error = fmt.Errorf("PollForGeometryUpdate: coudn't get decorated geometry: %v", err)
                updates <- &update
                return
            }
            if !util.RectEquals(new_geom, old_geom) {
                // the window has now changed size
                //log.Println("WaitForGeometryUpdate: did get resize, preparing to send update")
                update.Geometry = new_geom
                updates <- &update
                return
            }
            time.Sleep(time.Millisecond)
        }
    }
}

// simple wrapper around PollForGeometryUpdate that doesn't require goroutine and GeometryUpdate
// unpacking every time
func WaitForGeometryUpdate(win *xwindow.Window, old_geom xrect.Rect, timeout time.Duration) (xrect.Rect, error) {
    // start waiting
    updates := make(chan *GeometryUpdate)
    go PollForGeometryUpdate(win, old_geom, timeout, updates)

    // unpack results
    change := <-updates
    if change.Error != nil { return nil, change.Error }
    return change.Geometry, nil
}

// re-implemented here because under Fluxbox, win.WMMove() results in the window
// growing vertically by the height of the titlebar!
// So we snapshot the size of the window before we move it, 
// move it, compare the sizes, then resize it vertically to be in line with our intentions
//
// this is synchronous: it waits for the window to finish moving before it releases control
// because it would be impossible to selectivley poll for just the move.
func Move(win *xwindow.Window, x, y int) error {
    // snapshot both sorts of window geometries
    decor_geom, err := win.DecorGeometry()
    if err != nil { return err }

    geom, err := win.Geometry()
    if err != nil { return err }

    // move the window, then wait for it to finish moving
    err = win.WMMove(x, y)
    if err != nil { return err }

    post_move, err :=  WaitForGeometryUpdate(win, decor_geom, MoveResizeTimeout)
    if err != nil { return err }

    // compare window widths
    delta_w := decor_geom.Width() - post_move.Width()
    delta_h := decor_geom.Height() - post_move.Height()

    if delta_h != 0 || delta_w != 0 {
        // fluxbox has done it again. We issued a move, and we got a taller window, too!
        log.Printf("Move: resetting dimensions due to w/h delta: %v/%v\n", delta_w, delta_h)
        err = win.WMResize(geom.Width(), geom.Height())
        if err != nil {return err}
        _, err = WaitForGeometryUpdate(win, post_move, MoveResizeTimeout)
        if err != nil {return err}
    }
    return nil
}

// same as above, but moveresize instead of just move at the first step,
// then resize to the provided w/h instead of a snapshotted one
// this implementation differs from Move in that it makes no effort to be end-synchronous
func MoveResize(win *xwindow.Window, x, y, width, height int) error {
    // snapshot only DecorGeometry so we can tell when the move has completed
    pre_move, err := win.DecorGeometry()
    if err != nil {return err}
    // move window then wait...
    err = win.WMMoveResize(x, y, width, height)
    if err != nil {return err}
    _, err = WaitForGeometryUpdate(win, pre_move, MoveResizeTimeout)

    // check inner geometry for correct sizing
    geom, err := win.Geometry()
    if err != nil {return err}

    // this may be inadvisable: what about window hints?
    if geom.Width() != width || geom.Height() != height {
        // something derped! resize to make it right!
        log.Println("MoveResize: resizing again after incorrect new dimensions")
        err = win.WMResize(width, height)
        if err != nil {return err}
    }

    return nil
}

