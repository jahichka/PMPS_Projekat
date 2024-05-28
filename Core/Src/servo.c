/*
 * servo.c
 *
 *  Created on: Apr 26, 2024
 *      Author: Adel
 */

#include "servo.h"
#include "float.h"
#include "string.h"
#include "stdio.h"
#include "stm32f4xx_hal.h"

#define MIN_ANGLE 30 // 0 degrees
#define FULL_ANGLE 75 // 90 degrees
#define MAX_ANGLE 120 // 180 degrees

extern TIM_HandleTypeDef *PWM_Timer;
extern char msgbuf[1024];

void SERVO_On() {
	HAL_TIM_PWM_Start(PWM_Timer, TIM_CHANNEL_1);
}

void SERVO_Off() {
	HAL_TIM_PWM_Stop(PWM_Timer, TIM_CHANNEL_1);
}

void SERVO_Init(){
	HAL_TIM_Base_Start(PWM_Timer);
	HAL_TIM_PWM_Start(PWM_Timer, TIM_CHANNEL_1);
}

void SERVO_SetAngle(int angle) {
	if( angle < 0 || angle > 90 )
		return;

	float segment = ((float)(FULL_ANGLE - MIN_ANGLE) / 90);
	uint32_t set_angle = (MIN_ANGLE+(segment*angle));
	__HAL_TIM_SetCompare(PWM_Timer, TIM_CHANNEL_1, set_angle);
}

void SERVO_StartSweep(int count) {
	int angle = MIN_ANGLE, incr = 10;
	for (int i = 0; i < count; ++i) {
		do {
			__HAL_TIM_SetCompare(PWM_Timer, TIM_CHANNEL_1, angle);
			angle += incr;
			HAL_Delay(100);
		} while (angle <= MAX_ANGLE);
		do {
			__HAL_TIM_SetCompare(PWM_Timer, TIM_CHANNEL_1, angle);
			angle -= incr;
			HAL_Delay(100);
		} while (angle >= MIN_ANGLE);
	}
}
