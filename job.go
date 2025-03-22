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

	data := f.LoadGenericLocale(&jdata)
	job = f.RandomString(data.Get("jobs"))
	// Right now not handling profession
	return &Job{Title: job}
}
