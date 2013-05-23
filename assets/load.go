/*
    Load and access assets, which right now are icons for the user interface
*/
package assets

import (
    "os"
    "log"
    imglib "image"
    _ "image/png"
    "bytes"

    // for getting function names
    "reflect"
    "runtime"
)

// asset declaration
var (

    logger       *log.Logger

    SplitTop     imglib.Image
    SplitRight   imglib.Image
    SplitBottom  imglib.Image
    SplitLeft    imglib.Image

    ShoveTop     imglib.Image
    ShoveRight   imglib.Image
    ShoveBottom  imglib.Image
    ShoveLeft    imglib.Image

    SwapCenter   imglib.Image
)



// Load functions are generated from binary data by the
// go-bindata tool
type Loader func() ([]byte)

func fatal(name string, err error) {
    if err != nil {
        log.Panicf("Asset load for %s failed: %v\n", name, err)
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


// load all assets
func init() {
    // set up logging first
    logger = log.New(os.Stderr, "[assets]", log.LstdFlags | log.Lshortfile)

    SplitTop =     image(split_top_png)
    SplitRight =   image(split_right_png)
    SplitBottom =  image(split_bottom_png)
    SplitLeft =    image(split_left_png)

    ShoveTop =     image(shove_top_png)
    ShoveRight =   image(shove_right_png)
    ShoveBottom =  image(shove_bottom_png)
    ShoveLeft =    image(shove_left_png)

    SwapCenter =   image(swap_center_png)
}



