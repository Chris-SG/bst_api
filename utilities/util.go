package utilities

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
	"net/http"
	"reflect"
	"strings"
)

func RespondWithError(rw http.ResponseWriter, err bst_models.Error) {
	bytes, _ := json.Marshal(err)
	rw.WriteHeader(err.CorrespondingHttpCode)
	_, _ = rw.Write(bytes)
	return
}

// ValidateOrdering will ensure columns in orderRequest are correct for the
// provided interfae. Additionally, it will only allow ASC and DESC.
// TODO: support fields such as LENGTH(id)
func ValidateOrdering(i interface{}, orderRequest []string) (ordering string) {
	t := reflect.TypeOf(i)

	validFields := make(map[string]interface{}, 0)

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		gormTag := tag.Get("gorm")
		gormTagSections := strings.Split(gormTag, ";")
		for _, section := range gormTagSections {
			if strings.Contains(section, "column:") {
				col := strings.Split(section, ":")
				validFields[col[1]] = nil
				break
			}
		}
	}

	toJoin := make([]string, 0)
	for _, req := range orderRequest {
		s := strings.Split(req, " ")
		if len(s) > 2 {
			continue
		}
		if _, exists := validFields[s[0]]; !exists {
			continue
		}
		if len(s) == 2 {
			if strings.ToUpper(s[1]) == "ASC" || strings.ToUpper(s[1]) == "DESC" {
				toJoin = append(toJoin, s[0] + " " + strings.ToUpper(s[1]))
			}
		} else {
			toJoin = append(toJoin, s[0])
		}
	}

	ordering = strings.Join(toJoin, ", ")
	return
}

// ValidateFiltering will ensure filtered select statements will only
// use valid rows from the interface.
// TODO: Do validation on clauses.
func ValidateFiltering(i interface{}, filterRequest []string) (filtering string) {
	t := reflect.TypeOf(i)

	validFields := make(map[string]interface{}, 0)

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		gormTag := tag.Get("gorm")
		gormTagSections := strings.Split(gormTag, ";")
		for _, section := range gormTagSections {
			if strings.Contains(section, "column:") {
				col := strings.Split(section, ":")
				validFields[col[1]] = nil
				break
			}
		}
	}

	toJoin := make([]string, 0)
	for _, req := range filterRequest {
		s := strings.Split(req, " ")
		if _, exists := validFields[s[0]]; !exists {
			continue
		}
		toJoin = append(toJoin, req)
	}

	filtering = strings.Join(toJoin, " AND ")
	return
}

func PrintErrors(errMsg string, errs []error) bool {
	if len(errs) > 0 {
		glog.ErrorDepth(1, fmt.Sprintf("%s\n", errMsg))
		for _, err := range errs {
			glog.ErrorDepth(1, fmt.Sprintf("\t%s\n", err.Error()))
		}
		return true
	}
	return false
}

func CleanString(in string) string {
	return strings.ReplaceAll(in, "'", "&#39;")
}