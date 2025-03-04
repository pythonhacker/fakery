// Functions related to a person's job but managed in separate data source
package gofakelib

var jdata DataLoader

func init() {
	jdata.Init("jobs.json")
}

type Job struct {
	// title/name of the job
	Title      string `json:"title"`
	Profession string `json:"profession,omitempty"`
}

func (f *Faker) Job() *Job {
	var job string

	// job data is not locale specific
	localeData := jdata.EnsureLoaded(GenericLocale)
	job = f.RandomString(localeData.Get("jobs"))
	// Right now not handling profession
	return &Job{Title: job}
}
