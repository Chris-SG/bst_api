package ddr

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"strconv"
	"strings"
)

func musicDataSingleDocument(client util.EaClient, pageNumber int) (document *goquery.Document, err error) {
	const musicDataSingleResource = "/game/ddr/ddra20/p/playdata/music_data_single.html?offset={page}&filter=0&filtertype=0&sorttype=0"
	musicDataURI := util.BuildEaURI(musicDataSingleResource)

	currentPageURI := strings.Replace(musicDataURI, "{page}", strconv.Itoa(pageNumber), -1)
	document, err = util.GetPageContentAsGoQuery(client.Client, currentPageURI)
	return
}

func musicDetailDocument(client util.EaClient, songId string) (document *goquery.Document, err error) {
	const baseDetail = "/game/ddr/ddra20/p/playdata/music_detail.html?index="
	musicDetailURI := util.BuildEaURI(baseDetail)

	musicDetailURI += songId
	document, err = util.GetPageContentAsGoQuery(client.Client, musicDetailURI)
	return
}

func musicDetailDifficultyDocument(client util.EaClient, songId string, mode ddr_models.Mode, difficulty ddr_models.Difficulty) (document *goquery.Document, err error) {
	const baseDetail = "/game/ddr/ddra20/p/playdata/music_detail.html?index={id}&diff={diff}"
	musicDetailURI := util.BuildEaURI(baseDetail)

	difficultyId := int(difficulty)
	if mode == ddr_models.Double {
		difficultyId += 4
	}

	musicDetailURI = strings.Replace(musicDetailURI, "{id}", songId, -1)
	musicDetailURI = strings.Replace(musicDetailURI, "{diff}", strconv.Itoa(difficultyId), -1)
	document, err = util.GetPageContentAsGoQuery(client.Client, musicDetailURI)
	return
}

func playerInformationDocument(client util.EaClient) (document *goquery.Document, err error) {
	const playerInformationResource = "/game/ddr/ddra20/p/playdata/index.html"
	playerInformationUri := util.BuildEaURI(playerInformationResource)

	document, err = util.GetPageContentAsGoQuery(client.Client, playerInformationUri)
	return
}

func recentScoresDocument(client util.EaClient) (document *goquery.Document, err error) {
	const recentSongsResource = "/game/ddr/ddra20/p/playdata/music_recent.html"
	recentSongsUri := util.BuildEaURI(recentSongsResource)

	document, err = util.GetPageContentAsGoQuery(client.Client, recentSongsUri)
	return
}

func workoutDocument(client util.EaClient) (document *goquery.Document, err error) {
	const workoutResource = "/game/ddr/ddra20/p/playdata/workout.html"
	workoutUri := util.BuildEaURI(workoutResource)

	document, err = util.GetPageContentAsGoQuery(client.Client, workoutUri)
	return
}