package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "encoding/json"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/textinput"
)

type oitem struct {
    // unique id
    Id string

    // time created
    Created uint64

    // time last changed
    Changed uint64

    // main text
    Txt string

    // use auto-numbering for subs
    Numbered bool

    // labels
    // TODO: better as kv pairs
    Labels []string

    // selection state
    // TODO: as label
    Selected bool

    // expansion state
    Expanded bool

    // child items
    Subs []*oitem

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

type model struct {
    
    Title *oitem
    
    Config *oitem

    Cursor int
    
    linearized []*oitem
    
    linearCount int

    filename string

    textinput textinput.Model

    editingItem bool
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

func (m *model) AddItem(parent *oitem) {
    new_item := &oitem{parent: parent, Txt: ""}
    parent.Subs = append(parent.Subs, new_item)
    new_item.Txt = fmt.Sprintf("new %s.%d", parent.Txt, len(parent.Subs) - 1)
    m.UpdateLinearizedMapping()
}

func (m *model) Expand(item *oitem) {
    item.Expanded = true
    m.UpdateLinearizedMapping()
}

func (m *model) Collapse(item *oitem) {
    item.Expanded = false
    m.UpdateLinearizedMapping()
}

func (m *model) DeleteItem(item *oitem) {
    if nil == item.parent {
        return
    }

    // make sure we're not deleting the last item
    if (item.Level(nil) == 1) && (len(item.parent.Subs) == 1) {
        item.Txt = "empty"

        for _, sub := range item.Subs {
            sub.parent = nil
        }

        item.Subs = nil
    } else {
        item.parent.Delete(item)
    }

    m.UpdateLinearizedMapping()
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
            item.parent = prec
            prec.Subs = append(prec.Subs, item)

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

            item.parent = item.parent.parent
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    cur := m.linearized[m.Cursor]

    if m.editingItem {
        switch msg := msg.(type) {

        case tea.KeyMsg:

            switch msg.String() {

            case "ctrl+c", "esc":
                cur.edited = false
                m.editingItem = false

            case "enter":
                cur.edited = false
                m.editingItem = false
                cur.Txt = m.textinput.Value()
            }
        }

        if m.editingItem {
            m.textinput, _ = m.textinput.Update(msg)
        }

    } else {
        switch msg := msg.(type) {

        case tea.KeyMsg:

            switch msg.String() {

            case "e":
                cur.edited = true
                m.editingItem = true
                m.textinput.SetValue(cur.Txt)

            case "ctrl+c", "q":
                return m, tea.Quit

            case "up", "k":
                if m.Cursor > 0 {
                    m.Cursor--
                }

            case "down", "j":
                if m.Cursor < len(m.linearized) - 1 {
                    m.Cursor++
                }

            case " ":
                cur.Selected = !cur.Selected

            case "tab":
                m.Promote(cur)

            case "shift+tab":
                m.Demote(cur)

            case "ctrl+p":
                m.AddItem(cur)
                m.Expand(cur)

            case "enter", "o":
                m.AddItem(cur.parent)
                m.Expand(cur)

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

    return m, nil
}

func drawItem(m model, i int, item *oitem, open_elements map[int]bool) string {
    cursor_left := " "
    cursor_right := ""

    if m.Cursor == i {
        cursor_left = ">"
        cursor_right = " <"
    }

    checked := " "
    if item.Selected {
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

    //color_yellow := lipgloss.Color("227")
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

func (m model) View() string {
    // keep track of which elements are open on each level (displayed part of subs, but more subs
    // will be painted after painting intermediate subs of higher levels)
    open_elements := make(map[int]bool)
    //level_headers := "   012345\n"


    s := m.Title.Txt + "\n\n"
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
            "\nPress q to quit.      cursor: %d  IsLastSibling: %t IsFirstSibling: %t \n",
            m.Cursor,
            m.linearized[m.Cursor].IsLastSibling(),
            m.linearized[m.Cursor].IsFirstSibling())
    }

    return s
}

func main() {
    var m model

    if len(os.Args) > 1 {
        filename := os.Args[1]
        var err error
        m, err = ModelFromFile(filename)

        if err != nil {
            fmt.Printf("Could not load file: %s; using default contents\n\n", filename)
            m = InitialModel()
            m.filename = filename
        }
    } else {
        m = InitialModel()
    }
    
    p := tea.NewProgram(m)
    
    if err := p.Start(); err != nil {
        fmt.Printf("There has been an error: %v", err)
        os.Exit(1)
    }
}

