// create a correctly-sized window from an image

package ui

import (
    "github.com/BurntSushi/xgb/xproto"

    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xgraphics"
    "github.com/BurntSushi/xgbutil/xwindow"

    "log"
    "image"
    "image/color"
)

type IconState int
const (
    StateNormal = iota
    // the icon is blended against the background at 50% opacity
    StateDisabled
    // the icon is unmapped as a window
    StateHidden

    // opacity multiplier
    iconDisabledShade = 0.35
)

type Icon struct {
    Image   *fadeImage          // original icon, from PNG usually, inside a fader
    Parent  xproto.Window       // parent X window of the icon
    Window  *xwindow.Window     // window that the icon is drawn to
    ximage  *xgraphics.Image    // X11 image swap buffer, used to paint Image to Window
    state   IconState
    Background color.Color
}

// Create a new Icon from an image, with a given X11 window parent
func NewIcon(X *xgbutil.XUtil, img image.Image, parent xproto.Window) *Icon {
    ximg := xgraphics.NewConvert(X, img)
    win := ximg.Window(parent)
    fader := fadeImage{img, 1.0}
    icn := Icon{&fader, parent, win, ximg, StateNormal, color.RGBA{0xcc, 0xcc, 0xcc, 0xff}}
    return &icn
}

// get the pixels directly behind the icon
// right now just returns an image.Uniform hard-coded to the right shade of grey
// TODO actually do this
func (icn *Icon) getBackground() image.Image {
    log.Println("TODO: Icon.Blend()")
    bg_color := icn.Background
    bg := image.NewUniform(bg_color)
    return bg
}


// blend the RGBA original icon against the background of Icon.Parent
// Allows us to simulate transparency
func (icn *Icon) Blend() {
    // get pixels "parent-pixels" of Parent that are behind Window
    bg := icn.getBackground()

    // copy "parent-pixels" into buffer "ximage", overwriting existing completely
    xgraphics.Blend(icn.ximage, bg, image.Point{0,0})

    // alpha-blend Image into buffer "ximage"
    xgraphics.Blend(icn.ximage, icn.Image, image.Point{0,0})

    // swap ximage into Window as background
    icn.ximage.XSurfaceSet(icn.Window.Id)
    icn.ximage.XDraw()
    icn.ximage.XPaint(icn.Window.Id)

    // free the pixbuff memory!
    icn.ximage.Destroy()
}

// Wrapper around xwindow.Window.Move that blends the icon post-move
func (icn *Icon) Move(x, y int) {
    icn.Window.Move(x, y)
    icn.Blend()
}

// change image states
func (icn *Icon) SetState(newstate IconState) {
    if newstate == icn.state {
        // do nothing
        return
    }

    // if we're switching away from StateHidden, then show the icon
    if icn.state == StateHidden {
        icn.Window.Map()
    }

    // fade out disabled icons
    if newstate == StateDisabled {
        icn.Image.Factor = iconDisabledShade
    }

    if newstate == StateNormal {
        icn.Image.Factor = 1.0
    }

    // set state
    icn.state = newstate
    // re-blend icon
    icn.Blend()
}


// the fade image allows us to quickly change the overall opacity of our
// icon's original image, losslessly
// It playes back colors multiplied by a certain fade Factor
type fadeImage struct {
    Image  image.Image
    Factor float64
}

func (fade *fadeImage) ColorModel() color.Model {
    if fade.Factor == 1.0 {
        return fade.Image.ColorModel()
    }
    return color.RGBAModel
}

func (fade *fadeImage) Bounds() image.Rectangle {
    return fade.Image.Bounds()
}

func (fade *fadeImage) At(x, y int) color.Color {
    // do nothing if we're a normal image
    if fade.Factor == 1.0 {
        return fade.Image.At(x, y)
    }

    // get RGBA color
    clr := color.RGBAModel.Convert(fade.Image.At(x,y))
    rgba := clr.(color.RGBA)

    // multiply alpha
    rgba.A = uint8(float64(rgba.A) * fade.Factor)
    //return fade.Image.ColorModel().Convert(rgba)
    return rgba
}
