//go:build !custom || inputs || inputs.neo4j

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/neo4j" // register plugin
