# goutline

!!!ALPHA SOFTWARE!!

A cross-platform terminal outliner; using it on:
- Windows (mainly in PowerShell, but old conhost seems to also work)
- Linux (suckless st termninal and whatever Ubuntu's default terminal is; I think gnome-terminal)

Inspired by the amazing https://github.com/void-rs/void (I have a branch that adds very rough Windows support via porting to crossterm-rs here: https://github.com/pmf/void/tree/crossterm-for-windows-support, but lost interest in Rust for recreational programming). I also looked at Zig (which I would love to like), but after spending 4 hours trying to just get at the args array on Windows, I decided to use Go.

Development happens on branch "el_loco".

Best viewed with Pragmata Pro.

Powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).


## Screenshots

On Windows:
![Screenshot on Windows](/screenshots/windows1.png)

## Build
```
go get goutline
go build
```

## Run
```
./goutline # implicitly uses out.json in current directory when saving (warning: no auto saving right now; do s, then q)
./goutline my_file.json # use this file
```

## Key bindings
(I try to keep these up to date, but refer directly to the implementation if something seems to behave oddly.)

SUBJECT TO CHANGE!

| Key                  | Action                                |
|----------------------|---------------------------------------|
| enter                | Insert new sibling below current item |
| ctrl+p               | Add new item as child of current item |
| e                    | Enter edit mode (in edit mode, the bubbletea input widget conveniently offers pseudo-readline key bindings) |
| (in edit mode) esc   | Exit edit mode and discard changes |
| (in edit mode) enter | Confirm changes and leave edit mode |
| backspace, d         | Delete current item |
| right, l             | Expand current item |
| left, h              | Collapse current item |
| down, j              | Next item |
| up, k                | Previous item |
| tab                  | Demote item (is that even a word?) |
| shift-tab            | Promote item |}
| q                    | Leave (*without* saving currently!) |
| s                    | Save current file (out.json if nothing else has been specified) |

