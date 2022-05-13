package goutlinelib

import(
    "errors"
    "fmt"
    "time"
)

type OItem interface {
    Init()
    GetId() string
    GetCreated() int64
    SetTimestampCreatedNow()
    GetChanged() int64
    SetTimestampChangedNow()
    GetTxt() string
    SetTxt(txt string)
    IsChecked() bool
    SetChecked(checked bool)

    GetSubs() []OItem
    SetSubs(subs []OItem)

    GetMeta() OItem

    GetParent() OItem
    SetParent(item OItem)

    // TODO: this is more of a view method
    IsExpanded() bool
    SetExpanded(expanded bool)

    IsEdited() bool
    SetEdited(edited bool)

    Delete(item OItem)
    AddSubAfterThis(item OItem)
    AddSubAt(item OItem, pos int)
    IsFirstSibling() bool
    IsLastSibling() bool
    HasSubs() bool
    Level(upTo OItem) int
    HasSub(sub OItem) bool
    IndexOfItem() int

    DeepCopy() OItem
    DeepCopyForUndo() OItem
}

type oitem struct {
    // interface serialization hack
    Type string

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
    Subs []OItem

    // generic meta information
    Meta OItem

    parent OItem
    
    edited bool
}

type oitemProxy struct {
    Type string

    target OItem
}

func IsOitemTypeEntry(temp map[string]interface{}) bool {
    typeSwitch := temp["Type"]
    
    if  typeSwitch == nil || typeSwitch.(string) == "oitem" || typeSwitch.(string) == "" {
        return true
    }

    return false
}

func UnmarshalJSONOItem(temp map[string]interface{}) (item OItem, err error) {
    err = nil

    if nil == temp {
        item = nil
        return
    }

    if  IsOitemTypeEntry(temp) {
        tmpItem := &oitem{}
        err = tmpItem.UnmarshalOitemStruct(temp)
        if nil == err {
            item = tmpItem
        } else {
            err = fmt.Errorf("Error when unmarshalling item: %w", err)
            return
        }
    } else {
        err = errors.New("Unsupported type")
        return
    }

    return
}

func (o *oitem) UnmarshalOitemStruct(temp map[string]interface{}) (err error) {
    err = nil

    if IsOitemTypeEntry(temp) {
        o.Type = "oitem"
    } else {
        err = errors.New("Unsupported type")
        return
    }

    o.Id = temp["Id"].(string)
    o.Created = int64(temp["Created"].(float64))
    o.Changed = int64(temp["Changed"].(float64))
    o.Txt = temp["Txt"].(string)
    o.Numbered = temp["Numbered"].(bool)
    o.Checked = temp["Checked"].(bool)
    o.Expanded = temp["Expanded"].(bool)

    tempMeta := temp["Meta"]
    if nil != tempMeta {
        var metaTmp OItem
        metaTmp, err = UnmarshalJSONOItem(tempMeta.(map[string]interface{}))
        
        if err != nil {
            err = fmt.Errorf("Could not UnmarshalJSONOItem: %w", err)
            return
        }

        o.Meta = metaTmp
    }

    tempSubs := temp["Subs"]
    if nil != tempSubs {
        arr := tempSubs.([]interface{})
        o.Subs = make([]OItem, 0, len(arr))
        for _, cur := range arr {
            tempCur := cur.(map[string]interface{})
            var new_item OItem
            new_item, err = UnmarshalJSONOItem(tempCur)

            if err != nil {
                err = fmt.Errorf("Could not create sub: %w", err)
                return
            }

            o.Subs = append(o.Subs, new_item)
        }
    }    

    return
}

func (o *oitem) Init() {
    o.Type = "oitem"

    for _, sub := range o.Subs {
        sub.SetParent(o);
        sub.Init();
    }
}

func (o *oitem) GetId() string {
    return o.Id
}

func (o *oitem) GetCreated() int64 {
    return o.Created
}

func (o *oitem) GetChanged() int64 {
    return o.Changed
}

func (o *oitem) GetTxt() string {
    return o.Txt
}

func (o *oitem) SetTxt(txt string) {
    o.Txt = txt
}

func (o *oitem) IsChecked() bool {
    return o.Checked
}

func (o *oitem) SetChecked(checked bool) {
    o.Checked = checked
}

func (o *oitem) GetSubs() []OItem {
    return o.Subs
}

func (o *oitem) SetSubs(subs []OItem) {
    o.Subs = subs
}

func (o *oitem) GetMeta() OItem {
    return o.Meta
}

func (o *oitem) GetParent() OItem {
    return o.parent
}

func (o *oitem) SetParent(item OItem) {
    o.parent = item
}

func (o *oitem) IsExpanded() bool {
    return o.Expanded
}

func (o *oitem) SetExpanded(expanded bool) {
    o.Expanded = expanded
}

func (o *oitem) IsEdited() bool {
    return o.edited
}

func (o *oitem) SetEdited(edited bool) {
    o.edited = edited
}

func (o *oitem) SetTimestampCreatedNow() {
    o.Created = time.Now().UTC().Unix()
    //o.Created = time.Now().UTC().Format(time.RFC3339)
}

func (o *oitem) SetTimestampChangedNow() {
    o.Changed = time.Now().UTC().Unix()
    //o.Created = time.Now().UTC().Format(time.RFC3339)
}

func (o *oitem) DeepCopy() OItem {
    result := &oitem{Txt: o.Txt}
    result.SetTimestampCreatedNow()

    result.Numbered = o.Numbered
    result.Checked = o.Checked

    if nil != o.Meta {
        result.Meta = o.Meta.DeepCopy()
    }

    for _, sub := range o.Subs {
        new_sub := sub.DeepCopy()
        new_sub.SetParent(result)
        result.Subs = append(result.Subs, new_sub)
    }

    return result
}

func (o *oitem) DeepCopyForUndo() OItem {
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
        new_sub.SetParent(result)
        result.Subs = append(result.Subs, new_sub)
    }

    return result
}

func (o *oitem) AddSubAfterThis(item OItem) {
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

func (o *oitem) AddSubAt(item OItem, pos int) {
    if pos < 0 {
        return
    }

    if pos < len(o.Subs) {
        o.Subs = append(o.Subs, &oitem{})
        copy(o.Subs[pos + 1:], o.Subs[pos:])
        o.Subs[pos] = item
        item.SetParent(o)
        o.SetTimestampChangedNow()
    } else if pos == len(o.Subs) {
        // special case: insert position is one after the existing items
        o.Subs = append(o.Subs, item)
        item.SetParent(o)
    } 
}

func (o *oitem) Delete(item OItem) {
    if item.GetParent() != o {
        return
    }

    idx := item.IndexOfItem()

    if -1 == idx {
        return
    }

    item.SetParent(nil)
    o.Subs = append(o.Subs[:idx], o.Subs[idx + 1 :]...)

    o.SetTimestampChangedNow()
}

func (o *oitem) IsFirstSibling() bool {
    if (nil != o.parent) && (o.parent.GetSubs()[0] == o) {
        return true
    }

    return false
}

func (o *oitem) IsLastSibling() bool {
    if (nil != o.parent) && (o.parent.GetSubs()[len(o.parent.GetSubs()) - 1] == o) {
        return true
    }

    return false
}

func (o *oitem) HasSubs() bool {
    return 0 != len(o.Subs)
}

func (o *oitem) Level(upTo OItem) int {
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

func (o *oitem) HasSub(sub OItem) bool {
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
        for i, cur := range o.parent.GetSubs() {
            if cur == o {
                result = i
                break
            }
        }
    }

    return result
}
