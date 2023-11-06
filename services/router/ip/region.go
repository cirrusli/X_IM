package ip

import (
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

//type IpInfo struct {
//	Country string
//	Region  string
//	City    string
//	ISP     string
//}

type Region interface {
	Search(ip string) (string, error)
}

type Searcher struct {
	searcher *xdb.Searcher
}

func NewSearcher(path string) (Searcher, error) {
	if path == "" {
		path = "ip2region.db"
	}
	searcher, err := xdb.NewWithFileOnly(path)
	if err != nil {
		return struct{ searcher *xdb.Searcher }{searcher: nil}, err
	}
	defer searcher.Close()
	return Searcher{}, err

}

func (r *Searcher) Search(ip string) (string, error) {
	region, err := r.searcher.SearchByStr(ip)
	if err != nil {
		return "", err
	}
	return region, nil
	//return &IpInfo{
	//	Country: info.Country,
	//	Region:  info.Region,
	//	City:    info.City,
	//	ISP:     info.ISP,
	//}, nil
}
