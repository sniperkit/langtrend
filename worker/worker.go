package worker

import (
	"log"
	"time"
)

type Worker struct {
	Options Options
}

func New(options Options) *Worker {
	return &Worker{options}
}

func (w *Worker) Run() {
	log.Print("Worker running")
	if waiting, err := w.Options.Db.WaitingJobCount(); err != nil {
		log.Printf("Error getting waiting job count, %s", err)
	} else if waiting == 0 {
		log.Print("No jobs found, enqueuing GitHub jobs")
		if err := w.EnqueueCreateGHJobs(time.Now()); err != nil {
			log.Printf("Error enqueuing GitHub jobs, %s", err)
		}
	}
	for {
		job, ok, err := w.Options.Db.NextJob()
		if err != nil {
			log.Printf("Error fetching next job, %s, waiting 10 seconds", err)
			time.Sleep(10 * time.Second)
			continue
		}
		if !ok {
			time.Sleep(5 * time.Second)
			continue
		}
		if err := w.RunJob(job); err == nil {
			if err := w.Options.Db.JobComplete(job.Id); err != nil {
				log.Printf("Error marking job complete, %s", err)
			}
		} else {
			log.Printf("Error running job, %s", err)
			if err := w.Options.Db.JobFailed(job.Id, err); err != nil {
				log.Printf("Error marking job failed, %s", err)
			}
		}
	}
}
