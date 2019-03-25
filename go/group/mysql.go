// Provide a MySQL backend as alternative to raft consensus

// Expect the following backend tables:
/*

CREATE TABLE service_election (
  domain varchar(32) NOT NULL,
  service_id varchar(128) NOT NULL,
  last_seen_active timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (domain)
);


*/
package group

import (
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/github/freno/go/config"

	"github.com/outbrain/golib/sqlutils"
	metrics "github.com/rcrowley/go-metrics"
)

type MySQLBackend struct {
	db          *sql.DB
	domain      string
	serviceId   string
	leaderState int64
}

const maxConnections = 3

func NewMySQLBackend() (*MySQLBackend, error) {
	if config.Settings().BackendMySQLHost == "" {
		return nil, nil
	}
	dbUri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true&charset=utf8mb4,utf8,latin1&timeout=500ms",
		config.Settings().BackendMySQLUser, config.Settings().BackendMySQLPassword, config.Settings().BackendMySQLHost, config.Settings().BackendMySQLPort, config.Settings().BackendMySQLSchema,
	)
	db, _, err := sqlutils.GetDB(dbUri)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxConnections)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	backend := &MySQLBackend{
		db:        db,
		domain:    fmt.Sprintf("%s:%s", config.Settings().DataCenter, config.Settings().Environment),
		serviceId: hostname,
	}
	go backend.continuousElections()
	return backend, nil
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Monitor is a utility function to routinely observe leadership state.
// It doesn't actually do much; merely takes notes.
func (backend *MySQLBackend) continuousElections() {
	t := time.NewTicker(1 * time.Second)

	for range t.C {
		backend.AttemptLeadership()
		leaderState, err := backend.ReadLeadership()
		if err == nil {
			// otherwise maintain state
			atomic.StoreInt64(&backend.leaderState, leaderState)
		}
		go metrics.GetOrRegisterGauge("backend.mysql.is_leader", nil).Update(leaderState)
		go metrics.GetOrRegisterGauge("backend.mysql.is_healthy", nil).Update(boolToInt64(err == nil))
	}
}

func (backend *MySQLBackend) IsLeader() bool {
	return atomic.LoadInt64(&backend.leaderState) > 0
}

func (backend *MySQLBackend) AttemptLeadership() error {
	query := `
    insert ignore into service_election (
        domain, service_id, last_seen_active
      ) values (
        ?, ?, now()
      ) on duplicate key update
      service_id = if(last_seen_active < now() - interval 20 second, values(service_id), service_id),
      last_seen_active = if(service_id = values(service_id), values(last_seen_active), last_seen_active)
  `
	args := sqlutils.Args(backend.domain, backend.serviceId)
	_, err := sqlutils.ExecNoPrepare(backend.db, query, args...)
	return err
}

func (backend *MySQLBackend) ForceLeadership() error {
	query := `
    replace into service_election (
        domain, service_id, last_seen_active
      ) values (
        ?, ?, now()
      )
  `
	args := sqlutils.Args(backend.domain, backend.serviceId)
	_, err := sqlutils.ExecNoPrepare(backend.db, query, args...)
	return err
}

func (backend *MySQLBackend) Reelect() error {
	query := `
    delete from service_election where domain=?
  `
	args := sqlutils.Args(backend.domain)
	_, err := sqlutils.ExecNoPrepare(backend.db, query, args...)
	return err
}

func (backend *MySQLBackend) ReadLeadership() (int64, error) {
	query := `
    select count(*) > 0
      from service_election
      where domain=?
      and service_id=?
  `
	args := sqlutils.Args(backend.domain, backend.serviceId)

	var count int64
	err := backend.db.QueryRow(query, args).Scan(&count)

	return count, err
}
