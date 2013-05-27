package ui

/*
Drawing the cross is a somewhat complicated and involved process.
First, we create two rects (vertical and horizontal cross pieces) based on
the size of whatever icons we will be putting on the cross, and our 
icon margin/padding padding values.

Then, we create a window that covers the bounds of both rects using Bounds()
and xwindow.Create.

Finally, we compose the cross window from our rect geometries using ComposeShape.
*/

import (
    "github.com/BurntSushi/xgb/xproto"
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xwindow"
    "github.com/BurntSushi/xgbutil/xrect"

    "github.com/justjake/j3/util"

    "errors"
    "image"
    "image/color"
    "log"
)

// convert an RGB color specified as a unsiged 32-bit integer hex number, eg 0xff00ff
// into an RGBA color
func RGB(c uint32) (color.RGBA) {
    r := uint8((c & 0xff0000) / 0xffff)
    g := uint8((c & 0x00ff00) / 0xff)
    b := uint8( c & 0x0000ff)
    return color.RGBA{r, g, b, 0xff}
}

func centerChild(child, parent xrect.Rect) (x, y int) {
    a := child
    b := parent
    b_center_x := b.Width() / 2
    b_center_y := b.Height() / 2

    x = b_center_x - a.Width() / 2
    y = b_center_y - a.Height() / 2
    return
}

// the Cross (+) ui: two batches of icons indersecting at a 90* angle.
// a large part of what j3 is about
type Cross struct {
    Icons           map[string]*Icon
    Window          *xwindow.Window
    IconSize        int
    IconMargin      int
    IconPadding     int
    imagesToBecomeIcons map[string]image.Image
}

func NewCross(icons map[string]image.Image, size, margin, padding int) (*Cross) {
    // now time to create the window!
    cross := Cross{nil, nil, size, margin, padding, icons}
    log.Printf("New Cross created: %v\n", cross)
    return &cross
}

// When a cross is declared in its object literal form, it may not have the appropriate window.
// this function creates a new X11 window for the cross with the correct geometry depending 
// on its Icon* parameters.
func (c *Cross) CreateWindow(X *xgbutil.XUtil, icons_per_direction int, bg_color uint32) (*xwindow.Window, error) {

    // calculate the dimensions of the spars of our cross +
    // width/height reflect the vertical-orientation rectangle
    width := c.IconMargin * 2 + c.IconSize
    // padding between the icons, margin on the edges
    height := c.IconSize * icons_per_direction + c.IconPadding * (icons_per_direction-1) + c.IconMargin * 2


    // intitialize a basic window for the cross
    win, err := xwindow.Generate(X)
    if err != nil { return nil, err }

    win.Create(X.RootWin(), 0, 0, height, height, 
        xproto.CwBackPixel | xproto.CwOverrideRedirect, bg_color, 1)


    // the rects we will be adding together to form the shape of the cross
    vert := xrect.New(0, 0, width, height)
    horiz := xrect.New(0, 0, height, width)
    struts := []xrect.Rect{vert, horiz}

    geom, err := win.Geometry()
    if err != nil { return nil, err }

    // center struts over window
    x, y := util.CenterChild(vert, geom)
    vert.XSet(x)
    vert.YSet(y)
    x, y = util.CenterChild(horiz, geom)
    horiz.XSet(x)
    horiz.YSet(y)

    // build the cross shape from our friendly rectangles
    err = ComposeShape(X, win.Id, struts)
    if err != nil { return nil, err }

    // add the window to our cross struct
    c.Window = win

    // create icons from our images
    clr := RGB(bg_color)
    if c.imagesToBecomeIcons != nil {
        icons := make(map[string]*Icon, len(c.imagesToBecomeIcons))
        for name, img := range c.imagesToBecomeIcons {
            icon := NewIcon(X, img, win.Id)
            icon.Background = clr
            icons[name] = icon
        }
        c.Icons = icons
    } else {
        return nil, errors.New("Cross: you must create crosses using the NewCross function (this cross has now iconsToBecomeImage)")
    }

    return win, nil
}

// show the appropraite icons in the correct positions
func (c *Cross) LayoutHorizontalIcons(icon_names []string, offsetY int) {
    deltaX := c.IconPadding + c.IconSize
    offsetX := c.IconMargin

    for i, name := range icon_names {
        if icon, ok := c.Icons[name]; ok {
            icon.Move(offsetX + deltaX * i, offsetY)
            icon.Window.Map()
        } else {
            log.Printf("Skipping image '%v': not found in icon store %v\n", name, c.Icons)
        }
    }
}

// Icons named in `icon_names` are positioned across the horizontal spar in the
// order they were named. Icons that are not in c.Icons are skipped, leaving
// thier space blank. In such a manner you can leave blank spaces by passing 
// strings like "HurrDurr" in that position
func (c *Cross) LayoutVerticalIcons(icon_names []string, offsetX int) {
    deltaY := c.IconPadding + c.IconSize
    offsetY := c.IconMargin

    for i, name := range icon_names {
        if icon, ok := c.Icons[name]; ok {
            icon.Move(offsetX, offsetY + deltaY * i)
            icon.Window.Map()
        } else {
            log.Printf("Skipping image '%v': not found in icon store %v\n", name, c.Icons)
        }
    }
}
