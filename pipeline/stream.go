package pipeline

import (
	"reflect"
	"time"

	"github.com/influxdata/kapacitor/tick"
)

// A StreamNode represents the source of data being
// streamed to Kapacitor via any of its inputs.
// The `stream` variable in stream tasks is an instance of
// a StreamNode.
// StreamNode.From is the method/property of this node.
type StreamNode struct {
	node
}

func newStreamNode() *StreamNode {
	return &StreamNode{
		node: node{
			desc:     "stream",
			wants:    StreamEdge,
			provides: StreamEdge,
		},
	}
}

// Creates a new FromNode that can be further
// filtered using the Database, RetentionPolicy, Measurement and Where properties.
// From can be called multiple times to create multiple
// independent forks of the data stream.
//
// Example:
//    // Select the 'cpu' measurement from just the database 'mydb'
//    // and retention policy 'myrp'.
//    var cpu = stream
//        |from()
//            .database('mydb')
//            .retentionPolicy('myrp')
//            .measurement('cpu')
//    // Select the 'load' measurement from any database and retention policy.
//    var load = stream
//        |from()
//            .measurement('load')
//    // Join cpu and load streams and do further processing.
//    cpu
//        |join(load)
//            .as('cpu', 'load')
//        ...
//
func (s *StreamNode) From() *FromNode {
	f := newFromNode()
	s.linkChild(f)
	return f
}

// A FromNode selects a subset of the data flowing through a StreamNode.
// The stream node allows you to select which portion of the stream you want to process.
//
// Example:
//    stream
//        |from()
//           .database('mydb')
//           .retentionPolicy('myrp')
//           .measurement('mymeasurement')
//           .where(lambda: "host" =~ /logger\d+/)
//        |window()
//        ...
//
// The above example selects only data points from the database `mydb`
// and retention policy `myrp` and measurement `mymeasurement` where
// the tag `host` matches the regex `logger\d+`
type FromNode struct {
	chainnode

	// An expression to filter the data stream.
	// tick:ignore
	Expression tick.Node `tick:"Where"`

	// The dimensions by which to group to the data.
	// tick:ignore
	Dimensions []interface{} `tick:"GroupBy"`

	// The database name.
	// If empty any database will be used.
	Database string

	// The retention policy name
	// If empty any retention policy will be used.
	RetentionPolicy string

	// The measurement name
	// If empty any measurement will be used.
	Measurement string

	// Optional duration for truncating timestamps.
	// Helpful to ensure data points land on specific boundaries
	// Example:
	//    stream
	//       |from()
	//           .measurement('mydata')
	//           .truncate(1s)
	//
	// All incoming data will be truncated to 1 second resolution.
	Truncate time.Duration
}

func newFromNode() *FromNode {
	return &FromNode{
		chainnode: newBasicChainNode("from", StreamEdge, StreamEdge),
	}
}

//tick:ignore
func (n *FromNode) ChainMethods() map[string]reflect.Value {
	return map[string]reflect.Value{
		"GroupBy": reflect.ValueOf(n.chainnode.GroupBy),
		"Where":   reflect.ValueOf(n.chainnode.Where),
	}
}

// Creates a new stream node that can be further
// filtered using the Database, RetentionPolicy, Measurement and Where properties.
// From can be called multiple times to create multiple
// independent forks of the data stream.
//
// Example:
//    // Select the 'cpu' measurement from just the database 'mydb'
//    // and retention policy 'myrp'.
//    var cpu = stream
//        |from()
//            .database('mydb')
//            .retentionPolicy('myrp')
//            .measurement('cpu')
//    // Select the 'load' measurement from any database and retention policy.
//    var load = stream
//        |from()
//            .measurement('load')
//    // Join cpu and load streams and do further processing.
//    cpu
//        |join(load)
//            .as('cpu', 'load')
//        ...
//
func (s *FromNode) From() *FromNode {
	f := newFromNode()
	s.linkChild(f)
	return f
}

// Filter the current stream using the given expression.
// This expression is a Kapacitor expression. Kapacitor
// expressions are a superset of InfluxQL WHERE expressions.
// See the [expression](https://docs.influxdata.com/kapacitor/latest/tick/expr/) docs for more information.
//
// Multiple calls to the Where method will `AND` together each expression.
//
// Example:
//    stream
//       |from()
//          .where(lambda: condition1)
//          .where(lambda: condition2)
//
// The above is equivalent to this
// Example:
//    stream
//       |from()
//          .where(lambda: condition1 AND condition2)
//
//
// NOTE: Becareful to always use `|from` if you want multiple different streams.
//
// Example:
//  var data = stream
//      |from()
//          .measurement('cpu')
//  var total = data
//      .where(lambda: "cpu" == 'cpu-total')
//  var others = data
//      .where(lambda: "cpu" != 'cpu-total')
//
// The example above is equivalent to the example below,
// which is obviously not what was intended.
//
// Example:
//  var data = stream
//      |from()
//          .measurement('cpu')
//          .where(lambda: "cpu" == 'cpu-total' AND "cpu" != 'cpu-total')
//  var total = data
//  var others = total
//
// The example below will create two different streams each selecting
// a different subset of the original stream.
//
// Example:
//  var data = stream
//      |from()
//          .measurement('cpu')
//  var total = stream
//      |from()
//          .measurement('cpu')
//          .where(lambda: "cpu" == 'cpu-total')
//  var others = stream
//      |from()
//          .measurement('cpu')
//          .where(lambda: "cpu" != 'cpu-total')
//
//
// If empty then all data points are considered to match.
// tick:property
func (s *FromNode) Where(expression tick.Node) *FromNode {
	if s.Expression != nil {
		s.Expression = &tick.BinaryNode{
			Operator: tick.TokenAnd,
			Left:     s.Expression,
			Right:    expression,
		}
	} else {
		s.Expression = expression
	}
	return s
}

// Group the data by a set of tags.
//
// Can pass literal * to group by all dimensions.
// Example:
//  stream
//      |from()
//          .groupBy(*)
// tick:property
func (s *FromNode) GroupBy(tag ...interface{}) *FromNode {
	s.Dimensions = tag
	return s
}
