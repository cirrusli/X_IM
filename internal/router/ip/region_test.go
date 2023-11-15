package ip

import (
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"testing"
)

func TestIP2regionSearch(t *testing.T) {
	//region, err := NewSearcher("D:/Gooo/Dev_projects/X_IM/internal/router/data/ip2region.db")
	//t.Log(region, "\nerror: ", err)
	//assert.Nil(t, err)
	//
	//got, err := region.Search("3.166.231.6")
	//t.Log(got)
	//assert.Nil(t, err)
	//t.Log(got)

	//use origin ip2region
	s, err := xdb.NewWithFileOnly("D:/Gooo/Dev_projects/X_IM/internal/router/data/ip2region.db")
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	str, err := s.SearchByStr("127.0.0.1")
	if err != nil {
		t.Error(err)
	}
	t.Log(str)
}
