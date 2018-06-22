package server

import "github.com/satori/go.uuid"

const (
	taskTypePuppetCertDestroy = uint8(iota)
	taskTypeInstallerStart
)

type baseTask struct {
	id string
	action uint8
}


func newTask() *baseTask {
	return &baseTask{
		id: uuid.NewV4().String(),
	}
}

func (m *baseTask) save() (bool,*appError) {

	if _,e := globSqlDB.Exec("INSERT INTO tasks (id, type) VALUES (?, ?)", m.id, m.action); e != nil {
		return false,newAppError(errInternalSqlError).log(e, "")
	}

	return true,nil
}

func (m *baseTask) update() *appError {

	if _,e := globSqlDB.Exec("UPDATE tasks SET id = ?, type = ?", m.id, m.action); e != nil {
		return newAppError(errInternalSqlError).log(e, "")
	}

	return nil
}

func getTaskByHost(hId string) (*baseTask,*appError) {

	rws,e := globSqlDB.Query("SELECT id, type FROM tasks where host = ? LIMIT 2", hId)
	if e != nil {
		return nil,newAppError(errInternalSqlError).log(e, "")
	}
	defer rws.Close()

	if !rws.Next() {
		if rws.Err() != nil {
			return nil,newAppError(errInternalSqlError).log(e, "")
		}

		return nil,nil
	}

	var tsk = &baseTask{}
	if e = rws.Scan(&tsk.id, &tsk.action); e != nil {
		return nil,newAppError(errInternalSqlError).log(e, "")
	}

	if rws.Next() {
		return nil,newAppError(errInternalCommonError).log(e, "Found two or more records with same host! Database is broken!")
	}

	return tsk,nil
}
