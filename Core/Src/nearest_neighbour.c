/*
 * nearest_neighbour.c
 *
 *  Created on: Jun 4, 2024
 *      Author: adeldsk
 */
#include "stdint.h"

int abs(int num) {
  if (num < 0)
    return -num;
  else
    return num;
}

int find_min_distance(uint8_t* arr, int val, int size, int index, int left, int right) {
  int min = abs(arr[index] - val);
  if (right < size - 1)
    ++right;
  if (left > 0)
    --left;
  if (abs(arr[left] - val) < min) {
    index = left;
    min = abs(arr[left] - val);
  }
  if (abs(arr[right] - val) < min)
    index = right;
  return index;
}

int nearest_neighbor(int val, uint8_t* arr, int size) {
  int left = 0, right = size - 1;
  int index = left + ( right - left ) / 2;

  while (arr[index] != val && left <= right) {
	  if ( arr[index] < val ){
		  left = index + 1;
	  } else {
		  right = index - 1;
	  }
	  index = left + ( right - left ) / 2;
  }
  if(index >= size) {
	  index = size - 1;
  }
  if (arr[index] == val)
    return index;
  else
    return find_min_distance(arr, val, size, index, left, right);
}
