package util
import "log"

// panic if an error occurs
// intended to be used in critial main executable code
func Fatal(err error) {
    if err != nil {
        log.Panic(err)
    }
}
