# goutline

!!!ALPHA SOFTWARE!!

A cross-platform terminal outliner

Inspired by the amazing https://github.com/void-rs/void

Best viewed with Pragmata Pro

Powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)

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

