package server

import "sync"

const (
	jobActServerPing = uint8(iota)
	jobActRsviewParse // todo
	jobActIcqSendMess // todo

	jobStatusCreated = uint8(iota)
	jobStatusPending
	jobStatusFailed
	jobStatusDone
)


type (
	queueJob struct {
		id string
		action uint8
		state uint8
		errors []*apiError
	}
	queueDispatcher struct {
		jobQueue chan *queueJob
		pool chan chan *queueJob

		done chan struct {}
		workerDone chan struct {} }
	queueWorker struct {
		pool chan chan *queueJob
		inbox chan *queueJob

		done chan struct{}
	}
)


func newQueueJob() *queueJob {
	return &queueJob{
		state: jobStatusCreated,
	}
}


func (m *queueJob) newError(e uint8) (err *apiError) {
	err = newApiError(e)
	m.errors = append(m.errors, err)
	return err
}

func (m *queueJob) collectAndSave() {

	var stmtQuery = "INSERT INTO errors (id,job_id,internal_code,displayed_title,displayed_detail) VALUES (?,?,?,?,?)"
	stmt,e := globSqlDB.Prepare(stmtQuery); if e != nil {
		globLogger.Error().Err(e).Msg("[QUEUE]: Could not prepare DB statement!") }

	for _,v := range m.errors {
		globLogger.Error().Uint8("errcode", v.e).Str("detail", apiErrorsDetail[v.e]).Msg("[NOT SAVED!]: " + apiErrorsTitle[v.e])
		if e != nil { continue } // do not save if statement prepare has failed

		_,e = stmt.Exec(v.getId(), m.id, v.e, apiErrorsTitle[v.e], apiErrorsDetail[v.e]); if e != nil {
			globLogger.Error().Err(e).Str("errorid", v.getId()).Msg("[QUEUE][NOT SAVED!]: Could not write error report!") }
	}

	if e == nil { stmt.Close() }
}


func newQueueDispatcher() *queueDispatcher {
	return &queueDispatcher{
		jobQueue: make(chan *queueJob, globConfig.Base.Queue.Chain_Buffer),
		pool: make(chan chan *queueJob, globConfig.Base.Queue.Worker_Capacity),

		done: make(chan struct {}, 1),
		workerDone: make(chan struct {}, 1),
	}
}

func (m *queueDispatcher) getQueueChan() chan *queueJob {
	return m.jobQueue
}

func (m *queueDispatcher) bootstrap() {
	var wg sync.WaitGroup
	wg.Add(globConfig.Base.Queue.Workers + 1)

	for i := 0; i < globConfig.Base.Queue.Workers; i ++ {
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
		select{
			case <-m.done: return
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
		pool: dp.pool,
		inbox: make(chan *queueJob, globConfig.Base.Queue.Worker_Capacity),

		done: dp.workerDone,
	}
}

func (m *queueWorker) spawn() {

	defer close(m.inbox)

	for {

		m.pool <- m.inbox

		select {
			case <-m.done: return
			case buf := <-m.inbox: m.doJob(buf)
		}
	}
}

func (m *queueWorker) doJob(jb *queueJob) {

	//
}
