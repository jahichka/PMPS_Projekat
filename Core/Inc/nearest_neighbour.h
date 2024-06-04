/*
 * nearest_neighbour.h
 *
 *  Created on: Jun 4, 2024
 *      Author: adeldsk
 */

#ifndef INC_NEAREST_NEIGHBOUR_H_
#define INC_NEAREST_NEIGHBOUR_H_

int abs(int);
int nearest_neighbor(int val, int *arr, int size);
int find_min_distance(int *arr, int val, int size, int index, int left, int right);

#endif /* INC_NEAREST_NEIGHBOUR_H_ */
