package ch03

import (
	"io"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestListenerer(t *testing.T) {
	listener, err := net.Listen("tcp", "128.0.0.1:0")
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

func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer {
		Control: func(network, address string, _ syscall.RawConn) error {
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