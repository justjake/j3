package wm

/* Move.go
   handles moving windows about using EWMH interaction commands
   I think we can usually just use xwindow.Window objects for convinience
   */
import (
    "github.com/BurntSushi/xgbutil/xwindow"
    logLib "log"
    "os"
)

var log = logLib.New(os.Stderr, "[window manager] ", logLib.LstdFlags | logLib.Lshortfile)

type WindowInteraction func(*xwindow.Window, *xwindow.Window) (error)

// Split functions:
// split the target window's area in half. The incoming window takes the split
// half in the direction of the function name. So in SpitTop, the target is the
// bottom half of the initial area, and the incoming is the top.

// cutting windows in half on the X axis

var Actions = map[string]WindowInteraction{
    "SplitTop"    :  SplitTop,
    "SplitRight"  :  SplitRight,
    "SplitBottom" :  SplitBottom,
    "SplitLeft"   :  SplitLeft,

    "ShoveTop"    :  ShoveTop,
    "ShoveRight"  :  ShoveRight,
    "ShoveBottom" :  ShoveBottom,
    "ShoveLeft"   :  ShoveLeft,

    "Swap"  :  Swap,
}

func splitVertical(target, incoming *xwindow.Window, incomingOnTop bool) error {
    bounds, err := target.DecorGeometry()
    if err != nil {
        log.Printf("splitVertical: error getting bounds of target: %v\n", err)
        return err
    }

    bottom_height := bounds.Height() / 2
    top_height := bounds.Height() - bottom_height

    var top, bottom *xwindow.Window
    if incomingOnTop {
        top = incoming
        bottom = target
    } else {
        top = target
        bottom = incoming
    }

    // target goes on bottom...
    err = MoveResize(bottom, bounds.X(), bounds.Y() + top_height, bounds.Width(), bottom_height)
    if err != nil {
        log.Printf("splitVertical: error configuring bottom: %v\n", err)
        return err
    }

    // and incoming on top
    err = MoveResize(top, bounds.X(), bounds.Y(), bounds.Width(), top_height)
    if err != nil {
        log.Printf("splitVertical: error configuring top: %v\n", err)
        return err
    }

    // cool
    return nil
}

// cutting windows in half on the Y axis
func splitHorizontal(target, incoming *xwindow.Window, incomingOnLeft bool) error {
    bounds, err := target.DecorGeometry()
    if err != nil {
        log.Printf("splitHorizontal: error getting bounds of target: %v\n", err)
        return err
    }

    left_width := bounds.Width() / 2
    right_width := bounds.Width() - left_width

    var left, right *xwindow.Window
    if incomingOnLeft {
        left = incoming
        right = target
    } else {
        left = target
        right = incoming
    }

    err = MoveResize(right, bounds.X() + left_width, bounds.Y(), right_width, bounds.Height())
    if err != nil {
        log.Printf("splitHorizontal: error configuring right: %v\n", err)
        return err
    }

    err = MoveResize(left, bounds.X(), bounds.Y(), left_width, bounds.Height())
    if err != nil {
        log.Printf("splitHorizontal: error configuring left: %v\n", err)
        return err
    }

    // cool
    return nil
}

// Exported split actions

// Split the target window, putting the incoming window in the top half
func SplitTop(target, incoming *xwindow.Window) error {
    return splitVertical(target, incoming, true)
}
// Split the target window, putting the incoming window in the bottom half
func SplitBottom(target, incoming *xwindow.Window) error {
    return splitVertical(target, incoming, false)
}
// Split the target window, putting the incoming window in the left half
func SplitLeft(target, incoming *xwindow.Window) error {
    return splitHorizontal(target, incoming, true)
}
// Split the target window, putting the incoming window in the right half
func SplitRight(target, incoming *xwindow.Window) error {
    return splitHorizontal(target, incoming, false)
}

// Swap the position and size of the target and incoming windows
func Swap(target, incoming *xwindow.Window) error {
    // get bounds for both windows
    target_bounds, err := target.DecorGeometry()
    if err != nil {
        log.Printf("Swap: error getting bounds of target: %v\n", err)
        return err
    }
    incoming_bounds, err := incoming.DecorGeometry()
    if err != nil {
        log.Printf("Swap: error getting bounds of incoming: %v\n", err)
        return err
    }

    // configure windows, easy as pie!
    err = MoveResize(target, incoming_bounds.X(), incoming_bounds.Y(), 
        incoming_bounds.Width(), incoming_bounds.Height())
    if err != nil {
        log.Printf("Swap: error configuring target: %v\n", err)
        return err
    }
    err = MoveResize(incoming, target_bounds.X(), target_bounds.Y(),
        target_bounds.Width(), target_bounds.Height())
    if err != nil {
        log.Printf("Swap: error configuring incoming: %v\n", err)
        return err
    }

    // cool
    return nil
}

type Direction uint8
const (
    Top Direction = 1 << iota
    Right
    Bottom
    Left
)

// lol a string method
// TODO: support directions that have been masked together to form things like TopLeft, etc
func (d Direction) String() string {
    if d == Top { return "Top" }
    if d == Right { return "Right" }
    if d == Bottom { return "Bottom" }
    if d == Left { return "Left" }
    return "Unknown Direction"
}

func (d Direction) Opposite() Direction {
    switch d {
    case Top: return Bottom
    case Bottom: return Top
    case Left: return Right
    case Right: return Left
    }
    log.Panic("Direction.Opposite: Unreachable")
    return 0
}

// Put the incoming window on the `dir` side of the target,
// and transform the orthagonal dimension (eg, if `dir` is Up, then dim is `Width`
// to be the same as the target's dimension
// TODO: clip windows to display boundry
func Shove(target, incoming *xwindow.Window, dir Direction) error {
    // get geometries
    i, err := incoming.DecorGeometry()
    if err != nil { return err }

    t, err := target.DecorGeometry()
    if err != nil { return err }


    // move in the correct direction
    if dir == Top {
        err := MoveResize(incoming, t.X(), t.Y() - i.Height(), t.Width(), i.Height())
        if err != nil { return err }
    }

    if dir == Bottom {
        err := MoveResize(incoming, t.X(), t.Y() + t.Height(), t.Width(), i.Height())
        if err != nil { return err }
    }

    if dir == Left {
        err := MoveResize(incoming, t.X() - i.Width(), t.Y(), i.Width(), t.Height())
        if err != nil { return err }
    }

    if dir == Right {
        err := MoveResize(incoming, t.X() + t.Width(), t.Y(), i.Width(), t.Height())
        if err != nil { return err }
    }

    return nil
}

// see Shove
func ShoveTop(t, i *xwindow.Window) error {
    return Shove(t, i, Top)
}
func ShoveRight(t, i *xwindow.Window) error {
    return Shove(t, i, Right)
}
func ShoveBottom(t, i *xwindow.Window) error {
    return Shove(t, i, Bottom)
}
func ShoveLeft(t, i *xwindow.Window) error {
    return Shove(t, i, Left)
}


// TODO: all of the above in a tiling context
// Shoves are basically just splits, but performed one-level-up, on a window's parent
// THe current splits implementation works only for floating window managers,
// which don't have crazy-cray nesting stuff
// so shoves don't yet have any meaning.
