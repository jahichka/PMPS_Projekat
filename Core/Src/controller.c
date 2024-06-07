/*
 * controller.c
 *
 *  Created on: Jun 4, 2024
 *      Author: adeldsk
 */

#include "controller.h"
#include "nearest_neighbour.h"
#include "stdarg.h"
#include "stdio.h"
#include "stdlib.h"
#include "stm32f4xx_hal.h"

extern UART_HandleTypeDef *UART;
extern char msgbuf[1024];

struct Controller* CTRL_ControllerInit() {
	struct Controller *ctrl = malloc(sizeof(struct Controller));
	ctrl->load_count = 0;
	ctrl->wind_count = 0;
	ctrl->wind_speed = 0;
	ctrl->system_load = 0;
	ctrl->heatmap = 0;
	return ctrl;
}

struct Controller* CTRL_ControllerZeroInit() {
	struct Controller *ctrl = CTRL_ControllerInit();
	char temp[10] = "0101010100";
	CTRL_SetParameters(ctrl, temp);
	return ctrl;
}

void CTRL_Cleanup(struct Controller *ctrl) {
	for (int i = 0; i < ctrl->load_count; ++i) {
		free(ctrl->heatmap[i]);
	}
	if (ctrl->heatmap)
		free(ctrl->heatmap);
	if (ctrl->system_load)
		free(ctrl->system_load);
	if (ctrl->wind_speed)
		free(ctrl->wind_speed);
}

int CTRL_FindAngle(struct Controller *ctrl, int measured_wind, int sysload) {
	int nnw = nearest_neighbor(measured_wind, ctrl->wind_speed, ctrl->wind_count);
	int nna = nearest_neighbor(sysload, ctrl->system_load, ctrl->load_count);
	return ctrl->heatmap[nna][nnw];
}

void CTRL_SetParameters(struct Controller *ctrl, char *ethbuf) {
	CTRL_Cleanup(ctrl);

	char numbuf[2];
	memcpy(numbuf, &ethbuf[0], 2);
	ctrl->wind_count = atoi(numbuf);

	memcpy(numbuf, &ethbuf[2], 2);
	ctrl->load_count = atoi(numbuf);

	int offset = 4;
	ctrl->wind_speed = (uint8_t*) malloc(ctrl->wind_count * sizeof(uint8_t));
	for (int i = 0; i < ctrl->wind_count; ++i) {
		memcpy(numbuf, &ethbuf[offset], 2);
		ctrl->wind_speed[i] = atoi(numbuf);
		offset += 2;
	}

	ctrl->system_load = (uint8_t*) malloc(ctrl->load_count * sizeof(uint8_t));
	for (int i = 0; i < ctrl->load_count; ++i) {
		memcpy(numbuf, &ethbuf[offset], 2);
		ctrl->system_load[i] = atoi(numbuf);
		offset += 2;
	}

	ctrl->heatmap = malloc(ctrl->load_count * sizeof(uint8_t*));
	for (int i = 0; i < ctrl->load_count; ++i) {
		ctrl->heatmap[i] = malloc(ctrl->wind_count * sizeof(uint8_t));
		for (int j = 0; j < ctrl->wind_count; ++j) {
			memcpy(numbuf, &ethbuf[offset], 2);
			ctrl->heatmap[i][j] = atoi(numbuf);
			offset += 2;
		}
	}

	sprintf(msgbuf, "%s\r\n", CTRL_HeatmapString(ctrl));
	UART_Send();

	memset(ethbuf, 0, sizeof(ethbuf));
}

char* CTRL_HeatmapString(struct Controller *ctrl) {
	// Calculate the length of the string needed
	int buffer_size = 1000; // Initial buffer size (adjust as needed)
	char *buffer = (char*) malloc(buffer_size * sizeof(char));
	if (buffer == NULL) {
		fprintf(stderr, "Memory allocation failed\r\n");
		return NULL;
	}

	int position = 0; // Tracks the current position in the buffer

	// Print header row for wind speeds in green
	position += snprintf(buffer + position, buffer_size - position,
			"\033[0;32mHeatmap\t|\033[0;33m");
	for (int i = 0; i < ctrl->wind_count; ++i) {
		position += snprintf(buffer + position, buffer_size - position,
				"  %u\t|", ctrl->wind_speed[i]);
	}
	position += snprintf(buffer + position, buffer_size - position,
			"\r\n\033[0;32m---------\033[0;33m");
	// Print separator line
	for (int i = 0; i < ctrl->wind_count; ++i) {
		position += snprintf(buffer + position, buffer_size - position,
				"--------");
	}
	position += snprintf(buffer + position, buffer_size - position, "\r\n");

	// Print rows for blade angles and heatmap values
	for (int i = 0; i < ctrl->load_count; ++i) {
		// Print blade angle in yellow
		position += snprintf(buffer + position, buffer_size - position,
				"\033[0;33m  %u\t|\033[0;0m", ctrl->system_load[i]);

		// Print heatmap values for each wind speed
		for (int j = 0; j < ctrl->wind_count; ++j) {
			position += snprintf(buffer + position, buffer_size - position,
					"  %u\t|", ctrl->heatmap[i][j]);
		}
		position += snprintf(buffer + position, buffer_size - position, "\r\n");

		// Print separator line
		position += snprintf(buffer + position, buffer_size - position,
				"\033[0;33m---------\033[0;0m");
		for (int j = 0; j < ctrl->wind_count; ++j) {
			position += snprintf(buffer + position, buffer_size - position,
					"--------");
		}
		position += snprintf(buffer + position, buffer_size - position, "\r\n");
	}

	return buffer;
}
