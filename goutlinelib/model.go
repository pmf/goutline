package goutlinelib

import(
    "fmt"
    "io/ioutil"
    "encoding/json"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
)

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

    undoList []*oitem
    undoIndex int
    currentStateReachedViaUndoList bool

    newestItem *oitem
}

type Visitor interface {
    VisitTitle(m *model, item *oitem) error
    VisitConfig(m *model, item *oitem) error
    VisitItem(m *model, item *oitem, level int) error
}

func (m *model) CommonPostInit() {
    m.undoIndex = -1
    
    m.Title.Expanded = true


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

func (m *model) SetFilename(filename string) {
    m.filename = filename
}

func (m *model) PushUndo() {
    undo_entry := m.Title.DeepCopyForUndo()

    if m.undoIndex == len(m.undoList) - 1 {
        m.undoList = append(m.undoList, undo_entry)
        m.undoIndex++
    } else {
        // clear redo entries
        if m.undoIndex >= 0 {
            m.undoList = m.undoList[:m.undoIndex]
        } else {
            m.undoList = nil
        }

        m.undoList = append(m.undoList, undo_entry)
        m.undoIndex++
    }
}

func (m *model) PopUndo() {
    if len(m.undoList) == 0 {
        return
    }

    if m.undoIndex == -1 {
        return
    }

    // If we're not at the head of the list, we need to push
    // the current state so that redo will work
    if m.undoIndex == len(m.undoList) - 1 && !m.currentStateReachedViaUndoList{
        undo_entry := m.Title.DeepCopyForUndo()
        m.undoList = append(m.undoList, undo_entry)
    }

    m.Title = m.undoList[m.undoIndex]
    m.undoIndex--

    m.currentStateReachedViaUndoList = true
    m.UpdateLinearizedMapping()
}

func (m *model) Redo() {
    redo_index := m.undoIndex + 1

    if redo_index > len(m.undoList) - 1 {
        return
    }

    m.Title = m.undoList[redo_index]

    m.undoIndex = redo_index

    m.currentStateReachedViaUndoList = true
    m.UpdateLinearizedMapping()
}

func (m *model) SetTitle(title string) {
    m.Title.Txt = title
    m.Title.SetTimestampChangedNow()
}

func (m model) Init() tea.Cmd {
    return nil
}

// TODO: corresponding func m.VisitLinearized()? this could also be done with
//       a filter for visit that checks for item.Expanded

func (m *model) VisitAll(visitor Visitor) error {
    err := visitor.VisitTitle(m, m.Title)

    if nil != err {
        return err
    }

    err = visitor.VisitConfig(m, m.Config)

    if nil != err {
        return err
    }

    err = m.visitItemInternal(visitor, m.Title)

    if nil != err {
        return err
    }

    return nil
}

func (m *model) VisitLinearized(visitor Visitor) error {
    err := visitor.VisitTitle(m, m.Title)

    if nil != err {
        return err
    }

    err = visitor.VisitConfig(m, m.Config)

    if nil != err {
        return err
    }

    err = m.visitItemInternalOnlyExpanded(visitor, m.Title)

    if nil != err {
        return err
    }

    return nil
}

func (m *model) visitItemInternal(visitor Visitor, item *oitem) error {
    err := visitor.VisitItem(m, item, item.Level(nil))

    if nil != err {
        return err
    }

    for _, sub := range item.Subs {
        err = m.visitItemInternal(visitor, sub)

        if nil != err {
            return err
        }
    }

    return nil
}

func (m *model) visitItemInternalOnlyExpanded(visitor Visitor, item *oitem) error {
    err := visitor.VisitItem(m, item, item.Level(nil))

    if nil != err {
        return err
    }

    if item.Expanded {
        for _, sub := range item.Subs {
            err = m.visitItemInternalOnlyExpanded(visitor, sub)

            if nil != err {
                return err
            }
        }
    }

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
    was_expanded := item.Expanded
    item.Expanded = false
    m.UpdateLinearizedMapping()

    if was_expanded {
        m.Cursor = m.PosInLinearized(item)
    } else {
        if nil != item.parent {
            m.Cursor = m.PosInLinearized(item.parent)
        }
    }
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

    canUpdateViewport := true

    if m.editingItem {
        switch msg := msg.(type) {

        case tea.WindowSizeMsg:
            cmds = m.handleWinSizeChange(msg)

        case tea.KeyMsg:

            switch msg.String() {

            case "ctrl+c", "esc":
                cur.edited = false
                m.editingItem = false

                if nil != m.newestItem {
                    m.DeleteItem(m.newestItem)
                    m.newestItem = nil
                }

            case "enter":
                cur.edited = false
                m.editingItem = false

                new_text := m.textinput.Value()

                if "" == new_text && cur == m.newestItem {
                    m.DeleteItem(m.newestItem)
                    m.newestItem = nil
                } else if cur.Txt != new_text {
                    m.PushUndo()
                    cur.Txt = new_text
                    cur.SetTimestampChangedNow()
                }
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
                m.textinput.CursorEnd()

            case "c":
                m.copiedItem = cur.DeepCopy()

            case "x":
                // alternative operation would be:
                // m.copiedItem =m.DeepCopy(cur)
                // m.DeleteItem(cur)

                m.PushUndo()
                m.copiedItem = m.DeleteItem(cur)

            case "v":
                if nil != m.copiedItem {
                    m.PushUndo()
                    m.AddSubAfterThis(cur, m.copiedItem)
                    m.copiedItem = nil
                }

            case "delete", "d", "backspace":
                m.PushUndo()
                m.DeleteItem(cur)

            case "tab":
                m.PushUndo()
                m.Promote(cur)

            case "shift+tab":
                m.PushUndo()
                m.Demote(cur)

            case "ctrl+p":
                m.newestItem = m.AddNewItemAndEdit(cur)
                // not pushing onto undo stack; happens either on confirm, or we don't care about the item

            case "enter", "o":
                m.newestItem = m.AddNewItemAfterCurrentAndEdit(cur)
                // not pushing onto undo stack; happens either on confirm, or we don't care about the item

            case "u":
                m.PopUndo()

            case "ctrl+r":
                m.Redo()

            case "ctrl+c", "q":
                return m, tea.Quit

            case "up", "k":
                if m.Cursor > 0 {
                    m.Cursor--
                }

            case "ctrl+k":
                m.PushUndo()
                m.MoveUp(cur)

            case "down", "j":
                if m.Cursor < len(m.linearized) - 1 {
                    m.Cursor++
                }

            case "ctrl+j":
                m.PushUndo()
                m.MoveDown(cur)

            case " ":
                m.PushUndo()
                cur.Checked = !cur.Checked
                canUpdateViewport = false

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

            case "s":
                m.SaveCurrentAs(m.filename)

            /*
            case "O":
                orgConfig := org.New()
                orgDoc := orgConfig.Parse(strings.NewReader(""), "out.org")
                fmt.Printf("O %s\n", orgDoc)
            */
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

    switch msg.(type) {

        case tea.KeyMsg:
            // dispatching key messages while the prompt is open scrolls
            // the viewport on j/k in addition to j/k being added in
            // the input field ...
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

func drawItem(m *model, i int, item *oitem, open_elements map[int]bool) string {
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

    selected_style = lipgloss.NewStyle()
    
    if m.Cursor == i {
        selected_style = selected_style.Background(color_purple).Foreground(lipgloss.Color("255"))
    }

    if item.Checked {
        selected_style = selected_style.Strikethrough(true)
    }

    cur := m.linearized[m.Cursor]
    parent_of_cur := item.HasSub(cur)
    
    if parent_of_cur {
        selected_style = selected_style.Bold(true)
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

    header_text := fmt.Sprintf("%s [%s]", m.Title.Txt, m.filename)
    s := header_style.Render(header_text) + "\n\n"

    //s += level_headers

    for i, item := range m.linearized {
        s += drawItem(&m, i, item, open_elements)

        if !item.IsLastSibling() {
            open_elements[item.Level(nil)] = true
        } else {
            delete(open_elements, item.Level(nil))
        }
    }

    //s += level_headers

    copiedItemTxt := "-"

    if nil != m.copiedItem {
        copiedItemTxt = m.copiedItem.Txt
    }

    if m.Cursor < len(m.linearized) {
        s += fmt.Sprintf(
            "\n" + footer_style.Render("Press q to quit.      cursor: %d  undoIndex: %d len(undoList): %d reachedViaUndo: %t copied: %s") + "\n",
            m.Cursor,
            m.undoIndex,
            len(m.undoList),
            m.currentStateReachedViaUndoList,
            copiedItemTxt)
    }

    visualizeUndoList := true

    if visualizeUndoList {
        for idx, item := range m.undoList {
            prefix := "  "
            if idx == m.undoIndex {
                prefix = "->"
            }

            s += fmt.Sprintf("%s %s\n", prefix, item.Subs[0].Txt)
        }
    }

    return s
}

func (m model) View() string {
    if !m.winSizeReady {
        return "\n  Initializing..."
    }

    return fmt.Sprintf("%s", m.viewport.View())
}
