/*
 * w5500_client.h
 *
 *  Created on: May 14, 2024
 *      Author: Adel
 */

#ifndef INC_ETH_CLIENT_H_
#define INC_ETH_CLIENT_H_

uint8_t ETH_Init();
uint8_t ETH_SocketInit(uint8_t*);
uint8_t ETH_Connect(uint8_t*, char*);
int8_t ETH_Listen(uint8_t*, char*);
int32_t ETH_Recieve(uint8_t*, char*);
void ETH_ChipReset();
void ETH_Send(uint8_t*, char*);
void ETH_IRQ_Handler(uint8_t*);
void ETH_MessageHandler(uint8_t*, char*, int32_t);

#endif /* INC_ETH_CLIENT_H_ */
