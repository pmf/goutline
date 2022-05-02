package main

import (
    "fmt"
    "os"
    "time"
    "io/ioutil"
    "encoding/json"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"

    //"golang.org/x/sys/windows"
)

type oitem struct {
    // unique id
    Id string

    // time created
    Created int64

    // time last changed
    Changed int64

    // main text
    Txt string

    // use auto-numbering for subs
    Numbered bool

    // Checked state
    // TODO: as label?
    Checked bool

    // expansion state
    Expanded bool

    // child items
    Subs []*oitem

    // generic meta information
    Meta *oitem

    // TODO: maybe use id instead?
    parent *oitem `json:"-"`

    edited bool
}

func (o *oitem) Init() {
    for _, sub := range o.Subs {
        sub.parent = o;
        sub.Init();
    }
}

func (o *oitem) SetTimestampCreatedNow() {
    o.Created = time.Now().UTC().Unix()
    //o.Created = time.Now().UTC().Format(time.RFC3339)
}

func (o *oitem) SetTimestampChangedNow() {
    o.Changed = time.Now().UTC().Unix()
    //o.Created = time.Now().UTC().Format(time.RFC3339)
}

func (o *oitem) DeepCopy() *oitem {
    result := &oitem{Txt: o.Txt}
    result.SetTimestampCreatedNow()

    result.Numbered = o.Numbered
    result.Checked = o.Checked

    if nil != o.Meta {
        result.Meta = o.Meta.DeepCopy()
    }

    for _, sub := range o.Subs {
        new_sub := sub.DeepCopy()
        new_sub.parent = result
        result.Subs = append(result.Subs, new_sub)
    }

    return result
}

// TODO: convenience methods AddSubBefore() and AddSubAfter()

func (o *oitem) AddSubAfterThis(item *oitem) {
    if nil == o.parent {
        return
    }

    if nil == item {
        return
    }

    pos := o.IndexOfItem()

    if -1 == pos {
        return
    }

    o.parent.AddSubAt(item, pos)
}

func (o *oitem) AddSubAt(item *oitem, pos int) {
    if (pos >= 0) && (pos < len(o.Subs)) {
        o.Subs = append(o.Subs, &oitem{})
        copy(o.Subs[pos + 1:], o.Subs[pos:])
        o.Subs[pos] = item
        item.parent = o
        o.SetTimestampChangedNow()
    }
}

func (o *oitem) Delete(item *oitem) {
    if item.parent != o {
        return
    }

    idx := item.IndexOfItem()

    if -1 == idx {
        return
    }

    item.parent = nil
    o.Subs = append(o.Subs[:idx], o.Subs[idx + 1 :]...)

    o.SetTimestampChangedNow()
}

func (o *oitem) IsFirstSibling() bool {
    if (nil != o.parent) && (o.parent.Subs[0] == o) {
        return true
    }

    return false
}

func (o *oitem) IsLastSibling() bool {
    if (nil != o.parent) && (o.parent.Subs[len(o.parent.Subs) - 1] == o) {
        return true
    }

    return false
}

func (o *oitem) HasSubs() bool {
    return 0 != len(o.Subs)
}

func (o *oitem) Level(upTo *oitem) int {
    level := 0
    var cur *oitem = o.parent

    if nil != upTo {
        level++
    }

    for cur != upTo {
        cur = cur.parent
        level++
    }

    return level
}

func (o *oitem) IndexOfItem() int {
    result := -1

    if nil != o.parent {
        for i, cur := range o.parent.Subs {
            if cur == o {
                result = i
                break
            }
        }
    }

    return result
}

const useHighPerformanceRenderer = false
const verticalMarginHeight = 0
const headerHeight = 0

type model struct {
    
    Title *oitem
    
    Config *oitem

    Cursor int
    
    linearized []*oitem
    linearCount int

    copiedItem *oitem

    filename string

    textinput textinput.Model
    editingItem bool

    viewport viewport.Model
    winSizeReady bool
}

func (m *model) CommonPostInit() {
    for _, item := range m.Title.Subs {
        item.parent = m.Title
        item.Init()
    }

    m.UpdateLinearizedMapping()

    ti := textinput.New()
    ti.Placeholder = "text"
    ti.Prompt = ""
    ti.Focus()

    m.textinput = ti
}

func InitialModel() model {
    /*
    result := model{
        Title: &oitem{
            Txt: "TODO",
            Subs: []*oitem{
                &oitem{Txt: "AAA"},
                &oitem{Txt: "BBB", Subs: []*oitem{&oitem{Txt: "bbb"}}},
                &oitem{Txt: "CCC"}}}}
    */

    result := model{
        filename: "out.json",
        Title: &oitem{
            Txt: "TODO",
            Subs: []*oitem{
                &oitem{Txt: "Item"}}},
        Config: &oitem{}}

    result.Title.SetTimestampCreatedNow();

    result.CommonPostInit()

    return result
}

func ModelFromFile(filename string) (model, error) {
    b, err := ioutil.ReadFile(filename)

    var result model

    if err != nil {
        fmt.Printf("Could not read file %s", filename)
        return result, err
    }

    err = json.Unmarshal(b, &result)

    if err != nil {
        fmt.Printf("Could not unmarshal contents from %s", filename)
        return result, err
    }

    result.CommonPostInit()

    result.filename = filename

    return result, err
}

func (m model) SetTitle(title string) {
    m.Title.Txt = title
    m.Title.SetTimestampChangedNow()
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m *model) UpdateLinearizedMappingInternal(item *oitem) {
    m.linearCount++
    m.linearized = append(m.linearized, item)

    if item.Expanded {
        for _, sub := range item.Subs {
            m.UpdateLinearizedMappingInternal(sub)
        }
    }
}

func (m *model) UpdateLinearizedMapping() {
    m.linearCount = 0
    m.linearized = nil

    for _, item := range m.Title.Subs {
        m.UpdateLinearizedMappingInternal(item)
    }
}

func (m *model) AddNewItem(parent *oitem) *oitem {
    new_item := &oitem{parent: parent, Txt: ""}
    new_item.SetTimestampCreatedNow()
    parent.Subs = append(parent.Subs, new_item)
    //new_item.Txt = fmt.Sprintf("new %s.%d", parent.Txt, len(parent.Subs) - 1)
    m.UpdateLinearizedMapping()

    parent.SetTimestampChangedNow()
    return new_item
}

func (m *model) AddNewItemAndEdit(parent *oitem) *oitem {
    new_item := m.AddNewItem(parent)
    m.Expand(parent)
    new_cur_pos := m.PosInLinearized(new_item)

    if -1 != new_cur_pos {
        m.Cursor = new_cur_pos
        new_item.edited = true
        m.editingItem = true
        m.textinput.SetValue(new_item.Txt)
    }

    return new_item
}

func (m *model) AddNewItemAfterCurrentAndEdit(item *oitem) *oitem {
    insert_pos := item.IndexOfItem() + 1

    if -1 != insert_pos {
        new_item := &oitem{parent: item.parent}
        new_item.SetTimestampCreatedNow()
        item.parent.Subs = append(item.parent.Subs, &oitem{})
        copy(item.parent.Subs[insert_pos + 1:], item.parent.Subs[insert_pos:])
        item.parent.Subs[insert_pos] = new_item

        m.UpdateLinearizedMapping()

        new_cur_pos := m.PosInLinearized(new_item)

        if -1 != new_cur_pos {
            m.Cursor = new_cur_pos
            new_item.edited = true
            m.editingItem = true
            m.textinput.SetValue(new_item.Txt)
        }

        item.SetTimestampChangedNow()

        m.UpdateLinearizedMapping()

        return new_item
    }

    return nil
}

func (m *model) PosInLinearized(item *oitem) int {
    result := -1

    for idx, cur := range m.linearized {
        if cur == item {
            result = idx
            break
        }
    }

    return result
}

func (m *model) Expand(item *oitem) {
    item.Expanded = true
    m.UpdateLinearizedMapping()
}

func (m *model) Collapse(item *oitem) {
    item.Expanded = false
    m.UpdateLinearizedMapping()
}

func (m *model) DeleteItem(item *oitem) *oitem{
    if nil == item {
        return nil
    }

    if nil == item.parent {
        return nil
    }

    // save
    p := item.parent

    // make sure we're not deleting the last item
    if (item.Level(nil) == 1) && (len(item.parent.Subs) == 1) {
        item.Txt = "empty"

        for _, sub := range item.Subs {
            sub.parent = nil
        }

        item.Subs = nil
        item.SetTimestampChangedNow()
    } else {
        item.parent.Delete(item)
    }

    p.SetTimestampChangedNow()
    m.UpdateLinearizedMapping()

    return item
}

func (m *model) AddSubAfterThis(o *oitem, item *oitem) {
    if nil == o {
        return
    }

    if nil == item {
        return
    }

    o.AddSubAfterThis(item)

    m.UpdateLinearizedMapping()

    if pos := m.PosInLinearized(item); -1 != pos {
        m.Cursor = pos
    }
}

func (m *model) MoveUp(item *oitem) {
    idx := item.IndexOfItem()

    if -1 == idx {
        return
    }

    if 0 != idx {
        tmp := item.parent.Subs[idx - 1]
        item.parent.Subs[idx - 1] = item
        item.parent.Subs[idx    ] = tmp
        // parent references can stay the same
    }

    m.UpdateLinearizedMapping()

    if pos := m.PosInLinearized(item); -1 != pos {
        m.Cursor = pos
    }
}

func (m *model) MoveDown(item *oitem) {
    idx := item.IndexOfItem()

    if -1 == idx {
        return
    }

    if (len(item.parent.Subs) - 1) != idx {
        tmp := item.parent.Subs[idx + 1]
        item.parent.Subs[idx + 1] = item
        item.parent.Subs[idx    ] = tmp
        // parent references can stay the same
    }

    m.UpdateLinearizedMapping()

    if pos := m.PosInLinearized(item); -1 != pos {
        m.Cursor = pos
    }
}

func (m *model) Promote(item *oitem) {
    // Increases the level of the given item
    //
    // This means moving the current item
    // to be the last child of the preceding current sibling
    // first try to find the preceding current sibling
    //
    // From:
    //
    //    ├─ · Item
    // >  ├─ · new TODO.1 <
    //    └─ · new TODO.2
    //
    // To:
    //
    //    ├┐ ⊝ Item
    // >  │└─ · new TODO.1 <
    //    └─ · new TODO.2
    //
    if nil != item.parent {
        var prec *oitem = nil
        index_of_item_within_parent := item.IndexOfItem()

        for _, cur := range item.parent.Subs {
            if cur == item {
                break
            }

            prec = cur
        }

        if (nil != prec) && (-1 != index_of_item_within_parent) {
            // remove item from parent's subs
            // trick from https://github.com/golang/go/wiki/SliceTricks
            item.parent.Subs = append(item.parent.Subs[:index_of_item_within_parent], item.parent.Subs[index_of_item_within_parent + 1 :]...)

            // attach to new parent
            item.parent.SetTimestampChangedNow()
            item.parent = prec
            item.parent.SetTimestampChangedNow()
            prec.Subs = append(prec.Subs, item)
            item.SetTimestampChangedNow()

            m.Expand(item.parent)
            m.UpdateLinearizedMapping()
        }
    }
}

func (m *model) Demote(item *oitem) {
    // Decreases the level of the given item
    //
    // This means removing the item from its current parent,
    // and inserting it as a sibling of its current parent
    // in current parent's parent, after it current parent
    //
    // From:
    //
    //    ├┐ ⊝ Item
    // >  │└─ · new TODO.1 <
    //    └─ · new TODO.2
    //
    // To:
    //
    //    ├─ · Item
    // >  ├─ · new TODO.1 <
    //    └─ · new TODO.2
    //
    if (nil != item.parent) && (nil != item.parent.parent) {
        index_of_item_within_parent := item.IndexOfItem()
        index_of_parent_within_its_parent := item.parent.IndexOfItem()

        if (-1 != index_of_parent_within_its_parent) && (-1 != index_of_item_within_parent) {
            // remove item from parent's subs
            // trick from https://github.com/golang/go/wiki/SliceTricks
            item.parent.Subs = append(item.parent.Subs[:index_of_item_within_parent], item.parent.Subs[index_of_item_within_parent + 1 :]...)

            // insert item into item.parent.parent at position index_of_parent_within_its_parent + 1
            if item.parent.IsLastSibling() {
               item.parent.parent.Subs = append(item.parent.parent.Subs, item)
            } else {
                target_index := index_of_parent_within_its_parent + 1
                item.parent.parent.Subs = append(item.parent.parent.Subs, &oitem{})
                copy(item.parent.parent.Subs[target_index + 1:], item.parent.parent.Subs[target_index:])
                item.parent.parent.Subs[target_index] = item
            }

            item.parent.SetTimestampChangedNow()
            item.parent = item.parent.parent
            item.parent.SetTimestampChangedNow()
            item.SetTimestampChangedNow()
            m.UpdateLinearizedMapping()
        }
    }
}

func (m model) SaveCurrentAs(filename string) bool {
    b, err := json.MarshalIndent(m, "", "    ")

    if err != nil {
        fmt.Printf("Error when marshalling struct for %s", filename)
        return false
    }

    err = ioutil.WriteFile(filename, b, 0644)

    if err != nil {
        fmt.Printf("Error when saving %s", filename)
        return false
    }

    return true
}

func (m *model) handleWinSizeChange(msg tea.WindowSizeMsg) []tea.Cmd {
    var cmds []tea.Cmd

    if !m.winSizeReady {
        m.viewport = viewport.New(msg.Width, msg.Height - verticalMarginHeight)
        m.viewport.YPosition = headerHeight
        m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
        m.viewport.SetContent(m.contentView())
        m.winSizeReady = true

        m.viewport.YPosition = headerHeight + 1
    } else {
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height - verticalMarginHeight
    }

    if useHighPerformanceRenderer {
        cmds = append(cmds, viewport.Sync(m.viewport))
    }

    return cmds
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    var cmds []tea.Cmd

    cur := m.linearized[m.Cursor]

    if m.editingItem {
        switch msg := msg.(type) {

        case tea.WindowSizeMsg:
            cmds = m.handleWinSizeChange(msg)

        case tea.KeyMsg:

            switch msg.String() {

            case "ctrl+c", "esc":
                cur.edited = false
                m.editingItem = false

            case "enter":
                cur.edited = false
                m.editingItem = false
                cur.Txt = m.textinput.Value()

                cur.SetTimestampChangedNow()
            }
        }

        if m.editingItem {
            m.textinput, _ = m.textinput.Update(msg)
        }
    } else {
        switch msg := msg.(type) {

        case tea.WindowSizeMsg:
            cmds = m.handleWinSizeChange(msg)

        case tea.KeyMsg:

            switch msg.String() {

            case "i":
                cur.edited = true
                m.editingItem = true
                m.textinput.SetValue(cur.Txt)

            case "c":
                m.copiedItem = cur.DeepCopy()

            case "x":
                // alternative operation would be:
                // m.copiedItem =m.DeepCopy(cur)
                // m.DeleteItem(cur)

                m.copiedItem = m.DeleteItem(cur)

            case "v":
                if nil != m.copiedItem {
                    m.AddSubAfterThis(cur, m.copiedItem)
                    m.copiedItem = nil
                }

            case "ctrl+c", "q":
                return m, tea.Quit

            case "up", "k":
                if m.Cursor > 0 {
                    m.Cursor--
                }

            case "ctrl+k":
                m.MoveUp(cur)

            case "down", "j":
                if m.Cursor < len(m.linearized) - 1 {
                    m.Cursor++
                }

            case "ctrl+j":
                m.MoveDown(cur)

            case " ":
                cur.Checked = !cur.Checked

            case "tab":
                m.Promote(cur)

            case "shift+tab":
                m.Demote(cur)

            case "ctrl+p":
                m.AddNewItemAndEdit(cur)

            case "enter", "o":
                m.AddNewItemAfterCurrentAndEdit(cur)

            case "right", "l":
                m.Expand(cur)

            case "left", "h":
                if cur.Expanded {
                    m.Collapse(cur)
                } else {
                    // if already collapses, collapse parent
                    if nil != cur.parent {
                        m.Collapse(cur.parent)
                    }
                }

            case "delete", "d", "backspace":
                m.DeleteItem(cur)

            case "s":
                m.SaveCurrentAs(m.filename)
           }
        }
    }

    // make sure cursor is in valid range
    if m.Cursor < 0 {
        m.Cursor = 0
    }

    if m.Cursor >= m.linearCount {
        m.Cursor = m.linearCount - 1
    }

    m.viewport.SetContent(m.contentView())

    canUpdateViewport := true

    switch msg.(type) {

        case tea.KeyMsg:
            // only dispatch messages when
            if m.editingItem {
                canUpdateViewport = false
            }
    }

    if canUpdateViewport {
        m.viewport, cmd = m.viewport.Update(msg)
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}

func drawItem(m model, i int, item *oitem, open_elements map[int]bool) string {
    cursor_left := " "
    cursor_right := ""

    if m.Cursor == i {
        cursor_left = ">"
        cursor_right = " <"
    }

    checked := " "
    if item.Checked {
        checked = "✓"
    }

    level := item.Level(nil)

    branches := ""
    level_indicator := ""
    for i := 0; i < level; i++ {
        if i == (level - 1) {
            if item.IsLastSibling() {
                level_indicator += "└"
                branches += "1"
            } else {
                level_indicator += "├"
                branches += "2"
            }

            if item.HasSubs() && item.Expanded {
                if item.IsLastSibling() {
                    level_indicator += "┬" // "┐"
                    branches += "3a"
                } else {
                    level_indicator += "┐"
                    branches += "3b"
                }
            } else {
                level_indicator += "─"
                branches += "4"
            }
        } else {
            value, found := open_elements[i + 1]

            if found && value {
                level_indicator += "│"
                branches += "5"
            } else {
                level_indicator += " "
                //level_indicator += fmt.Sprintf("%d", i)
                branches += "6"
            }
        }
    }

    if len(item.Subs) > 0 {
        if item.Expanded {
            level_indicator += " ▽ " // " ⊝ "
        } else {
            level_indicator += " ▶ " // " ⊕ "
        }
    } else {
        level_indicator += " · " // " ▷ "
    }

    show_open_elements := false
    open_elements_indicator := ""

    if show_open_elements {
        open_elements_indicator += fmt.Sprintf(" (level: %d, oe: %v (len %d) branch: %s)", level, open_elements, len(open_elements), branches)
    }

    color_purple := lipgloss.Color("63")

    var selected_style lipgloss.Style

    if m.Cursor == i {
        //selected_style = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("227"));
        selected_style = lipgloss.NewStyle().Background(color_purple).Foreground(lipgloss.Color("255"));
    } else {
        selected_style = lipgloss.NewStyle()
    }

    if item.edited {
        return fmt.Sprintf("%s %s%s%s%s%s\n", cursor_left, checked, level_indicator, m.textinput.View(), open_elements_indicator, cursor_right)
    } else {
        return fmt.Sprintf("%s %s%s%s%s%s\n", cursor_left, checked, level_indicator, selected_style.Render(item.Txt), open_elements_indicator, cursor_right)
    }
}

func (m model) contentView() string {
    // keep track of which elements are open on each level (displayed part of subs, but more subs
    // will be painted after painting intermediate subs of higher levels)
    open_elements := make(map[int]bool)
    //level_headers := "   012345\n"

    color_yellow := lipgloss.Color("227")
    header_style := lipgloss.NewStyle().Background(color_yellow).Foreground(lipgloss.Color("0"));
    footer_style := lipgloss.NewStyle().Background(color_yellow).Foreground(lipgloss.Color("0"));

    s := header_style.Render(m.Title.Txt) + "\n\n"

    //s += level_headers

    for i, item := range m.linearized {
        s += drawItem(m, i, item, open_elements)

        if !item.IsLastSibling() {
            open_elements[item.Level(nil)] = true
        } else {
            delete(open_elements, item.Level(nil))
        }
    }

    //s += level_headers

    if m.Cursor < len(m.linearized) {
        s += fmt.Sprintf(
            "\n" + footer_style.Render("Press q to quit.      cursor: %d  IsLastSibling: %t IsFirstSibling: %t ") + "\n",
            m.Cursor,
            m.linearized[m.Cursor].IsLastSibling(),
            m.linearized[m.Cursor].IsFirstSibling())
    }

    return s
}

func (m model) View() string {
    if !m.winSizeReady {
        return "\n  Initializing..."
    }

    return fmt.Sprintf("%s", m.viewport.View())
}

func main() {
    var filename string

    if len(os.Args) > 1 {
        filename = os.Args[1]
    } else {
        filename = "out.json"
    }

    m, err := ModelFromFile(filename)

    if err != nil {
        fmt.Printf("Could not load file: %s; using default contents\n\n", filename)
        m = InitialModel()
        m.filename = filename
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
