package progress

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/vbauerster/mpb/v5"
    "github.com/vbauerster/mpb/v5/decor"
    "io"
    "sync"
)

type Bar struct {
    barName          string
    progress         *Progress
    uiBar            *mpb.Bar
    unitPlural       string
    total            int
    runningTotal     int
    completedMessage string
    mutex            *sync.Mutex
    log              *logrus.Entry

    // FIXME Not the place for this
    SecretTracker structures.Set
}

func newBar(progress *Progress, barName string, total int, completedMessage string, log *logrus.Entry) (result *Bar) {
    result = &Bar{
        barName:          barName,
        progress:         progress,
        total:            total,
        runningTotal:     total,
        completedMessage: completedMessage,
        mutex:            &sync.Mutex{},
        log:              log,
        SecretTracker:    structures.NewSet(nil),
    }
    return
}

func (b *Bar) Start() {
    b.mutex.Lock()
    defer b.mutex.Unlock()

    if b.uiBar != nil {
        return
    }

    b.uiBar = b.progress.uiProgress.AddBar(int64(b.total),
        mpb.BarNoPop(),
        mpb.BarRemoveOnComplete(),
        mpb.PrependDecorators(
            decor.Name(b.barName, decor.WC{W: 50, C: decor.DidentRight}),
        ),
        mpb.AppendDecorators(
            decor.CountersNoUnit("searched %d of %d commits"),
        ),
    )
}

func (b *Bar) RunningTotal() int {
    b.mutex.Lock()
    defer b.mutex.Unlock()

    return b.runningTotal
}

func (b *Bar) Incr() {
    b.mutex.Lock()
    defer b.mutex.Unlock()

    b.uiBar.Increment()

    b.runningTotal -= 1

    if b.runningTotal == 0 {
        var message string

        // FIXME Not the place for this
        secretsFound := b.SecretTracker.Len()
        if secretsFound > 0 {
            message = fmt.Sprintf("%d unique secrets found", secretsFound)
        }

        b.Finished(message)
    }
}

func (b *Bar) Finished(perensMessage string) {
    b.progress.Add(0, mpb.BarFillerFunc(func(writer io.Writer, width int, st *decor.Statistics) {
        message := fmt.Sprintf(b.completedMessage, b.barName)
        fmt.Fprintf(writer, "- %s", message)
        if perensMessage != "" {
            fmt.Fprintf(writer, " (%s)", perensMessage)
        }
    })).SetTotal(0, true)
}

func (b *Bar) BustThrough(fnc func()) {
    b.progress.BustThrough(fnc)
}

// FIXME Not the place for this
func (b *Bar) AddSecret(secretIdent string) {
    b.SecretTracker.Add(secretIdent)
}
