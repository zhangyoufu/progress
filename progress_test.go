package progress_test

import (
    "log"
    "time"

    "github.com/zhangyoufu/progress"
)

func Example() {
    log.SetOutput(progress.Writer)
    progress.Update("Status 1")
    log.Print("Log Output")
    time.Sleep(time.Second) // "Status 1" will be restored and shown below "Log Output"
    progress.Writer.WriteString("WriteString without LF")
    progress.Update("Status 2") // '\n' will be prepended
    time.Sleep(time.Second)
    progress.Update("Status 3") // "Status 3" will overwrite "Status 2"
    time.Sleep(time.Second)
    progress.Clear() // "Status 3" will be cleared
}
