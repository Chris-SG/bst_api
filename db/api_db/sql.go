package api_db

import (
	"fmt"
	"github.com/chris-sg/bst_api/models/api_models"
	"github.com/chris-sg/bst_api/models/bst_models"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"time"
)

type ApiDbCommunication interface {
	SetProfile(profile bst_models.BstProfile) (errs []error)

	RetrieveProfile(user string) (profile bst_models.BstProfile, errs []error)
	RetrieveUpdateableProfiles() (profiles []bst_models.BstProfile, errs []error)

}

func CreateApiDbCommunicationPostgres(db *gorm.DB) ApiDbCommunicationPostgres {
	return ApiDbCommunicationPostgres{db}
}

type ApiDbCommunicationPostgres struct {
	db *gorm.DB
}

func (dbcomm ApiDbCommunicationPostgres) SetProfile(profile bst_models.BstProfile) (errs []error) {
	resultDb := dbcomm.db.Model(&bst_models.BstProfile{}).Save(&profile)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm ApiDbCommunicationPostgres) RetrieveProfile(user string) (profile bst_models.BstProfile, errs []error) {
	glog.Infof("bst profile for %s", user)
	p := make([]bst_models.BstProfile, 0)
	resultDb := dbcomm.db.Model(&bst_models.BstProfile{}).Where("user_sub = ?", user).Scan(&p)
	glog.Infof("%d results for %s", len(p), user)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}

	if len(p) > 0 {
		profile = p[0]
	}

	return
}

func (dbcomm ApiDbCommunicationPostgres) RetrieveUpdateableProfiles() (profiles []bst_models.BstProfile, errs []error) {
	resultDb := dbcomm.db.Table("public.\"bstProfile\" p").
		Joins("public.\"eaGateUser\" e on " +
			"p.user_sub = e.web_user and " +
			"e.login_cookie <> '' and " +
			"e.subscription <> ''").
		Scan(&profiles)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	
	return
}


// AddAutomaticJob will create a new job.
func AddAutomaticJob(db *gorm.DB, job api_models.AutomaticJob) error {
	if !db.NewRecord(job) {
		return fmt.Errorf("job %s already exists", job.JobName)
	}
	db.Create(&job)
	return nil
}

// RetrievePendingJobs will return a slice of all AutomaticJobs in the
// database that are enabled and due to run at any time.
func RetrievePendingJobs(db *gorm.DB) []api_models.AutomaticJob {
	var jobs []api_models.AutomaticJob
	now := time.Now()
	db.Model(&api_models.AutomaticJob{}).Where("enabled = ? AND next_run <= ?", true, now).Scan(&jobs)
	return jobs
}

// RetrieveNamedJobs will return a slice of all AutomaticJobs that match
// the jobNames in the string slice provided.
func RetrieveNamedJobs(db *gorm.DB, jobNames []string) []api_models.AutomaticJob {
	var jobs []api_models.AutomaticJob
	db.Model(&api_models.AutomaticJob{}).Where("job_name IN (?)", jobNames).Scan(&jobs)
	return jobs
}

// RetrieveAllJobs will return a slice with all AutomaticJobs in the
// database.
func RetrieveAllJobs(db *gorm.DB) []api_models.AutomaticJob {
	var jobs []api_models.AutomaticJob
	db.Model(&api_models.AutomaticJob{}).Scan(&jobs)
	return jobs
}

// ActivateJobs should be called on any jobs that are to be triggered.
// This will increment their count, as well as set the next run and
// last run times.
func ActivateJobs(db *gorm.DB, jobNames []string) {
	var jobs []api_models.AutomaticJob
	db.Model(&api_models.AutomaticJob{}).Where("job_name IN (?)", jobNames).Scan(&jobs)

	for i := range jobs {
		jobs[i].Count++
		jobs[i].LastRun = jobs[i].NextRun
		jobs[i].NextRun.Add(jobs[i].Frequency)
	}

	db.Save(&jobs)
}

// UpdateJob will update an existing job, or create the job if it does not
// yet exist.
func UpdateJob(db *gorm.DB, job api_models.AutomaticJob) error {
	err := db.Save(&job).Error
	if err != nil {
		return err
	}
	return nil
}

// ToggleJob will either enable or disable a job, depending on its current state.
func ToggleJob(db *gorm.DB, jobName string) {
	job := api_models.AutomaticJob{}
	db.First(&job)
	job.Enabled = !job.Enabled
	db.Save(&job)
}

// DeleteJob will remove a job from the database.
func DeleteJob(db *gorm.DB, jobName string) {
	job := api_models.AutomaticJob{
		JobName: jobName,
	}
	db.Delete(&job)
}