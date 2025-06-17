// Functions related to a person's job but managed in separate data source
package fakery

var jobLoader DataLoader

func init() {
	jobLoader.Init("jobs.json")
}

type Job struct {
	// title/name of the job
	Title      string `json:"title"`
	Profession string `json:"profession,omitempty"`
}

func (f *Fakery) Job() *Job {
	var job string

	data := f.LoadGenericLocale(&jobLoader)
	job = f.RandomString(data.Get("jobs"))
	// Right now not handling profession
	return &Job{Title: job}
}
