#include "stm32f4xx_hal.h"
#include "stdio.h"
#include "stdlib.h"
#include "socket.h"
#include "dhcp.h"
#include "w5500.h"
#include "string.h"

#define false 0
#define true 1

#define DHCP_SOCKET     0
#define DNS_SOCKET      1
#define HTTP_SOCKET     2
#define SOCK_TCPS       0
#define SOCK_UDPS       1
#define PORT_TCPS       5000
#define PORT_UDPS       3000
#define MAX_HTTPSOCK    6

extern SPI_HandleTypeDef hspi1;
extern UART_HandleTypeDef huart6;
extern char msgbuf[1024];
extern uint8_t recieve_buf[1024];

uint8_t socknumlist[] = { 2, 3, 4, 5, 6, 7 };
uint8_t RX_BUF[1024];
uint8_t TX_BUF[1024];
wiz_NetInfo net_info = { .mac = { 0x0A, 0xAD, 0xBE, 0xEF, 0xFE, 0xE2 }, .dhcp =
		NETINFO_DHCP };

void wizchipSelect(void) {
	HAL_GPIO_WritePin(GPIOA, GPIO_PIN_4, GPIO_PIN_RESET);
}

void wizchipUnselect(void) {
	HAL_GPIO_WritePin(GPIOA, GPIO_PIN_4, GPIO_PIN_SET);
}

void wizchipReadBurst(uint8_t *buff, uint16_t len) {
	HAL_SPI_Receive(&hspi1, buff, len, HAL_MAX_DELAY);
}

void wizchipWriteBurst(uint8_t *buff, uint16_t len) {
	HAL_SPI_Transmit(&hspi1, buff, len, HAL_MAX_DELAY);
}

uint8_t wizchipReadByte(void) {
	uint8_t byte;
	wizchipReadBurst(&byte, sizeof(byte));
	return byte;
}

void wizchipWriteByte(uint8_t byte) {
	wizchipWriteBurst(&byte, sizeof(byte));
}

volatile uint8_t ip_assigned = false;

void Callback_IPAssigned(void) {
	ip_assigned = true;
}

void Callback_IPConflict(void) {
	ip_assigned = false;
}

uint8_t dhcp_buffer[1024];
uint8_t dns_buffer[1024];

uint8_t ETH_Init() {
	// Register W5500 callbacks
	reg_wizchip_cs_cbfunc(wizchipSelect, wizchipUnselect);
	reg_wizchip_spi_cbfunc(wizchipReadByte, wizchipWriteByte);
	reg_wizchip_spiburst_cbfunc(wizchipReadBurst, wizchipWriteBurst);

	uint8_t rx_tx_buff_sizes[] = { 2, 2, 2, 2, 2, 2, 2, 2 };
	wizchip_init(rx_tx_buff_sizes, rx_tx_buff_sizes);

	// set MAC address before using DHCP
	setSHAR(net_info.mac);
	DHCP_init(DHCP_SOCKET, dhcp_buffer);

	reg_dhcp_cbfunc(Callback_IPAssigned, Callback_IPAssigned,
			Callback_IPConflict);

	sprintf(msgbuf, "Obtaining network configuration from DHCP ... \r\n");
	HAL_UART_Transmit(&huart6, (uint8_t*)msgbuf, strlen(msgbuf), 100);
	uint32_t ctr = 100;
	while ((!ip_assigned) && ctr) {
		DHCP_run();
		--ctr;
		HAL_Delay(50);
	}
	if (!ip_assigned) {
		sprintf(msgbuf, "Failed to obtain IP, returning ... \r\n");
		HAL_UART_Transmit(&huart6, (uint8_t*) msgbuf, strlen(msgbuf), 100);
		return 0;
	}
	sprintf(msgbuf, "Network configuration obtained!\r\n-----------------------\r\n");
	HAL_UART_Transmit(&huart6, (uint8_t*)msgbuf, strlen(msgbuf), 100);

	getIPfromDHCP(net_info.ip);
	getGWfromDHCP(net_info.gw);
	getSNfromDHCP(net_info.sn);

//    char charData[200]; // Data holder
//    sprintf(charData,"IP:  %d.%d.%d.%d\r\nGW:  %d.%d.%d.%d\r\nNet: %d.%d.%d.%d\r\n",
//        net_info.ip[0], net_info.ip[1], net_info.ip[2], net_info.ip[3],
//        net_info.gw[0], net_info.gw[1], net_info.gw[2], net_info.gw[3],
//        net_info.sn[0], net_info.sn[1], net_info.sn[2], net_info.sn[3]
//    );
//    HAL_UART_Transmit(&huart1,(uint8_t *)charData,strlen(charData),1000);

	wizchip_setnetinfo(&net_info);
	wizchip_getnetinfo(&net_info);

	sprintf(msgbuf,
			"IP:\t%d.%d.%d.%d\r\nSN:\t%d.%d.%d.%d\r\nGW:\t%d.%d.%d.%d\r\n-----------------------\r\n",
			net_info.ip[0], net_info.ip[1], net_info.ip[2], net_info.ip[3],
			net_info.sn[0], net_info.sn[1], net_info.sn[2], net_info.sn[3],
			net_info.gw[0], net_info.gw[1], net_info.gw[2], net_info.gw[3]);
	HAL_UART_Transmit(&huart6, (uint8_t*) msgbuf, strlen(msgbuf), 100);

	return 1;
}

uint8_t ETH_SocketInit(uint8_t *sck) {
	int8_t ret;
	if ((ret = socket(*sck, Sn_MR_TCP, 5000, SF_TCP_NODELAY)) != 0) {
		return -1;
	}
	sprintf(msgbuf, "Socket initialized!\r\n");
	HAL_UART_Transmit(&huart6, (uint8_t*) msgbuf, strlen(msgbuf), 100);
	return 0;
}

uint8_t ETH_Connect(uint8_t *sck, char *server) {
	uint8_t server_ip[4];
	uint16_t server_port;

	char *ip_part;
	char *port_part;

	// Split the input string using ":" as delimiter
	ip_part = strtok(server, ":");
	port_part = strtok(NULL, ":");

	if (ip_part != NULL && port_part != NULL) {
		printf("IP address: %s\n", ip_part);
		printf("Port: %s\n", port_part);

		// Convert IP address to uint8_t array
		sscanf(ip_part, "%d.%d.%d.%d", &server_ip[0], &server_ip[1],
				&server_ip[2], &server_ip[3]);

		// Convert port to uint16_t
		server_port = atoi(port_part);
	} else {
		return -2;
	}

	sprintf(msgbuf, "Attempting to connect to %d.%d.%d.%d:%d\r\n", server_ip[0], server_ip[1],
				server_ip[2], server_ip[3], server_port);
	UART_Send();

	int8_t ret;
	if ((ret = connect(*sck, server_ip, server_port)) != SOCK_OK) {
		return -1;
	}

	sprintf(msgbuf, "Server connected!\r\n");
	HAL_UART_Transmit(&huart6, (uint8_t*) msgbuf, strlen(msgbuf), 100);
	return 0;
}

int8_t ETH_Listen(uint8_t *sck, char* buf){
	int RSR_Len = 0, repeat = 500;
	while(!RSR_Len && repeat){
		RSR_Len = getSn_RX_RSR(*sck);
	}
	if(!repeat){
		return 0;
	} else {
		return recv(*sck, buf, RSR_Len);
	}
}

void ETH_Send(uint8_t* sck, char* msg){
	send(*sck, (uint8_t*)msg, strlen(msg));
}



