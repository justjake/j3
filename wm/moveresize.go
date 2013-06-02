package wm

import (
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/xrect"

    "github.com/justjake/j3/util"

    "time"
    "fmt"
)

var MoveResizeTimeout = time.Millisecond * 30

// return channel type for WiatForGeometryUpdate
type GeometryUpdate struct {
    Base        xrect.Rect
    Geometry    xrect.Rect
    Error       error
}

type TimeoutError struct {
    // generic error message field
    Message string
    // how long we waited before producing this error
    Timeout time.Duration
}

func (err *TimeoutError) Error() string {
    return fmt.Sprintf("%s: timeout after %v.", err.Message, err.Timeout)
}

// A funciton for use with PollFor. Returning 'true' indicates success:
// we are done polling without any errors.
// Returning an error will immediatly stop polling, but the error will
// be returned by PollFor for handling.
type GeometryUpdateTester func(*xwindow.Window) (bool, error)

// Returns True once the window's DecorGeometry changes
// Use with PollFor.
func DecorDiffers(oldDecor xrect.Rect) GeometryUpdateTester {
    return func(win *xwindow.Window) (bool, error) {
        newDecor, err := win.DecorGeometry()
        if err != nil { return false, err }
        return (!util.RectEquals(oldDecor, newDecor)), nil
    }
}

// Returns True once the window's Geometry changes
// Use with PollFor.
func GeometryDiffers(oldGeom xrect.Rect) GeometryUpdateTester {
    return func(win *xwindow.Window) (bool, error) {
        newGeom, err := win.Geometry()
        if err != nil { return false, err }
        return (!util.RectEquals(oldGeom, newGeom)), nil
    }
}


// The previous PollForGeometryUpdate function and friends were very complex, and a bit brittle
// PollFor and the GeometryUpdateTester generator functions GeometryDiffers and DecorDiffers
// do the same job (polling for some change on a *xwindow.Window), usually a geometry change
//
// PollFor runs each GeometryUpdateTester in order until all return true *on the same run*. 
// If any GeometryUpdateTester throws an error, PollFor stops and returns that error.
// On a successful polling, PollFor returns 'nil' as its error
func PollForTimeout(win *xwindow.Window, timeout time.Duration, changes ...GeometryUpdateTester) error {
    timeout_channel := time.After(timeout)

    for {
        select {
        case <-timeout_channel:
            return &TimeoutError{"PollFor", timeout}
        default:
            // run each geometry test predicate
            should_exit := true
            for i, pred := range changes {
                exit, err := pred(win)
                if err != nil {
                    return fmt.Errorf("PollFor: error in predicate %d: %v", i, err)
                }
                should_exit = should_exit && exit
            }

            // exit when all pass
            if should_exit {
                return nil
            }

            // don't spam the server
            time.Sleep(time.Millisecond)
        }
    }

    return fmt.Errorf("PollFor: should be unreachable")
}

// Same as PollForTimeout, except uses the default timeout.
func PollFor(win *xwindow.Window, change_predicates ...GeometryUpdateTester) error {
    return PollForTimeout(win, MoveResizeTimeout, change_predicates...)
}


// Sometimes window managers are really slow about 
// re-implemented here because under Fluxbox, win.WMMove() results in the window
// growing vertically by the height of the titlebar!
// So we snapshot the size of the window before we move it, 
// move it, compare the sizes, then resize it vertically to be in line with our intentions
//
// this is synchronous: it waits for the window to finish moving before it releases control
// because it would be impossible to selectivley poll for just the move.
func Move(win *xwindow.Window, x, y int) error {
    // snapshot both sorts of window geometries
    decor_geom, geom, err := Geometries(win)
    if err != nil { return err }
    log.Printf("Move: detected geometry to be %v\n", geom)

    // move the window, then wait for it to finish moving
    err = win.WMMove(x, y)
    if err != nil { return err }

    // this waits 30MS under non-Fluxbox window manager
    // WHAT DO
    err = PollFor(win, GeometryDiffers(geom))
    if err != nil {
        // if we had a timeout, that means that the geometry didn't derp during
        // moving, and everything is A-OK!
        // skip the rest of the function
        if _, wasTimeout := err.(*TimeoutError); wasTimeout {
            return nil
        }
        return err
    }

    // compare window widths before/after move
    _, post_move_base, err := Geometries(win)
    if err != nil { return err }

    delta_w := post_move_base.Width() - geom.Width()
    delta_h := post_move_base.Height() - geom.Height()

    if delta_h != 0 || delta_w != 0 {
        // fluxbox has done it again. We issued a move, and we got a taller window, too!
        log.Printf("Move: resetting dimensions to %v due to w/h delta: %v/%v\n", geom, delta_w, delta_h)
        err = win.WMResize(geom.Width(), geom.Height())
        if err != nil {return err}

        // wait for that to succeed
        err = PollFor(win, GeometryDiffers(post_move_base))
        if err != nil {return err}
    }
    // make sure window did actually move
    err = PollFor(win, DecorDiffers(decor_geom))
    if err != nil {
        // if we had a timeout, that means that the window didn't move
        // we want to send an error mentioning that fact specifically
        // instead of a generic "lol timeout happan in polling :DDD"
        if te, wasTimeout := err.(*TimeoutError); wasTimeout {
            return &TimeoutError{"Move: window didn't move", te.Timeout}
        }
        // return whatever other error stymied the polling
        return err
    }
    return nil
}

// same as above, but moveresize instead of just move at the first step,
// then resize to the provided w/h instead of a snapshotted one
// this implementation differs from Move in that it makes no effort to be end-synchronous
// This function waits only on the window's inner geometry resizing, not on actual movement occuring
func MoveResize(win *xwindow.Window, x, y, width, height int) error {
    // snapshot window dimensions
    base, err := win.Geometry()
    if err != nil { return err }

    // move window then wait...
    err = win.WMMoveResize(x, y, width, height)
    if err != nil {return err}
    err = PollFor(win, GeometryDiffers(base))
    if err != nil {return err}

    // check that the new geometry is what we requested
    // this may be inadvisable: what about window hints?
    geom, err := win.Geometry()
    if err != nil {return err}
    if geom.Width() != width || geom.Height() != height {
        // something derped! resize to make it right!
        // if window hints constrained us, this won't upset them
        log.Println("MoveResize: resizing again after incorrect new dimensions")
        err = win.WMResize(width, height)
        if err != nil {return err}
    }

    return nil
}

func Geometries(win *xwindow.Window) (xrect.Rect, xrect.Rect, error) {
    decor, err := win.DecorGeometry()
    if err != nil { return nil, nil, err }
    base, err := win.Geometry()
    if err != nil { return nil, nil, err }
    return decor, base, nil
}
