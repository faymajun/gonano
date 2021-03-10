package snowflake

import (
	"encoding/json"
	"errors"
	"github.com/faymajun/gonano/corex/misc"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"

	"github.com/golang/protobuf/proto"
	"gopkg.in/mgo.v2/bson"
)

const (
	DEFAULT_EPOCH     = 1577808000000
	DEFAULT_NODE_BITS = 12
	DEFAULT_STEP_BITS = 13
)

func bson2pb(b bson.M, pb proto.Message) error {
	ctx, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return json.Unmarshal(ctx, pb)
}

type Node struct {
	id    int64 // nodeId
	epoch int64
	last  int64
	step  int64

	stepMask  int64
	timeShift uint8
	nodeShift uint8

	ids chan int64
}

func NewNode(id int64, epoch int64, nodeBits uint8, stepBits uint8) (*Node, error) {
	var nodeIdMax int64

	nodeIdMax = -1 ^ (-1 << nodeBits)
	if id < 0 || id > nodeIdMax {
		return nil, errors.New("invalid node id")
	}

	n := &Node{
		id:        id,
		epoch:     epoch,
		stepMask:  -1 ^ (-1 << stepBits),
		timeShift: nodeBits + stepBits,
		nodeShift: stepBits,
		ids:       make(chan int64, 1<<stepBits),
	}
	return n, nil
}

func (n *Node) Start() {
	go func() {
		for {
			now := misc.NowMS()
			if now == n.last {
				n.step = (n.step + 1) & n.stepMask
				if n.step == 0 {
					for now <= n.last {
						now = misc.NowMS()
					}
				}
			} else {
				n.step = 0
			}
			n.last = now
			r := ((now - n.epoch) << n.timeShift) | (n.id << n.nodeShift) | n.step
			n.ids <- r
		}
	}()
}

func (n *Node) GenID() int64 {
	return <-n.ids
}

var (
	_node *Node
	once  sync.Once
)

func Init(nodeId int32) {
	once.Do(func() {
		var err error
		_node, err = NewNode(int64(nodeId), DEFAULT_EPOCH, DEFAULT_NODE_BITS, DEFAULT_STEP_BITS)
		if err != nil {
			log.Fatal(err)
		}
		_node.Start()
	})
}

func GenID() string {
	if _node == nil {
		log.Fatal("snowflake don't been initialized!!!")
	}

	return strconv.FormatInt(_node.GenID(), 16)
}
