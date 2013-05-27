package util

type Rect interface {
    X() int
    Y() int
    Width() int
    Height() int
}

// return the correct X, Y to center rect A over rect B
func CenterOver(a, b Rect) (x, y int) {
    b_center_x := b.X() + b.Width() / 2
    b_center_y := b.Y() + b.Height() / 2

    x = b_center_x - a.Width() / 2
    y = b_center_y - a.Height() / 2
    return
}

// center a rectangle over a parent rectagle, as though the child were in a
// coordiante space that originated form `parent`'s top-left corner
func CenterChild(child, parent Rect) (x, y int) {
    a := child
    b := parent
    b_center_x := b.Width() / 2
    b_center_y := b.Height() / 2

    x = b_center_x - a.Width() / 2
    y = b_center_y - a.Height() / 2
    return
}


