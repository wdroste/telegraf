//go:generate ../../../tools/readme_config_includer/generator
package neo4j

import (
	_ "embed"
	"fmt"

	nj "github.com/neo4j/neo4j-go-driver/v4/neo4j"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Neo4j struct {
	Driver   nj.Driver
	Uri      string          `toml:"uri"`
	Username string          `toml:"username"`
	Password string          `toml:"password"`
	Log      telegraf.Logger `toml:"-"`
}

func (*Neo4j) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (s *Neo4j) Init() error {
	return nil
}

func (n *Neo4j) Gather(acc telegraf.Accumulator) error {
	if n.Driver == nil {
		token := nj.NoAuth()
		if len(n.Username) != 0 {
			token = nj.BasicAuth(n.Username, n.Password, "")
		}

		driver, err := nj.NewDriver(n.Uri, token)
		if err != nil {
			return err
		}
		n.Driver = driver
	}

	session := n.Driver.NewSession(nj.SessionConfig{AccessMode: nj.AccessModeRead})
	defer session.Close()

	query := "SHOW TRANSACTIONS YIELD" +
		" database, " +
		" transactionId, " +
		" currentQueryId, " +
		" status, " +
		" activeLockCount, " +
		" pageHits, " +
		" elapsedTime, " +
		" cpuTime, " +
		" waitTime, " +
		" idleTime " +
		"WHERE elapsedTime.milliseconds > 1000 " +
		"RETURN " +
		" database, " +
		" transactionId, " +
		" currentQueryId, " +
		" status, " +
		" activeLockCount, " +
		" pageHits, " +
		" elapsedTime.milliseconds AS elapsedTimeMillis, " +
		" cpuTime.milliseconds AS cpuTimeMillis, " +
		" waitTime.milliseconds AS waitTimeMillis, " +
		" idleTime.seconds AS idleTimeSeconds"

	session.ReadTransaction(func(tx nj.Transaction) (interface{}, error) {
		result, err := tx.Run(query, map[string]interface{}{})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			record := result.Record()

			db, _ := record.Get("database")
			database := fmt.Sprintf("%v", db)
			//transactionId, _ := record.Get("transactionId")
			currentQueryId, _ := record.Get("currentQueryId")
			activeLockCount, _ := record.Get("activeLockCount")
			pageHits, _ := record.Get("pageHits")
			elapsedTimeMillis, _ := record.Get("elapsedTimeMillis")
			cpuTimeMillis, _ := record.Get("cpuTimeMillis")
			waitTimeMillis, _ := record.Get("waitTimeMillis")
			idleTimeSeconds, _ := record.Get("idleTimeSeconds")

			tags := map[string]string{"database": database}
			// extract all the data
			fields := map[string]interface{}{
				//"transactionId":     transactionId,
				"currentQueryId":    currentQueryId,
				"activeLockCount":   activeLockCount,
				"pageHits":          pageHits,
				"elapsedTimeMillis": elapsedTimeMillis,
				"cpuTimeMillis":     cpuTimeMillis,
				"waitTimeMillis":    waitTimeMillis,
				"idleTimeSeconds":   idleTimeSeconds,
			}

			acc.AddFields("neo4j", fields, tags)
		}
		return nil, nil
	})
	return nil
}

func init() {
	inputs.Add("neo4j", func() telegraf.Input { return &Neo4j{} })
}
