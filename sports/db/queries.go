package db

const (
	eventsList = "list"
)

func getEventQueries() map[string]string {
	return map[string]string{
		eventsList: `
			SELECT 
				id,
				name,
				result,
				location,
				visible, 
				start_time, 
				end_time,
				advertised_start_time
			FROM sports
		`,
	}
}
