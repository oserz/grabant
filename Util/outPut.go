package Util

import (
	"encoding/csv"
	//	"fmt"
	"os"

	"github.com/bitly/go-simplejson"
)

type OutputCSV struct {
	chanOutput chan interface{}
	filePath   string
	csvHead    []string
	w          *csv.Writer
}

func NewOutputCSV(chanNum int, filename string) *OutputCSV {
	outPath, _ := os.Getwd()
	outPath = outPath + string(os.PathSeparator) + filename
	return &OutputCSV{chanOutput: make(chan interface{}, chanNum), filePath: outPath}
}

func (this *OutputCSV) InitOutput(head []string) {
	f, err := os.Create(this.filePath)
	if err != nil {
		return
	}
	f.WriteString("\xEF\xBB\xBF") // Write UTF-8 BOM
	w := csv.NewWriter(f)
	w.Write(head)
	w.Flush()
	this.w = w
	this.csvHead = head

	go func() {
		for {
			var wbuff []string
			item := <-this.chanOutput
			if v, ok := item.(*simplejson.Json); ok {
				for _, heads := range this.csvHead {
					tmpstr, _ := v.Get(heads).String()
					wbuff = append(wbuff, tmpstr)
				}
			}
			this.w.Write(wbuff)
			this.w.Flush()
		}
	}()

}

func (this *OutputCSV) Out(js *simplejson.Json) {
	//	w := csv.NewWriter(this.f)
	/*
		go func() {
			this.chanOutput <- js
		}()
	*/
	this.chanOutput <- js
}

func (this *OutputCSV) Wait() {

}
