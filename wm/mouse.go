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

/* Mouse query functions */
import (
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/ewmh"

    "errors"
    "fmt"
)

// translate an X, Y value from one window to the euclidean space of another. Useful for translating
// the arbitrary position of some window's (0,0) point into root window space
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

// wrapper around xproto.QueryPointer that performs a simple synchronous query
func FindNextUnderMouse(X *xgbutil.XUtil, parent xproto.Window) (xproto.Window, *xproto.QueryPointerReply, error) {
    // start query pointer request
    cookie := xproto.QueryPointer(X.Conn(), parent)

    // block and get reply for client
    reply, err := cookie.Reply()
    if err != nil {
        return 0, nil, err
    }
    return reply.Child, reply, nil
}

// find the EWHM window under the mouse cursor
func FindManagedWindowUnderMouse(X *xgbutil.XUtil) (xproto.Window, error) {
    // construct a hashset of the managed windows
    clients, err := ewmh.ClientListGet(X)
    if err != nil {
        return 0, fmt.Errorf("FindUnderMouse: could not retrieve EWHM client list: %v", err)
    }

    managed := make(map[xproto.Window]bool, len(clients))
    for _, win := range clients {
        managed[win] = true
    }

    cur_window := X.RootWin()

    // descend the QueryTree to the first child that is a EWMH managed window
    for {
        // return the parent if it is an EWHM window
        if _, ok := managed[cur_window]; ok {
            return cur_window, nil
        }

        cur_window, _, err = FindNextUnderMouse(X, cur_window)
        if err != nil {
            break
        }
    }

    // we didn't find the window :(
    return 0, errors.New("FindUnderMouse: no EWMH window found under mouse")
}

func FindWindowUnderMouse(X *xgbutil.XUtil, orig_window *xproto.Window) (xproto.Window, error) {
    var cur_window xproto.Window = 0
    for {
        cur_window, _, err := FindNextUnderMouse(X, cur_window)
        if err != nil {
            log.Printf("FindUnderMouse: deep query error: %v\n", err)
            if cur_window != 0 {
                return cur_window, nil
            } else {
                return 0, err
            }
        }
    }

    return 0, errors.New("FindWindowUnderMouse: unreachable error")
}



