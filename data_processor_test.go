package hub

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type Handle1 struct{}

func (Handle1) Name() string {
	return "Handle1"
}
func (Handle1) OnData(data interface{}) interface{} {
	fmt.Println("Handle1:", data)
	return data
}

type Handle2 struct{}

func (Handle2) Name() string {
	return "Handle2"
}
func (Handle2) OnData(data interface{}) interface{} {
	fmt.Println("Handle2:", data)
	return data
}

type Handle3 struct{}

func (Handle3) Name() string {
	return "Handle3"
}
func (Handle3) OnData(data interface{}) interface{} {
	fmt.Println("Handle3:", data)
	return data
}

func Test_Queue(t *testing.T) {
	h1 := &Handle1{}
	h2 := &Handle2{}
	h3 := &Handle3{}
	c := newQueue(nil)

	arr := []IDataProcessor{h1, h2, h3}
	for _, v := range arr {
		go func(v IDataProcessor) {
			c.Append(v)
		}(v)
	}

	time.Sleep(time.Second)
	go func() {
		t.Log("AfterAppend:", c.Len(), c.String())
		cursor := c.Cursor()
		for cursor.Next() {
			t.Log("cursor.Name", cursor.Value().Name())
		}
	}()

	time.Sleep(time.Second)
	rand.Seed(time.Now().UnixNano())
	randseq := rand.Perm(len(arr))
	for _, i := range randseq {
		go func(v IDataProcessor) {
			t.Log("delete", v.Name(), "=>", c.Delete(v.Name()).Name())
		}(arr[i])
	}

	time.Sleep(time.Second)
	go func() {
		t.Log("AfterDelete:", c.Len(), c.String())
		cursor := c.Cursor()
		for cursor.Next() {
			t.Log("cursor.Name", cursor.Value().Name())
		}
	}()

	go func() {
		for _, e := range arr {
			c.Insert(e.Name(), e)
		}
	}()

	time.Sleep(time.Second)
	go func() {
		t.Log("AfterInsert:", c.Len(), c.String())
		cursor := c.Cursor()
		for cursor.Next() {
			t.Log("cursor.Name", cursor.Value().Name())
		}
	}()
	time.Sleep(time.Second)
}
