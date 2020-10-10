package inverted_index

import (
	"testing"

	"stree-index/pkg/inverted-index/labels"
	"stree-index/pkg/inverted-index/index"
	"stree-index/pkg/testutil"
)

func TestQuery(t *testing.T) {

	h := NewHeadReader()
	matchers1 := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "a", "2"),
	}
	matchers2 := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchNotEqual, "b", "1"),
	}
	matchers3 := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchRegexp, "a", "1|2"),
	}

	matchers4 := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchRegexp, "c", "[0-9]+"),
	}

	matchers5 := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchNotRegexp, "c", "[0-9]+"),
	}

	l1 := labels.Label{"a", "1",}
	l2 := labels.Label{"a", "2",}
	l3 := labels.Label{"b", "1",}
	l4 := labels.Label{"b", "2",}
	l5 := labels.Label{"c", "1",}
	l6 := labels.Label{"c", "2",}

	lset1 := labels.Labels{
		l1, l3,
	}
	lset2 := labels.Labels{
		l1, l4,
	}

	lset3 := labels.Labels{
		l2, l3,
	}
	lset4 := labels.Labels{
		l2, l4,
	}
	lset5 := labels.Labels{
		l1, l5,
	}
	lset6 := labels.Labels{
		l2, l6,
	}

	h.GetOrCreateWithID(1, lset1)
	h.GetOrCreateWithID(1, lset1)
	h.GetOrCreateWithID(2, lset2)
	h.GetOrCreateWithID(3, lset3)
	h.GetOrCreateWithID(4, lset4)
	h.GetOrCreateWithID(5, lset5)
	h.GetOrCreateWithID(6, lset6)

	// equal
	p1, e := PostingsForMatchers(h, matchers1...)
	testutil.Ok(t, e)
	p1s := make([]uint64, 0)
	for p1.Next() {
		p1s = append(p1s, p1.At())
	}
	tt1 := []uint64{3, 4, 6}
	testutil.Equals(t, p1s, tt1)

	// not_equal
	p2, e := PostingsForMatchers(h, matchers2...)
	tt2 := []uint64{2, 4, 5, 6}
	p2s := make([]uint64, 0)
	testutil.Ok(t, e)
	for p2.Next() {

		p2s = append(p2s, p2.At())
	}
	testutil.Equals(t, p2s, tt2)

	// regex
	p3, e := PostingsForMatchers(h, matchers3...)
	testutil.Ok(t, e)
	p3s := make([]uint64, 0)
	for p3.Next() {
		p3s = append(p3s, p3.At())
	}
	tt3 := []uint64{1, 2, 3, 4, 5, 6}
	testutil.Equals(t, p3s, tt3)

	p4, e := PostingsForMatchers(h, matchers4...)
	testutil.Ok(t, e)
	p4s := make([]uint64, 0)
	for p4.Next() {
		p4s = append(p4s, p4.At())
	}
	tt4 := []uint64{5, 6}
	testutil.Equals(t, p4s, tt4)

	// not regex
	p5, e := PostingsForMatchers(h, matchers5...)
	testutil.Ok(t, e)
	p5s, e := index.ExpandPostings(p5)
	testutil.Ok(t, e)

	tt5 := []uint64{1, 2, 3, 4}
	testutil.Equals(t, p5s, tt5)

}

func TestAdd(t *testing.T) {

}
