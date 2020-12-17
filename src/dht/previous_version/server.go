/**
* @Author: huadong.hu@outlook.com
* @Date: 7/7/20 9:33 PM
* @Desc:
 */
package main

import (
	"log"
	"net"
	"net/rpc"
)

type SumServiceReq struct {
	FileName      string
	GoRoutineNums int
}

type SumServiceResp struct {
	Sum int
}

type SumService struct {
}

func (s *SumService) CalcSum(args SumServiceReq, resp *SumServiceResp) error {
	//TODO Add your code here
	resp.Sum = Sum(args.GoRoutineNums, args.FileName)
	return nil
}

func main() {
	sumService := new(SumService)
	err := rpc.Register(sumService)
	listener, err := net.Listen("tcp", ":8083")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}
		rpc.ServeConn(conn)
	}
}
