package db

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// SportsRepo provides repository access to sports.
type SportsRepo interface {
	// Init will initialise our sports repository.
	Init() error

	// List will return a list of events.
	List(filter *sports.ListEventsRequestFilter) ([]*sports.Event, error)
}

type sportsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewSportsRepo creates a new sports repository.
func NewSportsRepo(db *sql.DB) SportsRepo {
	return &sportsRepo{db: db}
}

// Init prepares the sport repository dummy data.
func (s *sportsRepo) Init() error {
	var err error

	s.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy sports.
		err = s.seed()
	})

	return err
}

func (s *sportsRepo) List(filter *sports.ListEventsRequestFilter) ([]*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventQueries()[eventsList]
	query, args = s.applyFilter(query, filter)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return s.scanSportsEvent(rows)
}

func (s *sportsRepo) applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if filter.Id > 0 {
		id := strconv.Itoa(int(filter.Id))
		clauses = append(clauses, "id = "+id+" and")
	}

	if filter.Visible == true {
		clauses = append(clauses, "visible = 1")
	} else {
		clauses = append(clauses, "visible = 0")
	}

	if len(clauses) != 0 && len(filter.Column) == 0 && len(filter.OrderBy) == 0 {
		query += " WHERE " + strings.Join(clauses, " ")
	}

	if len(filter.Column) > 0 && len(filter.OrderBy) > 0 {
		clauses = append(clauses, "ORDER BY "+filter.Column+" "+filter.OrderBy)
		query += " WHERE " + strings.Join(clauses, " ")
	}

	// check sql correct or not
	// log.Println(filter, query)
	return query, args
}

func (s *sportsRepo) scanSportsEvent(rows *sql.Rows) ([]*sports.Event, error) {
	var allEvents []*sports.Event

	for rows.Next() {
		var event sports.Event
		var advertisedStart time.Time
		var eventStart time.Time
		var eventEnd time.Time

		if err := rows.Scan(&event.Id, &event.Name, &event.Result, &event.Location, &event.Visible, &eventStart, &eventEnd, &advertisedStart); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}

			return nil, err
		}

		event.StartTime = timestamppb.New(eventStart)
		event.EndTime = timestamppb.New(eventEnd)
		ts := timestamppb.New(advertisedStart)
		event.AdvertisedStartTime = ts

		if time.Now().Unix() > ts.AsTime().Unix() {
			event.Status = "CLOSED"
		} else {
			event.Status = "OPEN"
		}

		allEvents = append(allEvents, &event)
	}

	return allEvents, nil
}
