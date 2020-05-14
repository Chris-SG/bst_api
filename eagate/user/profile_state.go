package user

import (
	"github.com/chris-sg/bst_api/eagate/util"
	bst_models "github.com/chris-sg/bst_server_models"
)

func ProfileEaSubscriptionState(client util.EaClient) (subscriptionType string, err bst_models.Error) {
	err = bst_models.ErrorOK
	const paybookResource = "/payment/mybook/paybook.html"
	PaybookUri := util.BuildEaURI(paybookResource)

	document, err := util.GetPageContentAsGoQuery(client.Client, PaybookUri)
	eaSubSelection := document.Find("div#id_paybook_all .cl_course_name").First()

	if eaSubSelection == nil {
		err = bst_models.ErrorBadCookie
		return
	}

	subscriptionType = eaSubSelection.Text()
	return
}