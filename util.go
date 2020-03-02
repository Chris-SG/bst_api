package main

import "github.com/chris-sg/bst_server_models/bst_api_models"

func WriteStatus(status string, message string) bst_api_models.Status {
	return bst_api_models.Status{
		Status:  status,
		Message: message,
	}
}