package server

import "sync"
import "time"
import "github.com/satori/go.uuid"

const (
	jobActServerPing = uint8(iota)
	jobActHostCreate
	jobActRsviewParse // todo
	jobActIcqSendMess // todo
)
const (
	jobStatusCreated = uint8(iota)
	jobStatusPending
	jobStatusFailed
	jobStatusDone
)

var (
	jobActHumanDetail = map[uint8]string{
		jobActServerPing:        "Server ping",
		jobActHostCreate: "Processing the received request to create a host",
		jobActRsviewParse:       "Rsview parsing",
		jobActIcqSendMess:       "ICQ message sending",
	}

	jobStatusHumanDetail = map[uint8]string{
		jobStatusCreated: "Created",
		jobStatusPending: "Pending",
		jobStatusFailed:  "Failed",
		jobStatusDone:    "Done",
	}
)

type (
	queueJob struct {
		payload *map[string]interface{}
		fail_count int
		errors []*appError

		id           string
		requested_by string
		action       uint8
		state        uint8
		is_failed    bool
		updated_at   time.Time
		created_at   time.Time
	}
	queueDispatcher struct {
		jobQueue chan *queueJob
		pool     chan chan *queueJob

		done       chan struct{}
		workerDone chan struct{}
	}
	queueWorker struct {
		pool  chan chan *queueJob
		inbox chan *queueJob

		done chan struct{}
	}
)

func newQueueJob(reqId *string, act uint8) (*queueJob, *appError) {

	var jb = &queueJob{
		id:           uuid.NewV4().String(),
		state:        jobStatusCreated,
		action:       act,
		requested_by: *reqId,
		updated_at:   time.Now(),
		created_at:   time.Now()}

	if _,e := globSqlDB.Exec(
		"INSERT INTO jobs (id, requested_by, action, updated_at, created_at) VALUES (?,?,?,?,?)",
		jb.id, jb.requested_by, jb.action,
		jb.updated_at.Format("2006-01-02 15:04:05.999999"), jb.created_at.Format("2006-01-02 15:04:05.999999")); e != nil {

		return nil,newAppError(errInternalCommonError).log(e, "Could not create a new job because of a database error!")
	}

	return jb,nil
}

func getJobById(jobId string) (*queueJob, *appError) {

	jb := new(queueJob)

	rws,e := globSqlDB.Query("SELECT action,state,updated_at,created_at FROM jobs WHERE id=? LIMIT 2", jobId);
	if e != nil {
		return nil,newAppError(errInternalSqlError).log(e, "Could not get result from DB!")
	}
	defer rws.Close()

	if ! rws.Next() {
		if rws.Err() != nil {
			return nil,newAppError(errInternalSqlError).log(rws.Err(), "Could not exec rows.Next method!")
		}
		return nil,newAppError(errJobsJobNotFound).log(nil, "The requested job was not found!")
	}

	if e = rws.Scan(&jb.action, &jb.state, &jb.updated_at, &jb.created_at); e != nil {
		return nil,newAppError(errInternalSqlError).log(e, "Could not scan the result from DB")
	}

	if rws.Next() {
		return nil,newAppError(errInternalSqlError).log(nil, "Rows is not equal to 1. The DB has broken!")
	}

	jb.id = jobId
	return jb,nil
}

func (m *queueJob) appendAppError(aErr *appError) *appError {

	m.errors = append(m.errors, aErr.setJobId(m.id))

	if len(m.errors) == globConfig.Base.Queue.Max_Job_Fails {
		globLogger.Error().Str("job_id", m.id).Str("job_action", jobActHumanDetail[m.action]).
			Msg("The job has reached the maximum number of failures!")

		m.setFailed()

		for _,v := range m.errors {
			v.save()
		}

	// TODO: add icq notify

		return aErr
	}

	// TODO: add interval between job starts

	m.addToQueue()
	return aErr
}

func (m *queueJob) setFailed() *appError {

	m.is_failed = true
	if err := m.stateUpdate(jobStatusFailed); err != nil {
		return err
	}

	if _,e := globSqlDB.Exec("UPDATE jobs SET is_failed = 1 WHERE id=?", m.id); e != nil {
		return newAppError(errInternalSqlError).log(e, "Could not exec the database query!")
	}

	return nil
}

func (m *queueJob) stateUpdate(state uint8) *appError {

	m.state = state

	if _,e := globSqlDB.Exec("UPDATE jobs SET state = ? WHERE id = ?", state, m.id); e != nil {
		return newAppError(errInternalSqlError).log(e, "Could not exec the database query!")
	}

	return nil
}

func (m *queueJob) setPayload(pl *map[string]interface{}) {
	m.payload = pl
}

func (m *queueJob) addToQueue() {
	globQueueChan <- m
}

func newQueueDispatcher() *queueDispatcher {
	return &queueDispatcher{
		jobQueue: make(chan *queueJob, globConfig.Base.Queue.Jobs_Chain_Buffer),
		pool:     make(chan chan *queueJob, globConfig.Base.Queue.Worker_Capacity),

		done:       make(chan struct{}, 1),
		workerDone: make(chan struct{}, 1),
	}
}

func (m *queueDispatcher) getQueueChan() chan *queueJob {
	return m.jobQueue
}

func (m *queueDispatcher) bootstrap() {
	var wg sync.WaitGroup
	wg.Add(globConfig.Base.Queue.Workers + 1)

	for i := 0; i < globConfig.Base.Queue.Workers; i++ {
		go func(wg sync.WaitGroup) {
			newQueueWorker(m).spawn()
			wg.Done()
		}(wg)
	}

	go func(wg sync.WaitGroup) {
		m.dispatch()
		close(m.workerDone)
		wg.Done()
	}(wg)

	wg.Wait()
}

func (m *queueDispatcher) dispatch() {

	var buf *queueJob
	var nextWorker chan *queueJob

	for {
		select {
		case <-m.done:
			return
		case buf = <-m.jobQueue:
			go func(job *queueJob) {

				if err := job.stateUpdate(jobStatusPending); err != nil {
					job.appendAppError(err)
					return
				}

				nextWorker = <-m.pool
				nextWorker <- job
			}(buf)
		}
	}


	// TODO: add sync.WaitGroup
	// BUG: jobQueue without close()
}

func (m *queueDispatcher) destruct() {
	close(m.done)
}

func newQueueWorker(dp *queueDispatcher) *queueWorker {
	return &queueWorker{
		pool:  dp.pool,
		inbox: make(chan *queueJob, globConfig.Base.Queue.Worker_Capacity),

		done: dp.workerDone,
	}
}

func (m *queueWorker) spawn() {

	defer close(m.inbox)

	for {

		m.pool <- m.inbox

		select {
		case <-m.done:
			return
		case buf := <-m.inbox:
			m.doJob(buf)
		}
	}
}

func (m *queueWorker) doJob(jb *queueJob) {
	globLogger.Debug().Uint8("job_code", jb.action).Str("code_human", jobActHumanDetail[jb.action]).
		Msg("The worker received a new job!")

	// get payload for job handler:
	var payload map[string]interface{} = *jb.payload

	// match job handler and exec it:
	switch jb.action {
	case jobActHostCreate:

		var host = payload["job_payload_host"].(*baseHost)

		if e := host.resolveIpmiHostname(); e != nil {
			jb.appendAppError(e)
			break
		}

		if e := host.updateOrCreate(jb.id); e != nil {
			jb.appendAppError(e)
			break
		}

		jb.stateUpdate(jobStatusDone)

	default:
		globLogger.Warn().Msg("Unknown job type!")
	}
}
