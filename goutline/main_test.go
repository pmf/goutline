package main

import (
    "testing"
)

func TestLevelNoParent(t *testing.T) {
    o := oitem{Txt: "o"}
    
    res := o.Level(nil)
    if res != 0 {
        t.Error("Expected", 0, "for level of oitem, but was", res)
    }
}

func TestLevelOneParent(t *testing.T) {
    op := oitem{Txt: "op"}
    o  := oitem{Txt: "o", parent: &op}

    res := o.Level(nil)
    if res != 1 {
        t.Error("Expected", 1, "for level of oitem when checking up to nil, but was", res)
    }

    res = o.Level(&op)
    if res != 1 {
        t.Error("Expected", 1, "for level of oitem when checking up to item, but was", res)
    }
}

func TestLevelTwoParents(t *testing.T) {
    opp := oitem{Txt: "opp"}
    op  := oitem{Txt: "op", parent: &opp}
    o   := oitem{Txt: "o",  parent: &op}

    res := o.Level(nil)
    if res != 2 {
        t.Error("Expected", 2, "for level of oitem when checking up to nil, but was", res)
    }

    res = o.Level(&opp)
    if res != 2 {
        t.Error("Expected", 2, "for level of oitem when checking up to nil, but was", res)
    }

    res = o.Level(&op)
    if res != 1 {
        t.Error("Expected", 1, "for level of oitem when checking up to item, but was", res)
    }
}

func TestLinearizationSimple(t *testing.T) {
    m := InitialModel()

    if m.linearCount != 3 {
        t.Error("Expected", 3, "for m.linearCount, but got", m.linearCount)
    }

    if len(m.Title.Subs) != 3 {
        t.Error("Expected", 3, "for m.Title.Subs, but got", len(m.Title.Subs))
    }

    if len(m.linearized) != 3 {
        t.Error("Expected", 3, "for m.linearized, but got", len(m.linearized))
    }

    m.Expand(m.Title.Subs[1])

    if m.linearCount != 3 {
        t.Error("Expected", 3, "for m.linearCount, but got", m.linearCount)
    }

    if len(m.linearized) != 3 {
        t.Error("Expected", 3, "for m.linearized, but got", len(m.linearized))
    }
}

func TestSetTitle(t *testing.T) {
    m := InitialModel()

    m.SetTitle("foo")
    if m.Title.Txt != "foo" {
        t.Error("Expected", "foo", "for m.Title, but got", m.Title)
    }
}
