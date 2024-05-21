#ifndef SERVO
#define SERVO

#include "main.h"

void SERVO_Init();
void SERVO_SetAngle(int angle);
void SERVO_StartSweep(int count);
void SERVO_On();
void SERVO_Off();

#endif
