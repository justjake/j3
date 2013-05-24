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

// compose a number of rectabgles into a window shape
func ComposeShape(X *xgbutil.XUtil, dst xproto.Window, rects []xrect.Rect) (err error) {
    log.Println("Constructing shape.")

    combine_bounds := make([]shape.CombineCookie, len(rects))
    combine_clip   := make([]shape.CombineCookie, len(rects))

    for i, rect := range rects {
        // make rectangular window of correct goemetry
        win, err := xwindow.Generate(X)
        if err != nil {
            log.Fatalf("ComposeShape: Error creating rectange %v window.", rect)
            return err
        }
        win.Create(X.RootWin(), rect.X(), rect.Y(), rect.Width(), rect.Height(), xproto.CwBackPixel, 0xffffff)
        log.Println("did create window")

        // combine window request
        combine_kind := shape.Kind(shape.SkBounding)
        combine_bounds[i] = shape.CombineChecked(X.Conn(), shape.SoUnion, combine_kind, combine_kind, dst, 0, 0, win.Id)
        combine_kind = shape.Kind(shape.SkClip)
        combine_clip[i] = shape.CombineChecked(X.Conn(), shape.SoUnion, combine_kind, combine_kind, dst, 0, 0, win.Id)
    }
    return nil
}
    
