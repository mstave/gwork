package worker

import (
	"fmt"
	"path/filepath"
	"os"
	"log"
)

// Job A piece of work.  Workers, who live in Pools do these
type Job interface {
	Work()
}

// Worker Thing that does Jobs, lives in a Pool.
type Worker struct {
	*Pool
}

// NewWorker Construct mah.  Not making the channels here, so it's easier to have multiple jobs share them to exploit
// parallelism
func NewWorker( pool *Pool) Worker {
	return Worker{ pool }
}

// Start Grab a job from the channel, keep on truckin' until the job channel is closed or there's a stop request
func (w Worker) Start() {
	go func() {
		for {
			select {
			case job := <-w.JobChannel:
				job.Work()
			case <-w.StopChannel:
				return
			}
		}
	}()
}

// RunJob no frills job running, for use outside of a Pool
func RunJob(j Job) {
	j.Work()
}

// StringChanJob Used for performing some transformation of data in one channel and return it as another
// A chan *string variant might save some memory if the strings are long
// io.Reader/Writer or io.Pipe might be a better abstraction, but that may make things harder
// if you wanted to have multiple Jobs operating on the same data in parallel, which channels may make easier
// depending on the semantics of your use case
type StringChanJob struct {
	inChan     chan string
	outChan    chan string
	customWork func(chan string, chan string)
}

// Work Satisfy the Job interface. Having the customWork indirection makes it easier to
// express the work as a function value or closure
func (s *StringChanJob) Work() {
	s.customWork(s.inChan, s.outChan)
}

// trivialIntJob We little example of another kind of job
type trivialIntJob struct {
	inVal int
}

func (i *trivialIntJob) Work() {
	fmt.Println("The int is: ", i.inVal)
}

// FileFinderJob Find all the file name + path under the directories pointed to in dirChan, optionally following
// symlinked directories
type FileFinderJob struct {
	dirChan     chan string // directories to search, as symlinks that are dirs are found, they're added to here
	fileChan    chan string // files found
	followLinks bool        // add symlinks that point to directories to dirChan?
}
// WalkFunc suitable for being called from filePath.Walk, which we do.
// add symlink dirs to channel so they can be searched, which Walk doesn't normally do
func (f *FileFinderJob) WalkFunc(fPath string, fi os.FileInfo, errStart error) error {
	if errStart != nil {
		return errStart
	}
	if f.followLinks {
		if fi.Mode()&os.ModeSymlink != 0 {
			lPath, err := filepath.EvalSymlinks(fPath)
			if err != nil {
				return err
			}
			fi, err = os.Lstat(lPath)
			if err != nil {
				return err
			}
			if fi.IsDir() {
				f.dirChan <- fPath + "/"
				return nil
			}
		}
	}
	if !fi.IsDir() {
		f.fileChan <- fPath
	}
	return nil
}
// Work satisfy the Job interface, kick off the find
func (f *FileFinderJob) Work() {
	for {
		fmt.Println("Hellooooo-------")
		select {
		case dir := <-f.dirChan:
			err := filepath.Walk(dir, f.WalkFunc)
			if err != nil {
				log.Fatal(err)
			}
		default:
			close(f.dirChan)
			close(f.fileChan)
			return
		}
	}
}
