package server

import "sync"
import "time"
import "net/http"
import "github.com/satori/go.uuid"
import "github.com/gorilla/context"

const (
	jobActServerPing = uint8(iota)
	jobActRequestHostCreate
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
		jobActRequestHostCreate: "Processing the received request to create a host",
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

func newQueueJob(reqId *string, act uint8) (jb *queueJob, e error) {

	jb = &queueJob{
		id:           uuid.NewV4().String(),
		state:        jobStatusCreated,
		action:       act,
		requested_by: *reqId,
		updated_at:   time.Now(),
		created_at:   time.Now()}

	_, e = globSqlDB.Exec(
		"INSERT INTO jobs (id, requested_by, action, updated_at, created_at) VALUES (?,?,?,?,?)",
		jb.id, jb.requested_by, jb.action,
		jb.updated_at.Format("2006-01-02 15:04:05.999999"), jb.created_at.Format("2006-01-02 15:04:05.999999"))

	return
}

func getJobById(req *http.Request) (jb *queueJob) {

	jb = new(queueJob)
	var r = context.Get(req, "internal_request").(*httpRequest)
	var jobId = context.Get(req, "param_jobid").(string)

	stmt, e := globSqlDB.Prepare("SELECT action,state,updated_at,created_at FROM jobs WHERE id=? LIMIT 2")
	if e != nil {
		r.newError(errInternalSqlError).log(e, "[QUEUE]: Could not prepare DB statement!")
		return
	}
	defer stmt.Close()

	rows, e := stmt.Query(jobId)
	if e != nil {
		r.newError(errInternalSqlError).log(e, "[QUEUE]: Could not get result from DB!")
		return
	}
	defer rows.Close()

	if !rows.Next() {
		if e = rows.Err(); e != nil {
			r.newError(errInternalSqlError).log(e, "[QUEUE]: Could not exec rows.Next method!")
			return
		}
		r.newError(errJobsJobNotFound).log(nil, "[QUEUE]: The requested job was not found!")
		return
	}

	if e = rows.Scan(&jb.action, &jb.state, &jb.updated_at, &jb.created_at); e != nil {
		r.newError(errInternalSqlError).log(e, "[QUEUE]: Could not scan the result from DB!")
		return
	}

	if rows.Next() {
		r.newError(errInternalSqlError).log(nil, "[QUEUE]: Rows is not equal to 1. The DB has broken!")
		return
	}

	jb.id = jobId
	return
}

func (m *queueJob) newError(e uint8) (ae *apiError) {

	if !m.is_failed {
		m.setFailed()
	}

	ae = newApiError(e)

	_, err := globSqlDB.Exec(
		"INSERT INTO errors (id,job_id,internal_code,displayed_title,displayed_detail) VALUES (?,?,?,?,?)",
		ae.getId(), m.id, ae.e, apiErrorsTitle[ae.e], apiErrorsDetail[ae.e])

	if err != nil {
		globLogger.Error().Uint8("code", ae.e).Str("detail", apiErrorsDetail[ae.e]).Msg("[NOT SAVED]: " + apiErrorsTitle[ae.e])
		return
	}

	return
}

func (m *queueJob) setFailed() {

	m.is_failed = true

	_, e := globSqlDB.Exec(
		"UPDATE jobs SET is_failed = 1 WHERE id=?", m.id)

	if e != nil {
		m.newError(errInternalSqlError).log(e, "[QUEUE]: Could not update job's failed flag!")
		return
	}
}

func (m *queueJob) stateUpdate(state uint8) {

	m.state = state

	_, e := globSqlDB.Exec(
		"UPDATE jobs SET state = ? WHERE id = ?", state, m.id)

	if e != nil {
		m.newError(errInternalSqlError).log(e, "[QUEUE]: Could not update job's state!")
		return
	}
}

func (m *queueJob) setPayload(pl *map[string]interface{}) {
	m.payload = pl
}

func (m *queueJob) addToQueue() {
	globQueueChan <- m
}

func newQueueDispatcher() *queueDispatcher {
	return &queueDispatcher{
		jobQueue: make(chan *queueJob, globConfig.Base.Queue.Chain_Buffer),
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
				nextWorker = <-m.pool
				nextWorker <- job
			}(buf)
		}
	}
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
	globLogger.Debug().Msg("LOL! JOB RECEIVED!")

	switch jb.action {
	case jobActRequestHostCreate:

		var payload map[string]interface{} = *jb.payload

		var host = payload["job_input_host"].(*baseHost)
		var macs = payload["job_input_macs"].([]string)

		globLogger.Info().Str("ipmi_ip", host.ipmi_address.String()).Msg("Job found new IPMI IP address!")

		for _, v := range macs {
			globLogger.Info().Str("mac", v).Msg("job found new mac address!")
		}
	default:
		globLogger.Warn().Msg("Unknown job type!")
	}
}
