package mem_index

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"stree-index/pkg"
	"stree-index/pkg/config"
	"stree-index/pkg/testutil"
	"stree-index/pkg/inverted-index"
	"stree-index/pkg/inverted-index/labels"
	"stree-index/pkg/inverted-index/index"
)

func TestMysqlGet(t *testing.T) {
	sc := &config.MysqlServerConfig{
		Host:         "localhost:3306",
		Username:     "root",
		Password:     "mysql123",
		Dbname:       "local_test",
		LogPrint:     true,
		MaxIdleConns: 10,
		MaxOpenConns: 1000,
	}
	pkg.InitDb(sc)
	InitIdx()
	logger := log.NewLogfmtLogger(os.Stdout)

	level.Warn(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
	FlushEcsIdxAdd(logger, []uint64{})
	//CommonQuery(logger)
	matchers1 := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "k2", "v2"),
	}
	//matchers2 := []*labels.Matcher{
	//	labels.MustNewMatcher(labels.MatchNotEqual, "b", "1"),
	//}
	//matchers3 := []*labels.Matcher{
	//	labels.MustNewMatcher(labels.MatchRegexp, "a", "1|2"),
	//}
	//
	//matchers4 := []*labels.Matcher{
	//	labels.MustNewMatcher(labels.MatchRegexp, "c", "[0-9]+"),
	//}
	//
	//matchers5 := []*labels.Matcher{
	//	labels.MustNewMatcher(labels.MatchNotRegexp, "c", "[0-9]+"),
	//}

	// equal
	p1, e := inverted_index.PostingsForMatchers(EcsIdx, matchers1...)
	testutil.Ok(t, e)
	p1s, e := index.ExpandPostings(p1)
	testutil.Ok(t, e)
	tt1 := []uint64{1, 2}
	testutil.Equals(t, p1s, tt1)
}
