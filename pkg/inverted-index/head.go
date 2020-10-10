package inverted_index

import (
	"sync"

	"stree-index/pkg/inverted-index/index"
	"stree-index/pkg/inverted-index/labels"
)

type IndexReader interface {
	LabelValues(name string) ([]string, error)
	Postings(name string, values ...string) (index.Postings, error)
}

type stringset map[string]struct{}

func (ss stringset) set(s string) {
	ss[s] = struct{}{}
}

type HeadIndexReader struct {
	postings *index.MemPostings
	values   map[string]stringset
	symbols  map[string]struct{}
	symMtx   sync.RWMutex
}
type stat struct {
	Name  string `json:"name"`
	Value uint64 `json:"value"`
}

type indexStatus struct {
	LabelValueCount     []stat `json:"label_value_count"`
	LabelValuePairCount []stat `json:"label_value_pair_count"`
}

type labelGroup struct {
	Group []stat `json:"group"`
	Message string `json:"message"`
}

func NewLabelGroup() labelGroup {
	ss := make([]stat, 0)
	a := labelGroup{
		Group: ss,
	}
	return a
}

func NewHeadReader() *HeadIndexReader {

	h := &HeadIndexReader{
		values:   make(map[string]stringset),
		symbols:  make(map[string]struct{}),
		postings: index.NewUnorderedMemPostings(),
	}

	defer h.postings.EnsureOrder()
	return h

}

func convertStats(stats []index.Stat) []stat {
	result := make([]stat, 0, len(stats))
	for _, item := range stats {
		item := stat{Name: item.Name, Value: item.Count}
		result = append(result, item)
	}
	return result
}

func (h *HeadIndexReader) PostingsCardinalityStats() *indexStatus {
	ss := h.postings.Stats("region")

	a := indexStatus{
		LabelValueCount:     convertStats(ss.CardinalityLabelStats),
		LabelValuePairCount: convertStats(ss.LabelValuePairsStats),
	}

	return &a

}

func (h *HeadIndexReader) Reset(newH *HeadIndexReader) {
	h.symMtx.Lock()
	defer h.symMtx.Unlock()
	h.symbols = newH.symbols
	h.values = newH.values
	h.postings = newH.postings

}

func (h *HeadIndexReader) GetGroupByLabel(label string) *labelGroup {
	ss := h.postings.LabelGroup(label)

	a := labelGroup{
		Group: convertStats(ss),
	}

	return &a

}

func (h *HeadIndexReader) GetGroupDistributionByLabel(label string, matchIds []uint64) *labelGroup {
	h.symMtx.RLock()
	ss := h.postings.LabelGroupDistribution(label, matchIds)
	h.symMtx.RUnlock()
	a := labelGroup{
		Group: convertStats(ss),
	}

	return &a

}

func (h *HeadIndexReader) GetHashMap() (map[string]struct{}) {
	h.symMtx.RLock()
	defer h.symMtx.RUnlock()
	return h.symbols
}

func (h *HeadIndexReader) GetOrCreateWithID(id uint64, hash string, lset labels.Labels) {

	h.symMtx.Lock()
	defer h.symMtx.Unlock()

	for _, l := range lset {
		valset, ok := h.values[l.Name]
		if !ok {
			valset = stringset{}
			h.values[l.Name] = valset
		}
		valset.set(l.Value)

	}
	h.symbols[hash] = struct{}{}
	h.postings.Add(id, lset)

}

func (h *HeadIndexReader) DeleteWithIDs(deletedIds map[uint64]struct{}, deletedHashs map[string]struct{}) {
	h.symMtx.Lock()
	defer h.symMtx.Unlock()
	h.postings.Delete(deletedIds)

	for dh, _ := range deletedHashs {
		delete(h.symbols, dh)
	}

	values := make(map[string]stringset, len(h.values))
	if err := h.postings.Iter(func(t labels.Label, _ index.Postings) error {

		ss, ok := values[t.Name]
		if !ok {
			ss = stringset{}
			values[t.Name] = ss
		}
		ss.set(t.Value)
		return nil
	}); err != nil {
		// This should never happen, as the iteration function only returns nil.
		panic(err)
	}
	h.values = values

}

// LabelValues returns label values present in the head for the
// specific label name that are within the time range mint to maxt.
func (h *HeadIndexReader) LabelValues(name string) ([]string, error) {
	h.symMtx.RLock()

	sl := make([]string, 0, len(h.values[name]))
	for s := range h.values[name] {
		sl = append(sl, s)
	}
	h.symMtx.RUnlock()
	return sl, nil
}
func (h *HeadIndexReader) Postings(name string, values ...string) (index.Postings, error) {
	res := make([]index.Postings, 0, len(values))
	for _, value := range values {
		res = append(res, h.postings.Get(name, value))
	}
	return index.Merge(res...), nil
}
