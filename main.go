package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"

    "goutline/goutlinelib"

    //org "github.com/niklasfasching/go-org/org"
    //"golang.org/x/sys/windows"
)


func main() {
    var filename string

    if len(os.Args) > 1 {
        filename = os.Args[1]
    } else {
        filename = "out.json"
    }

    m, err := goutlinelib.ModelFromFile(filename)

    if err != nil {
        fmt.Printf("Could not load file: %s; using default contents\n\n", filename)
        m = goutlinelib.InitialModel()
        m.SetFilename(filename)
    }

    p := tea.NewProgram(m)
    
    if err := p.Start(); err != nil {
        fmt.Printf("There has been an error: %v", err)
        os.Exit(1)
    }
}

/*
func main() {
    in, _ := windows.Open("CONIN$", windows.O_RDWR, 0)
    windows.SetConsoleMode(in, windows.ENABLE_WINDOW_INPUT)
    buf := make([]uint16, 1024)
    var iC byte = 0 
    var to_read uint32
    to_read = 1
    var read uint32
    windows.ReadConsole(in, &buf[0], to_read, &read, &iC)

    fmt.Printf("read: 0x%x\n", buf[0])
}
*/
