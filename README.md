# HTTPShell

A light-weight HTTP server shell.

## Usage

```go
import (
    "net/http"

    "github.com/deadblue/httpshell"
)

type App struct{}

func (a *App) BeforeStartup() (err error) {
    // TODO: Write your initialization code here.
    return nil
}

func (a *App) AfterShutdown() {
    // TODO: Write your finalization code here.
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // TODO: Handle HTTP request
}

func main() {
    shell := httpshell.New("tcp", ":8080", &App{})
    shell.Run()
}
```