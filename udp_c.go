package udp

/*
#cgo CFLAGS: -Iinclude
#cgo LDFLAGS: -Llib/ -llibwsock32
#include <stdio.h>
#include <stdlib.h>
#include <winsock2.h>
#pragma comment(lib, "ws2_32")

//static SOCKET Udp;
#define UR_CLIENT_SOCK_BUF_SIZE (65536)
#define UR_SERVER_SOCK_BUF_SIZE (UR_CLIENT_SOCK_BUF_SIZE * 32)

int udp_init_pion(char *ip, int port){
    struct sockaddr_in ser; //服务器端地址
	int Udp;
    WSADATA wsaData;
    if (WSAStartup(MAKEWORD(2, 2), &wsaData) != 0)
    {
        printf("Failed to load Winsock.\n"); //Winsock 初始化错误
        return -1;
    }
    ser.sin_family = AF_INET;                       //初始化服务器地址信息
    ser.sin_port = htons(port);                     //端口转换为网络字节序
    ser.sin_addr.s_addr = inet_addr(ip);            //IP 地址转换为网络字节序
	if (sizeof(ip) > 32) {
		ser.sin_family = AF_INET6;
		Udp = socket(AF_INET6, SOCK_DGRAM, IPPROTO_UDP); //创建UDP套接字
	} else {
		Udp = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP); //创建UDP套接字
	}

    if (Udp == INVALID_SOCKET)
    {
        printf("socket() Failed: %d\n", WSAGetLastError());
        return -1;
    }
	int nRecvBuf=3*1024*1024;//设置为32K
	setsockopt(Udp,SOL_SOCKET,SO_RCVBUF,(const char*)&nRecvBuf,sizeof(int));


	////发送缓冲区
	int nSendBuf=3*1024*1024;//设置为32K
	setsockopt(Udp,SOL_SOCKET,SO_SNDBUF,(const char*)&nSendBuf,sizeof(int));
    if (bind(Udp, (LPSOCKADDR)&ser, sizeof(ser)) == SOCKET_ERROR)
    {
        printf("绑定IP和端口\n");
        return 0;
    }
    printf("udp init ok\n");
	return Udp;
}

int udp_send_pion(int Udp, char *ip, int port, unsigned char *buff, int len){
    struct sockaddr_in s;
    int ret,addrlen;

	s.sin_family = AF_INET;
	if (sizeof(ip) > 32) {
		s.sin_family = AF_INET6;
	}
    s.sin_port = htons(port);
    s.sin_addr.s_addr = inet_addr(ip);
    addrlen = sizeof(s);
    sendto(Udp, buff, len, 0, (SOCKADDR *)&s, addrlen);
    return 0;
}

int udp_rcv_pion(int Udp, char *ip, int *port, unsigned char *buff, int *len){
	struct sockaddr_in s;
	int ret,addrlen;

	addrlen = sizeof(s);
	ret = recvfrom(Udp, buff, 1600, 0, (SOCKADDR *)&s, &addrlen);
	if (ret > 0)
	{
		memcpy(ip, inet_ntoa(s.sin_addr), strlen(inet_ntoa(s.sin_addr)));//copy ip to from_ip
		*port = ntohs(s.sin_port);
		ip = inet_ntoa(s.sin_addr);
	}
	*len = ret;

	return ret;
}

int close_udp_pion(int Udp){
    closesocket(Udp); //关闭 socket
    WSACleanup();
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

type UdpByC struct {
	Conn C.int
	Ip   string
	Port int
}

func NewUdp(ip string, port int) (*UdpByC, error) {
	udp := &UdpByC{
		Ip:   ip,
		Port: port,
	}

	udp_conn := C.udp_init_pion(C.CString(ip), C.int(port))
	udp.Conn = udp_conn
	if int(udp_conn) <= 0 {
		return nil, errors.New("create udp connect is fail")
	}
	return udp, nil
}

func (u *UdpByC) SendMsg(msg []byte, msglen int) {
	go func() {
		c_char := (*C.uchar)(unsafe.Pointer(&msg[0]))
		C.udp_send_pion(u.Conn, C.CString(u.Ip), C.int(u.Port), c_char, C.int(msglen))
		//if int(ret) <= 0 {
		//	fmt.Println("")
		//}
		//return errors.New("发送消息异常了")
	}()
}

func (u *UdpByC) RecvMsg(c_char *C.uchar) (int, string, int, error) {
	//c_char := (*C.uchar)(unsafe.Pointer(&msg[0]))
	var n int
	c_len := (*C.int)(unsafe.Pointer(&n))

	var ip string
	c_ip := (*C.char)(unsafe.Pointer(&ip))

	var port int
	c_port := (*C.int)(unsafe.Pointer(&port))

	ret := C.udp_rcv_pion(u.Conn, c_ip, c_port, c_char, c_len)
	if int(ret) > 0 {
		return 0, C.GoString(c_ip), int(*c_port), nil
	}
	return 0, "", 0, errors.New("读取消息异常了")
}

func (u *UdpByC) Close() error {
	ret := C.close_udp_pion(u.Conn)
	if int(ret) > 0 {
		return nil
	}
	return errors.New("关闭异常了")
}
