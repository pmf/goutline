package goutlinelib

import(
    "time"
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

func (o *oitem) DeepCopyForUndo() *oitem {
    result := &oitem{Txt: o.Txt}

    result.Numbered = o.Numbered
    result.Checked = o.Checked
    result.Expanded = o.Expanded
    result.Created = o.Created
    result.Changed = o.Changed
    result.Id = o.Id

    if nil != o.Meta {
        result.Meta = o.Meta.DeepCopyForUndo()
    }

    for _, sub := range o.Subs {
        new_sub := sub.DeepCopyForUndo()
        new_sub.parent = result
        result.Subs = append(result.Subs, new_sub)
    }

    return result
}

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

    o.parent.AddSubAt(item, pos + 1)
}

func (o *oitem) AddSubAt(item *oitem, pos int) {
    if pos < 0 {
        return
    }

    if pos < len(o.Subs) {
        o.Subs = append(o.Subs, &oitem{})
        copy(o.Subs[pos + 1:], o.Subs[pos:])
        o.Subs[pos] = item
        item.parent = o
        o.SetTimestampChangedNow()
    } else if pos == len(o.Subs) {
        // special case: insert position is one after the existing items
        o.Subs = append(o.Subs, item)
        item.parent = o
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

func (o *oitem) HasSub(sub *oitem) bool {
    result := false

    for _, item := range o.Subs {
        if item == sub {
            return true
        }

        result = item.HasSub(sub)

        if result {
            return result
        }
    }

    return false
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
