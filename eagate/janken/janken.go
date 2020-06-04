package janken

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/bst_api/eagate/util"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
	"math/rand"
)

func PlayJanken(client util.EaClient) (timesPlayed int, err bst_models.Error) {
	err = bst_models.ErrorOK
	const jankenResource = "/game/bemani/bjm2020/janken/index.html"
	jankenUri := util.BuildEaURI(jankenResource)

	running := true
	for running {
		content, _, err2 := util.GetPageContentAsGoQuery(client.Client, jankenUri)
		if err2 != bst_models.ErrorOK {
			glog.Errorf("failed to get janken page: %s", err2.Message)
			return
		}
		selection := content.Find("div#janken-select div.inner a")
		if selection == nil {
			running = false
			break
		}
		if selection.Length() != 3 {
			running = false
			break
		}
		choice := rand.Int() % 3
		selection.Each(func(i int, s *goquery.Selection) {
			if i != choice {
				return
			}
			attemptResource, success := s.Attr("href")
			if !success {
				glog.Warningf("fialed to get href from selection")
				return
			}
			attemptUri := util.BuildEaURI(attemptResource)
			client.Client.Get(attemptUri)
			timesPlayed++
		})
	}
	return
}