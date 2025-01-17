package influxdb

import (
	"fmt"
	"strconv"
	"time"

	bulkQuerygen "github.com/influxdata/influxdb-comparisons/bulk_query_gen"
)

type InfluxGroupAggregateQuery struct {
	InfluxCommon
	aggregate Aggregate
}

func NewInfluxGroupAggregateQuery(agg Aggregate, lang Language, dbConfig bulkQuerygen.DatabaseConfig, queriesFullRange bulkQuerygen.TimeInterval, scaleVar int) bulkQuerygen.QueryGenerator {
	if _, ok := dbConfig[bulkQuerygen.DatabaseName]; !ok {
		panic("need influx database name")
	}

	version, err := strconv.Atoi(dbConfig["influxVersion"])
	if err != nil {
		panic("invalid influx version")
	}

	return &InfluxGroupAggregateQuery{
		InfluxCommon: *newInfluxCommon(lang, dbConfig[bulkQuerygen.DatabaseName], queriesFullRange, scaleVar, version),
		aggregate:    agg,
	}
}

func (d *InfluxGroupAggregateQuery) Dispatch(i int) bulkQuerygen.Query {
	q := bulkQuerygen.NewHTTPQuery()
	d.GroupAggregateQuery(q)
	return q
}

func (d *InfluxGroupAggregateQuery) GroupAggregateQuery(qi bulkQuerygen.Query) {
	interval := d.AllInterval.RandWindow(time.Hour * 6)

	var query string
	if d.language == InfluxQL {
		query = fmt.Sprintf("SELECT %s(temperature) FROM air_condition_room WHERE time > '%s' AND time < '%s' GROUP BY home_id",
			d.aggregate, interval.StartString(), interval.EndString())
	} else {
		query = fmt.Sprintf(`from(bucket:"%s")
            |> range(start:%s, stop:%s)
            |> filter(fn:(r) => r._measurement == "air_condition_room" and r._field == "temperature")
            |> group(columns:["home_id"])
            |> %s()
            |> yield()`,
			d.DatabaseName,
			interval.StartString(), interval.EndString(),
			d.aggregate)
	}

	humanLabel := fmt.Sprintf("InfluxDB (%s) %s temperature, rand %s by home_id", d.language.String(), d.aggregate, interval.StartString())
	q := qi.(*bulkQuerygen.HTTPQuery)
	d.getHttpQuery(humanLabel, interval.StartString(), query, q)
}
