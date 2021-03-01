package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	var (
		wg sync.WaitGroup
		in chan interface{}
	)
	wg.Add(len(jobs))
	for _, v := range jobs {
		out := make(chan interface{})
		go func(v job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			v(in, out)
		}(v, in, out)
		in = out
	}
	wg.Wait()
}

func smth(buf string, out chan interface{}) {
	out <- buf
}

func SingleHash(in, out chan interface{}) {
	var (
		mainWG   sync.WaitGroup
		nestedWG sync.WaitGroup
		crc32    string
		crc32Md5 string
	)
	for data := range in {
		mainWG.Add(1)
		stringData := strconv.Itoa(data.(int))
		md5Data := DataSignerMd5(stringData)
		go func(stringData string, md5Data string, out chan interface{}) {
			defer mainWG.Done()
			nestedWG.Add(2)
			go func(str string, res *string) {
				defer nestedWG.Done()
				*res = DataSignerCrc32(str)
			}(stringData, &crc32)
			go func(str string, res *string) {
				defer nestedWG.Done()
				*res = DataSignerCrc32(str)
			}(md5Data, &crc32Md5)
			nestedWG.Wait()
			out <- crc32 + "~" + crc32Md5
		}(stringData, md5Data, out)
	}
	mainWG.Wait()
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for data := range in {
		wg.Add(1)
		var buf [6]string
		go func(data interface{}, out chan interface{}) {
			defer wg.Done()
			var wg1 sync.WaitGroup
			for i := 0; i < 6; i++ {
				wg1.Add(1)
				go func(i int, data interface{}) {
					defer wg1.Done()
					th := strconv.Itoa(i)
					buf[i] = DataSignerCrc32(th + data.(string))
				}(i, data)
			}
			wg1.Wait()
			var res string
			for _, v := range buf {
				res += v
			}
			out <- res
		}(data, out)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var res []string
	for data := range in {
		res = append(res, data.(string))
	}
	sort.Strings(res)
	var newres string
	for _, sortData := range res {
		newres += sortData + "_"
	}
	newres = strings.TrimRight(newres, "_")
	out <- newres
}

func main() {

}
