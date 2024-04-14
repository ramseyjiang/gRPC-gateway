package db

import (
	"time"

	"syreclabs.com/go/faker"
)

func (s *sportsRepo) seed() error {
	statement, err := s.db.Prepare(`CREATE TABLE IF NOT EXISTS sports (id INTEGER PRIMARY KEY, name TEXT, result TEXT, location TEXT, visible INTEGER, start_time DATETIME, end_time DATETIME, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	for i := 1; i <= 100; i++ {
		statement, err = s.db.Prepare(`INSERT OR IGNORE INTO sports(id, name, result, location, visible, start_time, end_time, advertised_start_time) VALUES (?,?,?,?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				faker.Team().Name(),
				faker.Team().State(),
				faker.Address().City(),
				faker.Number().Between(0, 1),
				// make sure the start time gather than the end time
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 0)).Format(time.RFC3339),
				faker.Time().Between(time.Now().AddDate(0, 0, 1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
			)
		}
	}

	return err
}
