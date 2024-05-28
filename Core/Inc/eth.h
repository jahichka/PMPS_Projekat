/*
 * w5500_client.h
 *
 *  Created on: May 14, 2024
 *      Author: Adel
 */

#ifndef INC_ETH_CLIENT_H_
#define INC_ETH_CLIENT_H_

const char* const ETH_EV_AUTH = "auth";
const char* const ETH_EV_PING = "ping";

uint8_t ETH_Init();
uint8_t ETH_SocketInit(uint8_t*);
uint8_t ETH_Connect(uint8_t*, char*);
int8_t ETH_Listen(uint8_t*, char*);
void ETH_Send(uint8_t*, char*);
void ETH_IRQ_Handler(uint8_t*);

#endif /* INC_ETH_CLIENT_H_ */
