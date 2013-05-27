// Create windows from polagonal shapes

package ui

import (
    "github.com/BurntSushi/xgb/shape"
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/xrect"

    "log"
)

// extract the top-left and bottom-right points of an xrect as a 4-tuple:  x, y, x2, y2
func coords(rect xrect.Rect) (min_x, min_y, max_x, max_y int) {
    min_x = rect.X()
    max_x = min_x + rect.Width()

    min_y = rect.Y()
    max_y = min_y + rect.Height()
    return
}

func min(a, b int) int { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }

// given a slice of rects, return a rect that covers all of them!
func Bound(rects []xrect.Rect) (xrect.Rect) {
    min_x, min_y, max_x, max_y := coords(rects[0])
    for _, rect := range rects[1:] {
        x1, y1, x2, y2 := coords(rect)
        min_x = min(x1, min_x)
        min_y = min(y1, min_y)
        max_x = max(x2, max_x)
        max_y = max(y2, max_y)
    }
    return xrect.New(min_x, min_y, max_x - min_x, max_y - min_y)
}

// compose a number of rectabgles into a window shape
func ComposeShape(X *xgbutil.XUtil, dst xproto.Window, rects []xrect.Rect) (err error) {

    combine_bounds := make([]shape.CombineCookie, len(rects))
    combine_clip   := make([]shape.CombineCookie, len(rects))

    var operation shape.Op

    for i, rect := range rects {
        // make rectangular window of correct goemetry
        win, err := xwindow.Generate(X)
        if err != nil {
            log.Fatalf("ComposeShape: Error creating rectange %v window.", rect)
            return err
        }
        win.Create(X.RootWin(), rect.X(), rect.Y(), rect.Width(), rect.Height(), xproto.CwBackPixel, 0xffffff)

        // choose operation. on the first one, we want to set the shape.
        if i == 0 {
            operation = shape.SoSet
        } else {
            operation = shape.SoUnion
        }

        // combine window request
        x, y := int16(rect.X()), int16(rect.Y())

        combine_kind := shape.Kind(shape.SkBounding)
        combine_bounds[i] = shape.CombineChecked(X.Conn(), operation, combine_kind, combine_kind, dst, x, y, win.Id)
        combine_kind = shape.Kind(shape.SkClip)
        combine_clip[i] = shape.CombineChecked(X.Conn(), operation, combine_kind, combine_kind, dst, x, y, win.Id)
    }
    return nil
}
    
