package common

import "stree-index/pkg/inverted-index/labels"

func FormatLabelMatchers(ls []*SingleTagReq) []*labels.Matcher {

	matchers := make([]*labels.Matcher, 0)
	for _, i := range ls {

		mType, ok := labels.MatchMap[i.Type];
		if !ok {
			continue
		}

		matchers = append(

			matchers,
			labels.MustNewMatcher(mType, i.Key, i.Value),

		)
	}
	return matchers
}
