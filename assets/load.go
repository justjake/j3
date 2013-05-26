/*
    Load and access assets, which right now are icons for the user interface
*/
package assets

import (
    "os"
    "log"
    imglib "image"
    _ "image/png"   // so far the assets are only PNGs
    "bytes"

    // for getting function names
    "reflect"
    "runtime"
)

var (

    logger       *log.Logger = log.New(os.Stderr, "[assets]", log.LstdFlags | log.Lshortfile)

    // icons indicating splitting a window in half, and moving a new window into the freed space
    SplitTop     imglib.Image = image(split_top_png)
    SplitRight   imglib.Image = image(split_right_png)
    SplitBottom  imglib.Image = image(split_bottom_png)
    SplitLeft    imglib.Image = image(split_left_png)

    // icons indicating moving a containter in a direction, and inserting a window into the freed space
    ShoveTop     imglib.Image = image(shove_top_png)
    ShoveRight   imglib.Image = image(shove_right_png)
    ShoveBottom  imglib.Image = image(shove_bottom_png)
    ShoveLeft    imglib.Image = image(shove_left_png)

    // incon indicating swapping two containers
    Swap         imglib.Image = image(swap_center_png)

    // Named map of the assets
    Named =      map[string]imglib.Image{
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
)

// Load functions are generated from binary data by the
// go-bindata tool
type Loader func() ([]byte)

// assets are go code so we should be worried if they fail
func fatal(name string, err error) {
    if err != nil {
        log.Fatalf("Asset load for %s failed: %v\n", name, err)
    }
}

// get the name of a function
// see http://stackoverflow.com/questions/7052693/how-to-get-the-name-of-a-function-in-go
func getName(any interface{}) string {
    return runtime.FuncForPC(reflect.ValueOf(any).Pointer()).Name()
}


// load an image from an asset function
func image(load Loader) imglib.Image {
    name := getName(load)
    logger.Printf("loading asset %s\n", name)

    // get asset byte data
    data := load()

    // need a reader for decoding
    reader := bytes.NewReader(data)

    // decode
    img, _, err := imglib.Decode(reader)
    fatal(name, err)

    return img
}
