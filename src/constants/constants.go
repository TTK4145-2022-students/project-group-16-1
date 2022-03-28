package constants

import "time"

const N_FLOORS = 4
const N_BTN_TYPES = 3
const HARDWARE_ADDR = "localhost:15657"
const PERIODIC_SEND_DURATION = 50 * time.Millisecond
const TRAVEL_DURATION = 2500 * time.Millisecond
const DOOR_OPEN_DURATION = 3000 * time.Millisecond
const MOTOR_FAILURE_DURATION = TRAVEL_DURATION + DOOR_OPEN_DURATION + 500*time.Millisecond
