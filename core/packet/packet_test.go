package packet

import (
	"fmt"
	"testing"
)

type Sender interface {
	Send()
}
type TestSender struct {
	a int
}

func (*TestSender) Send() {

}

var str = "aaaaaa"
var str1 = "aaaaasaa"
var str2 = "aaaewraaa"
var str3 = "aaavvcaaa"

func test(send Sender) {
	if len(str) == 100 {
		fmt.Println(str)
	}
}
func BenchmarkCPacket_interface(b *testing.B) {
	sender := &TestSender{}
	//str1 := "缓速移动达到最大速度,运动时间=%f,运动后的位置：%v"
	for i := 0; i < b.N; i++ {
		test(sender)
	}
}

func test1(send interface{}) {
	if len(str) == 100 {
		fmt.Println(str)
	}
}

func BenchmarkCPacket_interface1(b *testing.B) {
	sender := &TestSender{}
	//str1 := "缓速移动达到最大速度,运动时间=%f,运动后的位置：%v"
	for i := 0; i < b.N; i++ {
		test1(sender)
	}
}

func test2(send *TestSender) {
	if len(str) == 100 {
		fmt.Println(str)
	}
}

func BenchmarkCPacket_interface2(b *testing.B) {
	sender := &TestSender{}
	//str1 := "缓速移动达到最大速度,运动时间=%f,运动后的位置：%v"
	for i := 0; i < b.N; i++ {
		test2(sender)
	}
}

func test3(send string, send1 string, send2 string) {
	send = ""
	send2 = ""
	send1 = ""
}

func BenchmarkCPacket_value(b *testing.B) {
	for i := 0; i < b.N; i++ {
		test3(str, str1, str2)
	}
}

func test4(send string, args ...interface{}) {

}

func BenchmarkCPacket_valueAndInt(b *testing.B) {

	//sender1 := &TestSender{}
	//sender2 := &TestSender{}
	//sender3 := &TestSender{}
	for i := 0; i < b.N; i++ {
		test4(str)
	}
}

func test5(send string, send1 *TestSender, send2 *TestSender, send3 *TestSender) {
	send1.Send()
}

func BenchmarkCPacket_valueMul(b *testing.B) {
	sender := &TestSender{}
	sender1 := &TestSender{}
	sender2 := &TestSender{}
	for i := 0; i < b.N; i++ {
		test5(str, sender, sender1, sender2)
	}
}

func test6(a int) {
	a--
}
func BenchmarkCPacket_funccall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		test6(1)
	}
}

func BenchmarkCPacket_valueAssign(b *testing.B) {
	for i := 0; i < b.N; i++ {
		i := 1
		b := i
		b--
	}
}

func testpointerSize(a []*TestSender, index int) error {
	if index != len(a) {
		err := testpointerSize(a, index+1)
		index--
		return err
	} else {
		return nil
	}
}

func Benchmark_pointerSize(b *testing.B) {
	var a *TestSender = &TestSender{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := make([]*TestSender, 1)
		c[0] = a
		if len(c) == 1 {
			fmt.Println(c[0])
		}
	}
}
