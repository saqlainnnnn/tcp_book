package main

import (
	"context"
	"io"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestListenerer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func()  { _ = listener.Close() } ()

	t.Logf("bound to %q", listener.Addr())
}

func TestDial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
/*1*/go func() {
		defer func ()  { done <- struct{}{} }()

		for {
			conn, err := listener.Accept() /*2*/
			if err != nil {
				t.Log(err)
				return
			}

	/*3*/	go func(c net.Conn) {
				defer func ()  {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)/*4*/
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("recieved : %q", buf[:n])
				}
			}(conn)
		}
	}()
	
/*5*/conn, err := net.Dial("tcp", /*7*/listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
/*8*/	conn.Close()
	<-done
/*9*/	listener.Close()
	<-done
}

func /*1*/ DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer {
/*2*/	Control: func(network, address string, _ syscall.RawConn) error {
			return &net.DNSError{
				Err: 			"connection timed out",
				Name: 			address,
				Server: 		"127.0.0.1",
				IsTimeout: 		true,
				IsTemporary: 	true,
			}
		},
		Timeout: timeout,	
	}
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T)  {
	c, err := DialTimeout("tcp", "10.0.0.1:http",/*3*/ 5*time.Second)
	if err == nil {
		c.Close()
		t.Fatal("connection did not time out")
	}

	nErr, ok := /*4*/ err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	if /*5*/ !nErr.Timeout(){
		t.Fatal("error is not a timeout")
	}
}

func TestDialContext(t *testing.T)  {
	dl := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()

	var d net.Dialer 
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		time.Sleep(5 * time.Second + time.Millisecond)
		return nil
	}
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
	if err != nil {
		conn.Close()
		t.Fatal("connection didnt time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout :%v", err)
		}
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}