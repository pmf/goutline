package goutlinelib

type oitemproxy struct {
    Type string

    // unique id
    Id string

    // expansion state
    Expanded bool

    target OItem

    parent OItem

    cachedProxiedSubs []OItem

    edited bool
}

func NewProxy(target OItem) OItem {
    result := &oitemproxy{target: target}
    return result
}

func (o *oitemproxy) Init() {
    o.Type = "oitemproxy"
}

func (o *oitemproxy) GetId() string {
    return o.Id
}

func (o *oitemproxy) GetCreated() int64 {
    return o.target.GetCreated()
}

func (o *oitemproxy) GetChanged() int64 {
    return o.target.GetChanged()
}

func (o *oitemproxy) GetTxt() string {
    return o.target.GetTxt()
}

func (o *oitemproxy) SetTxt(txt string) {
    o.target.SetTxt(txt)
}

func (o *oitemproxy) IsChecked() bool {
    return o.target.IsChecked()
}

func (o *oitemproxy) SetChecked(checked bool) {
    o.target.SetChecked(checked)
}

func (o *oitemproxy) GetSubs() []OItem {
    if len(o.cachedProxiedSubs) != len(o.target.GetSubs()) {
        o.cachedProxiedSubs = make([]OItem, 0, len(o.target.GetSubs()))

        for _, cur := range o.target.GetSubs() {
            o.cachedProxiedSubs = append(o.cachedProxiedSubs, &oitemproxy{target: cur, parent: o})
        }
    }

    return o.cachedProxiedSubs
}

func (o *oitemproxy) SetSubs(subs []OItem) {
    o.target.SetSubs(subs)
}

func (o *oitemproxy) GetMeta() OItem {
    return o.target.GetMeta()
}

func (o *oitemproxy) GetParent() OItem {
    return o.parent
}

func (o *oitemproxy) SetParent(item OItem) {
    o.parent = item
}

func (o *oitemproxy) IsExpanded() bool {
    return o.Expanded
}

func (o *oitemproxy) SetExpanded(expanded bool) {
    o.Expanded = expanded
}

func (o *oitemproxy) IsEdited() bool {
    return o.edited
}

func (o *oitemproxy) SetEdited(edited bool) {
    o.edited = edited
}

func (o *oitemproxy) SetTimestampCreatedNow() {
}

func (o *oitemproxy) SetTimestampChangedNow() {
}

func (o *oitemproxy) DeepCopy() OItem {
    return nil
}

func (o *oitemproxy) DeepCopyForUndo() OItem {
    return nil
}

func (o *oitemproxy) AddSubAfterThis(item OItem) {
}

func (o *oitemproxy) AddSubAt(item OItem, pos int) {
}

func (o *oitemproxy) Delete(item OItem) {
}

func (o *oitemproxy) IsFirstSibling() bool {
    if (nil != o.GetParent()) && (o.GetParent().GetSubs()[0] == o) {
        return true
    }

    return false
}

func (o *oitemproxy) IsLastSibling() bool {
    if (nil != o.GetParent()) && (o.GetParent().GetSubs()[len(o.GetParent().GetSubs()) - 1] == o) {
        return true
    }

    return false
}

func (o *oitemproxy) HasSubs() bool {
    return 0 != len(o.GetSubs())
}

func (o *oitemproxy) Level(upTo OItem) int {
    level := 0
    var cur OItem = o.parent

    if nil != upTo {
        level++
    }

    for cur != upTo {
        cur = cur.GetParent()
        level++
    }

    return level
}

func (o *oitemproxy) HasSub(sub OItem) bool {
    result := false

    for _, item := range o.GetSubs() {
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

func (o *oitemproxy) IndexOfItem() int {
    result := -1

    if nil != o.parent {
        for i, cur := range o.parent.GetSubs() {
            if cur == o {
                result = i
                break
            }
        }
    }

    return result
}
