package link

import "io"

// receive 行为
// 1）执行后启动两个协程。协程1 进行消息的收取，解析和处理；协程2 消息响应数据的发送
// 2）协程1 会在收到消息后
//     a. 监听并接收消息
//     b. 判断是否handler是否可用
//        是->c
//        否->丢弃消息
//     c. 调用handler处理消息
//     d. 判断handler返回值是否为空
//        是->a
//        否->准备待响应消息
//     e. 向待发送消息chan写入响应信息
//
// receive implement Talk for receive async message
func (l *Linker) receive() {
	l.t.Go(func() error {
		for {
			in, err := l.stream.Recv()
			if err == io.EOF {
				close(l.msgAsync)
				return err
			}
			if err != nil {
				return err
			}
			l.log.Sugar().Debugf("link receive stream = %v\n", in)
			// check : is ack message
			if (in.Context.Flags & FlagAck) == FlagAck {
				l.acknowledge(in)
			} else {
				var resp []byte
				if l.handler != nil {
					resp = l.handler(in.Content)
				} else {
					l.log.Warn("handle not implemented, ack context is null")
				}
				msg := packetAckMsg(in, resp)
				l.msgAsync <- msg
			}
		}
	})
	l.t.Go(func() error {
		for {
			msg, ok := <-l.msgAsync
			if ok {
				err := l.stream.Send(msg)
				if err != nil {
					return err
				}
			}
		}
	})
}
